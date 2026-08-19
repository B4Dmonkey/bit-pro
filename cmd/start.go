package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/B4Dmonkey/bit-pro/daemon"
	"github.com/B4Dmonkey/bit-pro/store"
	"github.com/spf13/cobra"
)

const startCmdUse = "start"

func newStartCmd(lc daemon.Runner) *cobra.Command {
	return &cobra.Command{
		Use:   startCmdUse,
		Short: "Start the background daemon under launchd",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, err := enrollDaemon()
			if err != nil {
				return err
			}

			_, pid, alreadyRunning, err := daemon.Start(cmd.Context(), lc, path)
			if err != nil {
				return err
			}

			state := "started"
			if alreadyRunning {
				state = "already running"
			}

			fmt.Fprintf(cmd.OutOrStdout(), "%s (pid %d)\n", state, pid)

			return nil
		},
	}
}

func enrollDaemon() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("locating the running binary: %w", err)
	}

	dir, err := store.Dir()
	if err != nil {
		return "", err
	}

	path, err := daemon.PlistPath()
	if err != nil {
		return "", err
	}

	_, err = os.Stat(path)
	if err == nil {
		return path, nil
	}

	if !errors.Is(err, fs.ErrNotExist) {
		return "", fmt.Errorf("reading %s: %w", path, err)
	}

	if err := daemon.WritePlist(path, daemon.Plist(exe, filepath.Join(dir, "daemon.log"))); err != nil {
		return "", err
	}

	return path, nil
}
