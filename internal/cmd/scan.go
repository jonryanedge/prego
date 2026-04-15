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
)

var scanCmd = &cobra.Command{
	Use:   "scan [root]",
	Short: "Scan a directory tree and output or write entries",
	Long: `Walk the directory tree starting from the given root (or category root
from the config) and output discovered entries.

By default, scan prints results to stdout and does not modify any files.
Use --write to merge scanned entries into the config file specified by -c.
Git repositories are automatically detected and their remote URLs captured.

Examples:
  prego scan ~/repos                      # preview entries
  prego scan ~/repos -C repos --write     # write entries to config under "repos"
  prego scan ~/repos -d 2                 # limit depth
  prego scan -C core                      # scan the root of an existing category`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root := ""
		if len(args) > 0 {
			root = args[0]
		}

		if root == "" && scanCategory == "" {
			return fmt.Errorf("provide a root path or use --category to scan a config category")
		}

		if scanCategory != "" && root == "" {
			cfg, err := config.Load(cfgPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			cat, ok := cfg.Dirs[scanCategory]
			if !ok {
				return fmt.Errorf("category %q not found in config", scanCategory)
			}
			root = cat.Root
		}

		root = config.ExpandPath(root)

		entries, err := fs.Scan(root, scanDepth)
		if err != nil {
			return fmt.Errorf("scan failed: %w", err)
		}

		if len(entries) == 0 {
			cmd.Println("no directories found")
			return nil
		}

		enrichedEntries := detectVCSDetails(entries)

		if scanWrite {
			return writeScanEntries(cmd, enrichedEntries, root)
		}

		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "PATH\tMODE\tVCS\tREMOTE")
		for _, e := range enrichedEntries {
			fmt.Fprintf(w, "%s\t%04o\t%s\t%s\n", e.Path, e.Mode, e.VCS, e.Remote)
		}
		w.Flush()
		return nil
	},
}

type enrichedEntry struct {
	Path   string
	Mode   uint32
	VCS    string
	Remote string
}

func detectVCSDetails(entries []fs.ScanEntry) []enrichedEntry {
	result := make([]enrichedEntry, len(entries))
	for i, e := range entries {
		result[i] = enrichedEntry{
			Path: e.Path,
			Mode: e.Mode,
		}
		vcs, remote := fs.DetectVCS(e.Path)
		if vcs != "" {
			result[i].VCS = vcs
			result[i].Remote = remote
		}
	}
	return result
}

func writeScanEntries(cmd *cobra.Command, entries []enrichedEntry, root string) error {
	category := scanCategory
	if category == "" {
		category = "repos"
	}

	var cfg *Config
	cfg, err := config.Load(cfgPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to load config: %w", err)
		}
		hostname, _ := os.Hostname()
		cfg = &config.Config{
			Version: config.Version,
			Machine: config.Machine{
				Name: hostname,
				OS:   runtime.GOOS,
			},
			Dirs:  map[string]config.DirCategory{},
			Hooks: config.Hooks{},
		}
	}

	cat, exists := cfg.Dirs[category]
	if !exists {
		cat = config.DirCategory{
			Root:    root,
			Entries: []config.DirEntry{},
		}
	}

	existing := make(map[string]bool)
	for _, e := range cat.Entries {
		existing[config.ExpandPath(e.Path)] = true
	}

	added := 0
	for _, entry := range entries {
		rel, err := filepath.Rel(root, entry.Path)
		if err != nil {
			rel = entry.Path
		}

		absPath := entry.Path
		if absPath == rel || rel == "." {
			continue
		}

		if existing[absPath] {
			continue
		}

		mode := entry.Mode
		if mode == 0 {
			mode = 0755
		}

		dirEntry := config.DirEntry{
			Path:   absPath,
			Mode:   mode,
			VCS:    entry.VCS,
			Remote: entry.Remote,
		}

		cat.Entries = append(cat.Entries, dirEntry)
		existing[absPath] = true
		added++
	}

	cfg.Dirs[category] = cat

	if err := config.Validate(cfg); err != nil {
		return fmt.Errorf("config validation failed after adding entries: %w", err)
	}

	if err := config.Save(cfgPath, cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	cmd.Printf("added %d entries to category %q in %s\n", added, category, cfgPath)
	return nil
}

type Config = config.Config

func init() {
	scanCmd.Flags().IntVarP(&scanDepth, "depth", "d", 0, "max traversal depth (0 = unlimited)")
	scanCmd.Flags().StringVarP(&scanCategory, "category", "C", "", "config category to scan/write into (core/documents/repos)")
	scanCmd.Flags().BoolVar(&scanWrite, "write", false, "write scanned entries into the config file")
	rootCmd.AddCommand(scanCmd)
}
