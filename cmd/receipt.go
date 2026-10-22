package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/B4Dmonkey/bit-pro/db"
	"github.com/B4Dmonkey/bit-pro/db/orm"
	"github.com/B4Dmonkey/bit-pro/project"
	"github.com/B4Dmonkey/bit-pro/usage"
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

			p, err := project.Find(cmd.Context(), payload.Cwd)
			if errors.Is(err, project.ErrNotRegistered) ||
				errors.Is(err, project.ErrRemoved) ||
				errors.Is(err, project.ErrNeedsMigrate) {
				return nil
			}

			if err != nil {
				return err
			}

			return storeUsage(cmd.Context(), p.ID, payload)
		},
	}
}

func storeUsage(ctx context.Context, projectID int64, payload hookPayload) error {
	f, err := os.Open(payload.TranscriptPath)
	if err != nil {
		return fmt.Errorf("opening transcript: %w", err)
	}
	defer f.Close()

	entries, err := usage.Parse(f)
	if err != nil {
		return err
	}

	sqlDB, err := db.Open()
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	tx, err := sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	q := orm.New(tx)

	for _, e := range entries {
		if e.Track == "" {
			continue
		}

		err := q.UpsertUsage(ctx, orm.UpsertUsageParams{
			ProjectID:          projectID,
			SessionID:          payload.SessionID,
			Track:              e.Track,
			MessageID:          e.MessageID,
			RequestID:          e.RequestID,
			Ts:                 e.Timestamp,
			Model:              e.Model,
			Speed:              e.Speed,
			IsSidechain:        boolInt(e.IsSidechain),
			CcVersion:          e.Version,
			InputTokens:        e.InputTokens,
			CacheWrite5mTokens: e.CacheWrite5m,
			CacheWrite1hTokens: e.CacheWrite1h,
			CacheReadTokens:    e.CacheReadTokens,
			OutputTokens:       e.OutputTokens,
		})
		if err != nil {
			return fmt.Errorf("storing usage for %s: %w", e.MessageID, err)
		}
	}

	return tx.Commit()
}

func boolInt(b bool) int64 {
	if b {
		return 1
	}

	return 0
}
