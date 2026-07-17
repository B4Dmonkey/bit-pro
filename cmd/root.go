package cmd

import "github.com/spf13/cobra"

var version = "dev"

const bitDir = ".bit"

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "bit",
		Short:         "bit is a project-management CLI for LLM-driven development workflows",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	rootCmd.AddCommand(newInitCmd())
	rootCmd.AddCommand(newTaskCmd())
	return rootCmd
}
