package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/B4Dmonkey/bit-pro/launchd"
	"github.com/B4Dmonkey/bit-pro/store"
	"github.com/spf13/cobra"
)

const startCmdUse = "start"

func newStartCmd(lc launchd.Runner) *cobra.Command {
	return &cobra.Command{
		Use:   startCmdUse,
		Short: "Start the background daemon under launchd",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, err := enrollDaemon()
			if err != nil {
				return err
			}

			_, pid, err := launchd.Start(cmd.Context(), lc, path)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "started (pid %d)\n", pid)

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

	path, err := launchd.PlistPath()
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

	if err := launchd.WritePlist(path, launchd.Plist(exe, filepath.Join(dir, "daemon.log"))); err != nil {
		return "", err
	}

	return path, nil
}
