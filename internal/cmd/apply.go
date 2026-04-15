package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/jonryanedge/prego/internal/config"
	"github.com/jonryanedge/prego/internal/fs"
	"github.com/spf13/cobra"
)

var applyDryRun bool

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Create directories and symlinks from the config",
	Long: `Read the config file and create all directories and symlinks that
don't already exist. Respects declared permissions and symlink targets.
Idempotent: safe to run multiple times.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		if err := config.Validate(cfg); err != nil {
			return fmt.Errorf("config validation failed: %w", err)
		}

		var created, skipped, linked int

		for cat, dirCat := range cfg.Dirs {
			for _, entry := range dirCat.Entries {
				expanded := config.ExpandPath(entry.Path)
				mode := entry.Mode
				if mode == 0 {
					mode = 0755
				}

				if applyDryRun {
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

				if applyDryRun {
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

		if len(cfg.Hooks.PostCreate) > 0 {
			for _, hook := range cfg.Hooks.PostCreate {
				if applyDryRun {
					cmd.Printf("[dry-run] would run: %s\n", hook)
					continue
				}
				cmd.Printf("hook  %s\n", hook)
				if err := exec.Command("sh", "-c", hook).Run(); err != nil {
					cmd.Printf("hook failed: %v\n", err)
				}
			}
		}

		if !applyDryRun {
			cmd.Printf("\ncreated %d, skipped %d, linked %d\n", created, skipped, linked)
		}
		return nil
	},
}

func init() {
	applyCmd.Flags().BoolVar(&applyDryRun, "dry-run", false, "preview changes without making them")
	rootCmd.AddCommand(applyCmd)
}

func resetFlags() {
	applyDryRun = false
	buildDryRun = false
	diffExitCode = true
	scanDepth = 0
	scanCategory = ""
	scanWrite = false
	cfgPath = "~/.pregorc.yml"
}
