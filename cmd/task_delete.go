package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func newTaskDeleteCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			if !yes {
				fmt.Fprintf(cmd.OutOrStdout(), "Delete task %s? [y/N] ", id)
				line, _ := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
				answer := strings.ToLower(strings.TrimSpace(line))
				if answer != "y" && answer != "yes" {
					fmt.Fprintln(cmd.OutOrStdout(), "cancelled")
					return nil
				}
			}
			return os.Remove(taskPath(id))
		},
	}
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "skip confirmation")
	return cmd
}
