package cmd

import (
	"github.com/B4Dmonkey/bit-pro/bitdir"
	"github.com/B4Dmonkey/bit-pro/task"
	"github.com/spf13/cobra"
)

func newApproveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "approve <id>",
		Short: "Approve a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return task.New(bitdir.Current()).SetApproved(args[0], true)
		},
	}
}

func newUnapproveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unapprove <id>",
		Short: "Revoke approval for a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return task.New(bitdir.Current()).SetApproved(args[0], false)
		},
	}
}
