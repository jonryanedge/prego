package cmd

import "github.com/spf13/cobra"

var version string = "0.1.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Printf("prego v%s\n", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
