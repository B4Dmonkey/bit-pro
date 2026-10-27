package cmd

import (
	"strings"
	"testing"

	"github.com/B4Dmonkey/bit-pro/db"
	"github.com/B4Dmonkey/bit-pro/db/orm"
	"github.com/B4Dmonkey/bit-pro/project"
	"github.com/B4Dmonkey/bit-pro/task"
)

func TestCostFor(t *testing.T) {
	mcpSandbox(t)

	bitDir, acmDir := t.TempDir(), t.TempDir()

	for code, dir := range map[string]string{"BIT": bitDir, "ACM": acmDir} {
		path, err := project.CanonicalPath(dir)
		if err != nil {
			t.Fatalf("CanonicalPath(%q) returned error: %v", dir, err)
		}

		seedProject(t, orm.CreateProjectParams{Path: path, Code: code})
	}

	for id, dir := range map[string]string{testOwnTrack: bitDir, "ACM-1": acmDir} {
		if err := openProjectStore(t, dir).Save(&task.Task{ID: id, Title: testTitle, Status: task.StatusDoing}); err != nil {
			t.Fatal(err)
		}
	}

	acm, err := project.Find(t.Context(), acmDir)
	if err != nil {
		t.Fatal(err)
	}

	sqlDB, err := db.Open()
	if err != nil {
		t.Fatal(err)
	}

	err = orm.New(sqlDB).UpsertUsage(t.Context(), orm.UpsertUsageParams{
		ProjectID: acm.ID, Track: "ACM-1", MessageID: "m1", RequestID: "req_m1", Ts: testUsageTs,
		Model: "claude-sonnet-5", InputTokens: 2_000_000,
	})
	sqlDB.Close()

	if err != nil {
		t.Fatalf("UpsertUsage() returned error: %v", err)
	}

	t.Chdir(t.TempDir())

	stores, err := project.OpenAll(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	cost := costFor(t.Context(), stores)

	t.Run("builds the track cost by its code from an unrelated cwd", func(t *testing.T) {
		got, err := cost("ACM-1")
		if err != nil {
			t.Fatalf("cost(ACM-1) returned error: %v", err)
		}

		for _, want := range []string{"| Total cost |", "_as of 2026-10-22_"} {
			if !strings.Contains(got, want) {
				t.Errorf("cost(ACM-1) = %q, want it to contain %q", got, want)
			}
		}
	})

	t.Run("returns empty for a track without usage", func(t *testing.T) {
		got, err := cost(testOwnTrack)
		if err != nil {
			t.Fatalf("cost(%s) returned error: %v", testOwnTrack, err)
		}

		if got != "" {
			t.Errorf("cost(%s) = %q, want empty", testOwnTrack, got)
		}
	})

	t.Run("errors for a code matching no project", func(t *testing.T) {
		got, err := cost("ZZZ-1")
		if err == nil {
			t.Error("cost(ZZZ-1) error = nil, want non-nil")
		}

		if got != "" {
			t.Errorf("cost(ZZZ-1) = %q, want empty", got)
		}
	})
}
