package cmd

import (
	"log/slog"
	"time"

	"github.com/spf13/cobra"
)

const serveCmdUse = "serve"

var serveTick = 30 * time.Second

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

			log := slog.New(slog.NewTextHandler(cmd.OutOrStdout(), &slog.HandlerOptions{Level: level}))

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
