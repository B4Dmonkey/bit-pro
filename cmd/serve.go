package cmd

import (
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/B4Dmonkey/bit-pro/claude"
	"github.com/B4Dmonkey/bit-pro/daemon"
	"github.com/B4Dmonkey/bit-pro/db"
	"github.com/B4Dmonkey/bit-pro/db/orm"
)

const (
	serveCmdUse       = "serve"
	serveDaemonCmdUse = "daemon"
)

var serveTick = 5 * time.Second

func newHandler(w io.Writer, level slog.Level) slog.Handler {
	opts := &slog.HandlerOptions{Level: level}

	if f, ok := w.(*os.File); ok {
		if info, err := f.Stat(); err == nil && info.Mode()&os.ModeCharDevice != 0 {
			return slog.NewTextHandler(w, opts)
		}
	}

	return slog.NewJSONHandler(w, opts)
}

func newServeDaemonCmd() *cobra.Command {
	var verbose bool

	cmd := &cobra.Command{
		Use:   serveDaemonCmdUse,
		Short: "Run the automation loop in the foreground",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			level := slog.LevelInfo
			if verbose {
				level = slog.LevelDebug
			}

			log := slog.New(newHandler(cmd.OutOrStdout(), level))

			sqlDB, err := db.Open()
			if err != nil {
				return err
			}

			defer sqlDB.Close()

			queries := orm.New(sqlDB)

			return daemon.Loop(cmd.Context(), queries, log, serveTick, claude.ExecDirRunner)
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Log each tick of the loop")

	return cmd
}

func newServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   serveCmdUse,
		Short: "Run foreground servers",
	}

	cmd.AddCommand(newServeDaemonCmd())
	cmd.AddCommand(newServeMCPCmd())

	return cmd
}
