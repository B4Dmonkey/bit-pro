package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/B4Dmonkey/bit-pro/db"
	"github.com/B4Dmonkey/bit-pro/db/orm"
	"github.com/B4Dmonkey/bit-pro/git"
	"github.com/B4Dmonkey/bit-pro/migrate"
	"github.com/spf13/cobra"
)

func newMigrateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Copy this folder's v1 .bit/ into the central store and register it",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			wd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("reading the working directory: %w", err)
			}

			sqlDB, err := db.Open()
			if err != nil {
				return err
			}
			defer sqlDB.Close()

			res, err := migrate.Run(cmd.Context(), orm.New(sqlDB), migrate.Options{Dir: wd, Git: git.ExecRunner, Now: time.Now})
			if err != nil {
				return err
			}

			if res.Already {
				fmt.Fprintln(cmd.OutOrStdout(), "already migrated")

				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "migrated %s %s\n", res.Code, res.Path)

			return nil
		},
	}
}
