package cmd

import "github.com/spf13/cobra"

const version = "0.1.0-dev"

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "bit",
		Short:   "bit is a project-management CLI for LLM-driven development workflows",
		Version: version,
	}
	rootCmd.AddCommand(newInitCmd())
	rootCmd.AddCommand(newTaskCmd())
	return rootCmd
}
