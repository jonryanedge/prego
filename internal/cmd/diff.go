package cmd

import (
	"fmt"
	"text/tabwriter"

	"github.com/jonryanedge/prego/internal/config"
	"github.com/jonryanedge/prego/internal/fs"
	"github.com/spf13/cobra"
)

var diffExitCode bool

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Compare local filesystem state against config",
	Long: `Compare what the config declares against what actually exists on disk.
Reports missing directories, extra files, permission mismatches, and
symlink drift. Exit code 0 if no drift, 1 if drift found.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		if err := config.Validate(cfg); err != nil {
			return fmt.Errorf("config validation failed: %w", err)
		}

		drifts := fs.Diff(cfg)

		if len(drifts) == 0 {
			cmd.Println("no drift found — filesystem matches config")
			return nil
		}

		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "TYPE\tCATEGORY\tPATH\tEXPECTED\tACTUAL")
		for _, d := range drifts {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", d.Type, d.Category, d.Path, d.Expected, d.Actual)
		}
		w.Flush()

		if diffExitCode {
			cmd.SilenceErrors = true
		}
		return &driftError{count: len(drifts)}
	},
}

type driftError struct {
	count int
}

func (e *driftError) Error() string {
	return fmt.Sprintf("%d drift(s) found", e.count)
}

func init() {
	diffCmd.Flags().BoolVar(&diffExitCode, "exit-code", true, "exit with code 1 when drift is found")
	rootCmd.AddCommand(diffCmd)
}

func setIsExitCode(val bool) {
	diffExitCode = val
}
