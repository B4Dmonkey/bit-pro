package cmd

import (
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/spf13/cobra"
)

const serveCmdUse = "serve"

var serveTick = 10 * time.Second

func newHandler(w io.Writer, level slog.Level) slog.Handler {
	opts := &slog.HandlerOptions{Level: level}

	if f, ok := w.(*os.File); ok {
		if info, err := f.Stat(); err == nil && info.Mode()&os.ModeCharDevice != 0 {
			return slog.NewTextHandler(w, opts)
		}
	}

	return slog.NewJSONHandler(w, opts)
}

func newServeCmd() *cobra.Command {
	var verbose bool

	cmd := &cobra.Command{
		Use:   serveCmdUse,
		Short: "Run the automation loop in the foreground",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			level := slog.LevelInfo
			if verbose {
				level = slog.LevelDebug
			}

			log := slog.New(newHandler(cmd.OutOrStdout(), level))

			log.Info("started")
			defer log.Info("stopped")

			ticker := time.NewTicker(serveTick)
			defer ticker.Stop()

			ctx := cmd.Context()

			for {
				select {
				case <-ctx.Done():
					return nil
				case <-ticker.C:
					log.Debug("tick")
				}
			}
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Log each tick of the loop")

	return cmd
}
