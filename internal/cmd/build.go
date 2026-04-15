package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/jonryanedge/prego/internal/config"
	"github.com/jonryanedge/prego/internal/fs"
	"github.com/spf13/cobra"
)

var buildDryRun bool

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Apply directory structure and clone git repos",
	Long: `Apply the full directory structure from the config (like 'prego apply')
and then clone any git repos that have a remote URL but don't exist yet.

Idempotent: safe to run multiple times. Existing directories and repos
are skipped.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		if err := config.Validate(cfg); err != nil {
			return fmt.Errorf("config validation failed: %w", err)
		}

		var created, skipped, linked int
		var cloned []string

		for cat, dirCat := range cfg.Dirs {
			for _, entry := range dirCat.Entries {
				expanded := config.ExpandPath(entry.Path)
				mode := entry.Mode
				if mode == 0 {
					mode = 0755
				}

				if buildDryRun {
					cmd.Printf("[dry-run] would create directory %s (mode %04o)\n", expanded, mode)
					created++
					continue
				}

				info, statErr := os.Stat(expanded)
				if statErr == nil {
					if info.IsDir() {
						cmd.Printf("ok  %s (already exists)\n", expanded)
						skipped++
					} else {
						cmd.Printf("err %s (exists as file)\n", expanded)
						skipped++
					}
					continue
				}

				if err := fs.MkdirAll(expanded, mode); err != nil {
					return fmt.Errorf("[%s] failed to create %s: %w", cat, expanded, err)
				}
				cmd.Printf("created %s (mode %04o)\n", expanded, mode)
				created++
			}

			for _, sl := range dirCat.Symlinks {
				from := config.ExpandPath(sl.From)
				to := config.ExpandPath(sl.To)

				if buildDryRun {
					cmd.Printf("[dry-run] would create symlink %s -> %s\n", to, from)
					linked++
					continue
				}

				err := fs.Symlink(from, to)
				if err != nil {
					cmd.Printf("err symlink %s: %v\n", to, err)
					continue
				}
				cmd.Printf("linked %s -> %s\n", to, from)
				linked++
			}
		}

		for cat, dirCat := range cfg.Dirs {
			for _, entry := range dirCat.Entries {
				if entry.VCS != "git" || entry.Remote == "" {
					continue
				}

				expanded := config.ExpandPath(entry.Path)

				if buildDryRun {
					cmd.Printf("[dry-run] would clone %s into %s\n", entry.Remote, expanded)
					cloned = append(cloned, entry.Remote)
					continue
				}

				isDir := false
				if info, err := os.Stat(expanded); err == nil && info.IsDir() {
					isDir = true
					if fs.IsGitRepo(expanded) {
						cmd.Printf("ok  %s (repo exists)\n", expanded)
						continue
					}
				}

				if isDir {
					dirEntries, err := os.ReadDir(expanded)
					if err != nil {
						cmd.Printf("err cannot read %s: %v\n", expanded, err)
						continue
					}
					if len(dirEntries) > 0 {
						cmd.Printf("skip %s (directory not empty, cannot clone)\n", expanded)
						continue
					}
				} else {
					parent := ""
					for i := len(expanded) - 1; i >= 0; i-- {
						if expanded[i] == '/' {
							parent = expanded[:i]
							break
						}
					}
					if parent != "" {
						if err := os.MkdirAll(parent, 0755); err != nil {
							cmd.Printf("err creating parent %s: %v\n", parent, err)
							continue
						}
					}
				}

				cloneURL := entry.Remote

				gitCmd := exec.Command("git", "clone", cloneURL, expanded)
				output, err := gitCmd.CombinedOutput()
				if err != nil {
					cmd.Printf("err cloning %s: %v\n%s\n", cloneURL, err, string(output))
					continue
				}
				cmd.Printf("cloned %s into %s\n", cloneURL, expanded)
				cloned = append(cloned, cloneURL)
			}

			if len(cfg.Hooks.PostCreate) > 0 {
				for _, hook := range cfg.Hooks.PostCreate {
					if buildDryRun {
						cmd.Printf("[dry-run] would run: %s\n", hook)
						continue
					}
					cmd.Printf("hook  %s\n", hook)
					if err := exec.Command("sh", "-c", hook).Run(); err != nil {
						cmd.Printf("hook failed: %v\n", err)
					}
				}
			}
			_ = cat
		}

		if !buildDryRun {
			cmd.Printf("\ncreated %d dirs, skipped %d, linked %d, cloned %d repos\n", created, skipped, linked, len(cloned))
		}
		return nil
	},
}

func init() {
	buildCmd.Flags().BoolVar(&buildDryRun, "dry-run", false, "preview changes without making them")
	rootCmd.AddCommand(buildCmd)
}
