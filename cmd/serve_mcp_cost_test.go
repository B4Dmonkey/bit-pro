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
		ProjectID: p.ID, Track: testTrackID, MessageID: id, RequestID: "req_" + id, Ts: testUsageTs,
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
		for _, want := range []string{"| Total cost | $5.20 |", "| Total tokens | 2.10M (in 2.00M · out 100K) |"} {
			if !strings.Contains(md, want) {
				t.Errorf("markdown = %q, want it to contain %q", md, want)
			}
		}
	})
}
