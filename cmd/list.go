package cmd

import (
	"fmt"

	"github.com/B4Dmonkey/bit-pro/db"
	"github.com/B4Dmonkey/bit-pro/db/orm"
	"github.com/spf13/cobra"
)

const listCmdUse = "list"

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   listCmdUse,
		Short: "List the projects enrolled in the registry",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			sqlDB, err := db.Open()
			if err != nil {
				return err
			}
			defer sqlDB.Close()

			projects, err := orm.New(sqlDB).ListProjects(cmd.Context())
			if err != nil {
				return fmt.Errorf("listing projects: %w", err)
			}

			out := cmd.OutOrStdout()

			for _, p := range projects {
				fmt.Fprintf(out, "%s\t%s\n", p.Code, p.Path)
			}

			return nil
		},
	}
}
