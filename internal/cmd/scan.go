package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"text/tabwriter"

	"github.com/jonryanedge/prego/internal/config"
	"github.com/jonryanedge/prego/internal/fs"
	"github.com/spf13/cobra"
)

var (
	scanDepth    int
	scanCategory string
	scanWrite    bool
	scanLocal    bool
)

var scanCmd = &cobra.Command{
	Use:   "scan [root]",
	Short: "Scan a directory tree and output or write entries",
	Long: `Walk the directory tree starting from the given root (or category root
from the config) and output discovered entries.

By default, scan prints results to stdout and does not modify any files.
Use --write to merge scanned entries into the config file specified by -c.
Use --local with --write to save to .pregorc.yml in the current directory.
Git repositories are automatically detected — scan stops at repo boundaries
and captures the remote URL. Hidden directories like .git are skipped.

Examples:
  prego scan .                                # preview entries in current dir
  prego scan . --write --local               # write to local .pregorc.yml
  prego scan ~/repos -C repos --write         # write entries to system config
  prego scan ~/repos -d 2                     # limit depth
  prego scan -C core                          # scan the root of an existing category`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if scanLocal && !scanWrite {
			return fmt.Errorf("--local requires --write")
		}

		writePath := cfgPath
		if scanLocal {
			writePath = config.LocalConfigName
		}

		root := ""
		if len(args) > 0 {
			root = args[0]
		}

		if root == "" && scanCategory == "" {
			return fmt.Errorf("provide a root path or use --category to scan a config category")
		}

		if scanCategory != "" && root == "" {
			cfg, err := config.DiscoverConfig(cfgPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			cat, ok := cfg.Directory[scanCategory]
			if !ok {
				return fmt.Errorf("category %q not found in config", scanCategory)
			}
			root = cat.Root
		}

		root = config.ExpandPath(root)

		absRoot, err := filepath.Abs(root)
		if err != nil {
			return fmt.Errorf("failed to resolve path: %w", err)
		}

		result, err := fs.Scan(absRoot, scanDepth)
		if err != nil {
			return fmt.Errorf("scan failed: %w", err)
		}

		for _, ig := range result.Ignored {
			cmd.Printf("ignored %s (matched %q from %s)\n", ig.Path, ig.Pattern, ig.Source)
		}

		if len(result.Entries) == 0 {
			if len(result.Ignored) > 0 {
				cmd.Println("no directories found (some entries were ignored by .nosauce)")
			} else {
				cmd.Println("no directories found")
			}
			return nil
		}

		if scanWrite {
			return writeScanEntries(cmd, result.Entries, absRoot, writePath)
		}

		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "PATH\tMODE\tVCS\tREMOTE")
		for _, e := range result.Entries {
			fmt.Fprintf(w, "%s\t%04o\t%s\t%s\n", e.Path, e.Mode, e.VCS, e.Remote)
		}
		w.Flush()
		return nil
	},
}

func writeScanEntries(cmd *cobra.Command, entries []fs.ScanEntry, absRoot string, writePath string) error {
	category := scanCategory
	if category == "" {
		category = "repos"
	}

	var cfg *Config
	cfg, err := config.Load(writePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to load config: %w", err)
		}
		hostname, _ := os.Hostname()
		cfg = &config.Config{
			Version: config.Version,
			General: config.General{Color: true},
			System: config.System{
				Machine: config.Machine{Name: hostname, OS: runtime.GOOS},
				Hooks:   config.Hooks{},
			},
			Directory: map[string]config.DirCategory{},
		}
	}

	cat := config.DirCategory{
		Root:    config.ContractPath(absRoot),
		Entries: []config.DirEntry{},
	}

	if !scanLocal {
		if existing, ok := cfg.Directory[category]; ok {
			cat = existing
			cat.Root = config.ContractPath(absRoot)
		}
	} else {
		cat.Root = "."
	}

	seen := make(map[string]bool)
	for _, e := range cat.Entries {
		seen[config.ResolveEntryPath(e.Path, config.ResolveRoot(cat.Root))] = true
	}

	added := 0
	for _, entry := range entries {
		absPath := entry.Path
		rel, err := filepath.Rel(absRoot, entry.Path)
		if err != nil {
			rel = entry.Path
		}
		if absPath == rel || rel == "." {
			continue
		}

		resolved := config.ResolveEntryPath(rel, config.ResolveRoot(cat.Root))
		if seen[resolved] {
			continue
		}

		mode := entry.Mode
		if mode == 0 {
			mode = 0755
		}

		dirEntry := config.DirEntry{
			Path:   rel,
			Mode:   mode,
			VCS:    entry.VCS,
			Remote: entry.Remote,
		}

		if !scanLocal {
			dirEntry.Path = config.ContractPath(absPath)
		}

		cat.Entries = append(cat.Entries, dirEntry)
		resolved2 := config.ResolveEntryPath(dirEntry.Path, config.ResolveRoot(cat.Root))
		seen[resolved2] = true
		added++
	}

	cfg.Directory[category] = cat

	if err := config.Validate(cfg); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	if err := config.Save(writePath, cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	if scanLocal {
		cmd.Printf("wrote %d entries to category %q in %s\n", added, category, writePath)
	} else {
		cmd.Printf("added %d entries to category %q in %s\n", added, category, writePath)
	}
	return nil
}

type Config = config.Config

func init() {
	scanCmd.Flags().IntVarP(&scanDepth, "depth", "d", 0, "max traversal depth (0 = unlimited)")
	scanCmd.Flags().StringVarP(&scanCategory, "category", "C", "", "config category to scan/write into (core/documents/repos)")
	scanCmd.Flags().BoolVar(&scanWrite, "write", false, "write scanned entries into the config file")
	scanCmd.Flags().BoolVar(&scanLocal, "local", false, "write to .pregorc.yml in current directory (requires --write)")
	rootCmd.AddCommand(scanCmd)
}
