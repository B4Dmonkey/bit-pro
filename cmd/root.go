package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/B4Dmonkey/bit-pro/bitdir"
	"github.com/B4Dmonkey/bit-pro/claude"
	taskcmd "github.com/B4Dmonkey/bit-pro/cmd/task"
	"github.com/B4Dmonkey/bit-pro/daemon"
	"github.com/spf13/cobra"
)

var version = "dev"

const claudeDir = ".claude"

func signalContext() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
}

func Execute() error {
	ctx, stop := signalContext()
	defer stop()

	return NewRootCmd().ExecuteContext(ctx)
}

func NewRootCmd() *cobra.Command {
	return newRootCmd(claude.ExecRunner, daemon.ExecRunner)
}

func newRootCmd(run claude.Runner, lc daemon.Runner) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "bp",
		Short:         "bp is a project-management CLI for LLM-driven development workflows",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			bitdir.Resolve()

			return nil
		},
	}
	rootCmd.AddCommand(newAddCmd(run))
	rootCmd.AddCommand(newApproveCmd())
	rootCmd.AddCommand(newFeedbackCmd())
	rootCmd.AddCommand(newInitCmd(run))
	rootCmd.AddCommand(newInstructionsCmd())
	rootCmd.AddCommand(newListCmd())
	rootCmd.AddCommand(newServeCmd())
	rootCmd.AddCommand(newStartCmd(lc))
	rootCmd.AddCommand(newStatusCmd(lc))
	rootCmd.AddCommand(newStopCmd(lc))
	rootCmd.AddCommand(taskcmd.NewCmd())
	rootCmd.AddCommand(newTUICmd())
	rootCmd.AddCommand(newUnapproveCmd())

	return rootCmd
}
