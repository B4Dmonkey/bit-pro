package cmd

import (
	"github.com/B4Dmonkey/bit-pro/claude"
	"github.com/spf13/cobra"
)

var version = "dev"

const bitDir = ".bit"

const claudeDir = ".claude"

func NewRootCmd() *cobra.Command {
	return newRootCmd(claude.ExecRunner)
}

func newRootCmd(run claude.Runner) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "bp",
		Short:         "bp is a project-management CLI for LLM-driven development workflows",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	rootCmd.AddCommand(newInitCmd(run))
	rootCmd.AddCommand(newInstructionsCmd())
	rootCmd.AddCommand(newTaskCmd())
	rootCmd.AddCommand(newTUICmd())
	return rootCmd
}
