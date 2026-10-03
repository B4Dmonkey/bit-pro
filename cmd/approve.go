package cmd

import (
	"github.com/spf13/cobra"
)

func newApproveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "approve <id>",
		Short: "Approve a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return setApproved(cmd, args[0], true)
		},
	}
}

func newUnapproveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unapprove <id>",
		Short: "Revoke approval for a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return setApproved(cmd, args[0], false)
		},
	}
}

func setApproved(cmd *cobra.Command, id string, approved bool) error {
	s, err := openStore(cmd)
	if err != nil {
		return err
	}

	return s.SetApproved(id, approved)
}
