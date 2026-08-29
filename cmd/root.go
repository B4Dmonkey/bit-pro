package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/B4Dmonkey/bit-pro/bitdir"
	"github.com/B4Dmonkey/bit-pro/claude"
	taskcmd "github.com/B4Dmonkey/bit-pro/cmd/task"
	"github.com/B4Dmonkey/bit-pro/daemon"
	"github.com/spf13/cobra"
)

var version = "dev"

var pluginState = func() (installed, latest string, ok bool) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", false
	}

	return claude.PluginState(home, bitdir.Root())
}

var refreshMarketplace = claude.RefreshMarketplace

const claudeDir = ".claude"

const (
	quietAnnotation = "bit.quiet"
	quietEnabled    = "true"
)

func signalContext() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
}

func Execute() error {
	ctx, stop := signalContext()
	defer stop()

	return execute(ctx, NewRootCmd())
}

func execute(ctx context.Context, root *cobra.Command) error {
	cmd, err := root.ExecuteContextC(ctx)

	if suppressed(cmd) {
		return err
	}

	refreshMarketplace()

	if installed, latest, ok := pluginState(); ok && behind(installed, latest) {
		if cmd == nil {
			cmd = root
		}

		fmt.Fprintln(cmd.ErrOrStderr(), notice(installed, latest))
	}

	return err
}

func suppressed(cmd *cobra.Command) bool {
	if cmd == nil {
		return false
	}

	return cmd.Annotations[quietAnnotation] != ""
}

func behind(installed, latest string) bool {
	from, ok := parseVersion(installed)
	if !ok {
		return false
	}

	to, ok := parseVersion(latest)
	if !ok {
		return false
	}

	for i := range from {
		if from[i] != to[i] {
			return from[i] < to[i]
		}
	}

	return false
}

func parseVersion(v string) ([3]int, bool) {
	var parts [3]int

	fields := strings.Split(v, ".")
	if len(fields) != len(parts) {
		return parts, false
	}

	for i, f := range fields {
		n, err := strconv.Atoi(f)
		if err != nil {
			return parts, false
		}

		parts[i] = n
	}

	return parts, true
}

func notice(installed, latest string) string {
	const format = "bp: bit plugin %s → %s available — run: claude plugin update bit@bit-pro --scope project"

	return fmt.Sprintf(format, installed, latest)
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
