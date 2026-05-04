package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/jonryanedge/prego/internal/config"
	"github.com/spf13/cobra"
)

var initLocal bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a default config file",
	Long: `Create a prego config file with default values. By default, creates
the system config at ~/.pregorc.yml. Use --local to create a .pregorc.yml
in the current directory instead.

If the file already exists, init will not overwrite it.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var path string
		if initLocal {
			path = config.LocalConfigName
		} else {
			path = cfgPath
		}

		expanded := config.ExpandPath(path)
		if _, err := os.Stat(expanded); err == nil {
			return fmt.Errorf("config file already exists at %s", path)
		}

		hostname, _ := os.Hostname()
		cfg := &config.Config{
			Version: config.Version,
			General: config.General{
				Color:   true,
				Verbose: false,
			},
			System: config.System{
				Machine: config.Machine{
					Name: hostname,
					OS:   runtime.GOOS,
				},
				Hooks: config.Hooks{},
			},
			Directory: map[string]config.DirCategory{
				"core": {
					Root:    "~",
					Entries: []config.DirEntry{},
				},
				"documents": {
					Root:    "~/Documents",
					Entries: []config.DirEntry{},
				},
				"repos": {
					Root:    "~/repos",
					Entries: []config.DirEntry{},
				},
			},
		}

		if err := config.Save(path, cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		cmd.Printf("created config at %s\n", path)
		return nil
	},
}

func init() {
	initCmd.Flags().BoolVar(&initLocal, "local", false, "create config in current directory instead of home")
	rootCmd.AddCommand(initCmd)
}
