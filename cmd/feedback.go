package cmd

import "github.com/spf13/cobra"

func newFeedbackCmd() *cobra.Command {
	feedbackCmd := &cobra.Command{
		Use:   "feedback",
		Short: "Record feedback notes",
	}
	feedbackCmd.AddCommand(newFeedbackAddCmd())
	return feedbackCmd
}
