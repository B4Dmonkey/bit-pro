package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/B4Dmonkey/bit-pro/claude"
	taskcmd "github.com/B4Dmonkey/bit-pro/cmd/task"
	"github.com/spf13/cobra"
)

var version = "dev"

func signalContext() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
}

func Execute() error {
	ctx, stop := signalContext()
	defer stop()

	return NewRootCmd().ExecuteContext(ctx)
}

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
	rootCmd.AddCommand(newAddCmd(run))
	rootCmd.AddCommand(newListCmd())
	rootCmd.AddCommand(newMigrateCmd(run))
	rootCmd.AddCommand(newReceiptCmd())
	rootCmd.AddCommand(newServeCmd())
	rootCmd.AddCommand(taskcmd.NewCmd())
	rootCmd.AddCommand(newUICmd())

	return rootCmd
}
