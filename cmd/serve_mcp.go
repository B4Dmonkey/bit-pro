package cmd

import "github.com/spf13/cobra"

const serveMCPCmdUse = "mcp"

func newServeMCPCmd() *cobra.Command {
	return &cobra.Command{
		Use:   serveMCPCmdUse,
		Short: "Run the MCP server in the foreground",
		Args:  cobra.NoArgs,
	}
}
