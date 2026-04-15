package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var cfgPath string

var rootCmd = &cobra.Command{
	Use:   "prego",
	Short: "Save and reproduce directory structures across machines",
	Long: `Prego captures, stores, and replicates your directory structures
across multiple machines. Configuration is stored in a single
dotfile (~/.pregorc.yml) that can be version-controlled and shared.`,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgPath, "config", "c", "~/.pregorc.yml", "path to config file")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
