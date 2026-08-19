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

func newStartCmd(_ launchd.Runner) *cobra.Command {
	return &cobra.Command{
		Use:   startCmdUse,
		Short: "Start the background daemon under launchd",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return enrollDaemon()
		},
	}
}

func enrollDaemon() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locating the running binary: %w", err)
	}

	dir, err := store.Dir()
	if err != nil {
		return err
	}

	path, err := launchd.PlistPath()
	if err != nil {
		return err
	}

	_, err = os.Stat(path)
	if err == nil {
		return nil
	}

	if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("reading %s: %w", path, err)
	}

	return launchd.WritePlist(path, launchd.Plist(exe, filepath.Join(dir, "daemon.log")))
}
