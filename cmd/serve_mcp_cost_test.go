package cmd

import (
	"strings"
	"testing"

	"github.com/B4Dmonkey/bit-pro/db"
	"github.com/B4Dmonkey/bit-pro/db/orm"
	"github.com/B4Dmonkey/bit-pro/project"
	"github.com/B4Dmonkey/bit-pro/task"
)

const (
	testMarkdownKey = "markdown"
	testUsageTs     = "2026-10-22T10:00:00Z"
)

func seedUsage(t *testing.T, dir, id, model string, in, cacheRead, out int64) {
	t.Helper()

	seedBarUsage(t, dir, id, "", model, in, cacheRead, out)
}

func seedBarUsage(t *testing.T, dir, id, bar, model string, in, cacheRead, out int64) {
	t.Helper()

	p, err := project.Find(t.Context(), dir)
	if err != nil {
		t.Fatalf("project.Find(%q) returned error: %v", dir, err)
	}

	sqlDB, err := db.Open()
	if err != nil {
		t.Fatalf("db.Open() returned error: %v", err)
	}
	defer sqlDB.Close()

	err = orm.New(sqlDB).UpsertUsage(t.Context(), orm.UpsertUsageParams{
		ProjectID: p.ID, Track: testTrackID, Bar: bar, MessageID: id, RequestID: "req_" + id, Ts: testUsageTs,
		Model: model, InputTokens: in, CacheReadTokens: cacheRead, OutputTokens: out,
	})
	if err != nil {
		t.Fatalf("UpsertUsage() returned error: %v", err)
	}
}

func TestCostRead(t *testing.T) {
	t.Run("totals across models", func(t *testing.T) {
		dir := t.TempDir()
		seedTasks(t, dir, &task.Task{ID: testTrackID, Title: testTitle, Status: task.StatusDoing})
		seedUsage(t, dir, "m1", "claude-opus-5-5", 1_000_000, 0, 0)
		seedUsage(t, dir, "m2", "claude-sonnet-5", 0, 1_000_000, 100_000)

		got := callTool(t, mcpSession(t, dir), costReadTool, map[string]any{testTrackKey: testTrackID})

		md, _ := got[testMarkdownKey].(string)
		for _, want := range []string{
			"_as of 2026-10-22_", "| Total cost | $5.20 |", "| Total tokens | 2.10M (in 2.00M · out 100K) |",
		} {
			if !strings.Contains(md, want) {
				t.Errorf("markdown = %q, want it to contain %q", md, want)
			}
		}
	})

	t.Run("by verse and bar", func(t *testing.T) {
		dir := t.TempDir()
		seedTasks(t, dir,
			&task.Task{ID: testTrackID, Title: testTitle, Status: task.StatusDoing},
			&task.Task{
				ID: "FOO-1.1", Title: "a", Status: task.StatusDone, Phase: 2, PhaseLabel: "track total and per-agent cost",
			},
			&task.Task{
				ID: "FOO-1.2", Title: "b", Status: task.StatusDoing, Phase: 3, PhaseLabel: "nested agents and bars",
			},
		)
		seedUsage(t, dir, "m1", "claude-opus-5-5", 1_000_000, 0, 0)
		seedBarUsage(t, dir, "m2", "FOO-1.1", "claude-opus-5-5", 500_000, 0, 0)
		seedBarUsage(t, dir, "m3", "FOO-1.2", "claude-opus-5-5", 250_000, 0, 0)

		got := callTool(t, mcpSession(t, dir), costReadTool, map[string]any{testTrackKey: testTrackID})

		md, _ := got[testMarkdownKey].(string)
		for _, want := range []string{
			"### By verse / bar",
			"| **Verse 2 — track total and per-agent cost** | | **$2.00** |",
			"| ↳ Bar 1.1 | | $2.00 |",
			"| **Verse 3 — nested agents and bars** | | **$1.00** |",
			"| ↳ Bar 1.2 | | $1.00 |",
			"| Total cost | $7.00 |",
			"_bar-level work only — the agent totals above include it too_",
		} {
			if !strings.Contains(md, want) {
				t.Errorf("markdown = %q, want it to contain %q", md, want)
			}
		}
	})

	t.Run("completed track", func(t *testing.T) {
		dir := t.TempDir()
		seedTasks(t, dir, &task.Task{ID: testTrackID, Title: testTitle, Status: task.StatusDone})
		seedUsage(t, dir, "m1", "claude-opus-5-5", 1_000_000, 0, 0)

		if err := openProjectStore(t, dir).Complete(testTrackID); err != nil {
			t.Fatal(err)
		}

		got := callTool(t, mcpSession(t, dir), costReadTool, map[string]any{testTrackKey: testTrackID})

		md, _ := got[testMarkdownKey].(string)
		if want := "_as of 2026-10-22 (track complete)_"; !strings.Contains(md, want) {
			t.Errorf("markdown = %q, want it to contain %q", md, want)
		}
	})
}
