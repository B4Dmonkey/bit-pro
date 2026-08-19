package cmd

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/B4Dmonkey/bit-pro/claude"
	"github.com/B4Dmonkey/bit-pro/launchd"
	"github.com/spf13/cobra"
)

var version = "dev"

var bitDir = ".bit"

const (
	claudeDir    = ".claude"
	worktreesDir = "worktrees"
)

func canonicalBitDir(wd string) string {
	sep := string(filepath.Separator)
	segments := strings.Split(wd, sep)

	for i := 0; i+1 < len(segments); i++ {
		if segments[i] == claudeDir && segments[i+1] == worktreesDir {
			return filepath.Join(strings.Join(segments[:i], sep), ".bit")
		}
	}

	return ".bit"
}

func signalContext() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
}

func Execute() error {
	ctx, stop := signalContext()
	defer stop()

	return NewRootCmd().ExecuteContext(ctx)
}

func NewRootCmd() *cobra.Command {
	return newRootCmd(claude.ExecRunner, launchd.ExecRunner)
}

func newRootCmd(run claude.Runner, lc launchd.Runner) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "bp",
		Short:         "bp is a project-management CLI for LLM-driven development workflows",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			bitDir = ".bit"
			if wd, err := os.Getwd(); err == nil {
				bitDir = canonicalBitDir(wd)
			}

			return nil
		},
	}
	rootCmd.AddCommand(newApproveCmd())
	rootCmd.AddCommand(newFeedbackCmd())
	rootCmd.AddCommand(newInitCmd(run))
	rootCmd.AddCommand(newInstructionsCmd())
	rootCmd.AddCommand(newServeCmd())
	rootCmd.AddCommand(newStatusCmd(lc))
	rootCmd.AddCommand(newTaskCmd())
	rootCmd.AddCommand(newTUICmd())
	rootCmd.AddCommand(newUnapproveCmd())

	return rootCmd
}
