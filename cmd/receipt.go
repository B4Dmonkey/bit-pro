package cmd

import (
	"encoding/json"
	"errors"

	"github.com/B4Dmonkey/bit-pro/project"
	"github.com/spf13/cobra"
)

type hookPayload struct {
	SessionID           string `json:"session_id"`
	TranscriptPath      string `json:"transcript_path"`
	Cwd                 string `json:"cwd"`
	HookEventName       string `json:"hook_event_name"`
	AgentID             string `json:"agent_id"`
	AgentType           string `json:"agent_type"`
	AgentTranscriptPath string `json:"agent_transcript_path"`
}

func newReceiptCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "receipt",
		Short:  "Record usage from a Stop or SubagentStop hook payload",
		Hidden: true,
		Args:   cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			var payload hookPayload
			if err := json.NewDecoder(cmd.InOrStdin()).Decode(&payload); err != nil {
				return err
			}

			_, err := project.Find(cmd.Context(), payload.Cwd)
			if errors.Is(err, project.ErrNotRegistered) ||
				errors.Is(err, project.ErrRemoved) ||
				errors.Is(err, project.ErrNeedsMigrate) {
				return nil
			}

			return err
		},
	}
}
