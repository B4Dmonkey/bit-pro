package cmd

import (
	"os"

	"github.com/B4Dmonkey/bit-pro/claude"
	"github.com/spf13/cobra"
)

var version = "dev"

var bitDir = ".bit"

const claudeDir = ".claude"

func NewRootCmd() *cobra.Command {
	return newRootCmd(claude.ExecRunner)
}

func newRootCmd(run claude.Runner) *cobra.Command {
	var bitDirFlag string

	rootCmd := &cobra.Command{
		Use:           "bp",
		Short:         "bp is a project-management CLI for LLM-driven development workflows",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			bitDir = ".bit"
			if bitDirFlag != "" {
				bitDir = bitDirFlag
			} else if v := os.Getenv("BIT_DIR"); v != "" {
				bitDir = v
			}

			return nil
		},
	}
	rootCmd.PersistentFlags().StringVar(&bitDirFlag, "bit-dir", "", "canonical .bit directory (overrides BIT_DIR)")
	rootCmd.AddCommand(newApproveCmd())
	rootCmd.AddCommand(newFeedbackCmd())
	rootCmd.AddCommand(newInitCmd(run))
	rootCmd.AddCommand(newInstructionsCmd())
	rootCmd.AddCommand(newTaskCmd())
	rootCmd.AddCommand(newTUICmd())
	rootCmd.AddCommand(newUnapproveCmd())

	return rootCmd
}
