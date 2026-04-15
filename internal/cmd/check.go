package cmd

import (
	"fmt"

	"github.com/jonryanedge/prego/internal/config"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Validate the config file",
	Long:  `Parse and validate the prego config file. Reports any errors in structure, paths, or fields.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if err := config.Validate(cfg); err != nil {
			return fmt.Errorf("config validation failed: %w", err)
		}

		cmd.Println("config is valid")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}
