package cmd

import (
	"github.com/B4Dmonkey/bit-pro/tui"
	"github.com/spf13/cobra"
)

const tuiCmdUse = "tui"

func newTUICmd() *cobra.Command {
	return &cobra.Command{
		Use:         tuiCmdUse,
		Short:       "Browse tasks in a terminal UI",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{quietAnnotation: quietEnabled},
		RunE: func(cmd *cobra.Command, _ []string) error {
			s, err := openStore(cmd)
			if err != nil {
				return err
			}

			tasks, err := s.List()
			if err != nil {
				return err
			}

			return tui.Run(tui.New(tasks).
				WithReload(s.List).
				WithApprove(func(id string, approved bool) error { return s.SetApproved(id, approved) }))
		},
	}
}
