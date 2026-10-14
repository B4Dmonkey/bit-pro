package cmd

import (
	"github.com/B4Dmonkey/bit-pro/project"
	"github.com/B4Dmonkey/bit-pro/tui"
	"github.com/spf13/cobra"
)

const uiCmdUse = "ui"

func newUICmd() *cobra.Command {
	return &cobra.Command{
		Use:   uiCmdUse,
		Short: "Browse tasks in a terminal UI",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			stores, err := project.OpenAll(cmd.Context())
			if err != nil {
				return err
			}

			tasks, err := stores.List()
			if err != nil {
				return err
			}

			return tui.Run(tui.New(tasks).
				WithReload(stores.List).
				WithApprove(func(id string, approved bool) error {
					s, err := stores.For(id)
					if err != nil {
						return err
					}

					return s.SetApproved(id, approved)
				}))
		},
	}
}
