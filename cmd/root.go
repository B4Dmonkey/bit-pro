package cmd

import "github.com/spf13/cobra"

const version = "0.1.0-dev"

func NewRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "bit",
		Short:   "bit is a project-management CLI for LLM-driven development workflows",
		Version: version,
	}
}
