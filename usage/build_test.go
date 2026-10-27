package usage

import (
	"strings"
	"testing"

	"github.com/B4Dmonkey/bit-pro/db"
	"github.com/B4Dmonkey/bit-pro/db/orm"
	"github.com/B4Dmonkey/bit-pro/task"
)

const (
	buildTrack = "BIT-1"
	buildTs    = "2026-10-22T10:00:00Z"
)

func buildQueries(t *testing.T) (*orm.Queries, int64) {
	t.Helper()

	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_DATA_HOME", t.TempDir())

	sqlDB, err := db.Open()
	if err != nil {
		t.Fatalf("db.Open() returned error: %v", err)
	}

	t.Cleanup(func() { sqlDB.Close() })

	q := orm.New(sqlDB)

	if err := q.CreateProject(t.Context(), orm.CreateProjectParams{Path: t.TempDir(), Code: "BIT"}); err != nil {
		t.Fatalf("CreateProject() returned error: %v", err)
	}

	projects, err := q.ListProjects(t.Context())
	if err != nil || len(projects) != 1 {
		t.Fatalf("ListProjects() = %v, %v, want one project", projects, err)
	}

	return q, projects[0].ID
}

func seedBuildUsage(t *testing.T, q *orm.Queries, projectID int64, id, model string) {
	t.Helper()

	err := q.UpsertUsage(t.Context(), orm.UpsertUsageParams{
		ProjectID: projectID, Track: buildTrack, MessageID: id, RequestID: "req_" + id, Ts: buildTs,
		Model: model, InputTokens: 2_000_000, OutputTokens: 100_000,
	})
	if err != nil {
		t.Fatalf("UpsertUsage() returned error: %v", err)
	}
}

func TestBuild(t *testing.T) {
	q, projectID := buildQueries(t)
	seedBuildUsage(t, q, projectID, "m1", "claude-sonnet-5")
	seedBuildUsage(t, q, projectID, "m2", "claude-sonnet-5")

	md, rows, err := Build(t.Context(), q, projectID, buildTrack, task.StatusDoing, nil)
	if err != nil {
		t.Fatalf("Build() returned error: %v", err)
	}

	if rows != 2 {
		t.Errorf("rows = %d, want 2", rows)
	}

	for _, want := range []string{"_as of 2026-10-22_", "| Total cost |"} {
		if !strings.Contains(md, want) {
			t.Errorf("markdown = %q, want it to contain %q", md, want)
		}
	}
}

func TestBuild_NoUsage(t *testing.T) {
	q, projectID := buildQueries(t)

	_, rows, err := Build(t.Context(), q, projectID, "BIT-9", task.StatusDoing, nil)
	if err != nil {
		t.Fatalf("Build() returned error: %v", err)
	}

	if rows != 0 {
		t.Errorf("rows = %d, want 0", rows)
	}
}

func TestBuild_UnpricedModel(t *testing.T) {
	q, projectID := buildQueries(t)
	seedBuildUsage(t, q, projectID, "m1", "claude-nextgen-9")

	md, _, err := Build(t.Context(), q, projectID, buildTrack, task.StatusDoing, nil)
	if err != nil {
		t.Fatalf("Build() returned error: %v", err)
	}

	if want := "| Total cost | unknown |"; !strings.Contains(md, want) {
		t.Errorf("markdown = %q, want it to contain %q", md, want)
	}
}
