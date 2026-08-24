package cmd

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/B4Dmonkey/bit-pro/task"
)

const (
	testCreateBody = "## Why\n\nBecause `sed` owns the file format today.\n\n## Summary\n\nSix tools."
	testUpdateBody = "## Why\n\nnew reason\n\n## Decisions\n\n- settled"

	testSeedBarTitle       = "Contradiction forces real fan-out"
	testSeedBarBody        = "## **Verse 2**\n\nstep detail"
	testSeedPhase          = 2
	testSeedPhaseLabel     = "Plan writes"
	testRenamedBarTitle    = "Renamed step"
	testRetaggedPhase      = 3
	testRetaggedPhaseLabel = "Run writes"

	testMisspelledStatus = "doen"
)

func seedConfig(t *testing.T, dir string) {
	t.Helper()

	if err := task.New(filepath.Join(dir, ".bit")).SaveConfig(&task.Config{Prefix: testCode}); err != nil {
		t.Fatal(err)
	}
}

func TestServeMCPCmd_TaskCreateMintsATrack(t *testing.T) {
	dir := t.TempDir()
	seedConfig(t, dir)

	got := callTool(t, mcpSession(t, dir), taskCreateTool, map[string]any{
		testTitleKey: testTitle,
		"body":       testCreateBody,
	})

	if got["id"] != testTrackID {
		t.Fatalf("id = %v, want %s", got["id"], testTrackID)
	}

	created, err := task.New(filepath.Join(dir, ".bit")).Load(testTrackID)
	if err != nil {
		t.Fatal(err)
	}

	if created.Title != testTitle {
		t.Errorf("Title = %q, want %q", created.Title, testTitle)
	}

	if created.Status != task.StatusTodo {
		t.Errorf("Status = %q, want %q", created.Status, task.StatusTodo)
	}

	if created.Body != testCreateBody {
		t.Errorf("Body = %q, want %q", created.Body, testCreateBody)
	}
}

func TestServeMCPCmd_TaskUpdateRewritesBodyAndReportsRevocation(t *testing.T) {
	dir := t.TempDir()
	seedTasks(t, dir, &task.Task{
		ID: testTrackID, Title: testTitle, Status: task.StatusTodo, Approved: true, Body: "## Why\n\nold reason",
	})

	got := callTool(t, mcpSession(t, dir), taskUpdateTool, map[string]any{
		"id":   testTrackID,
		"body": testUpdateBody,
	})

	if got["id"] != testTrackID {
		t.Errorf("id = %v, want %s", got["id"], testTrackID)
	}

	if got["approved"] != false {
		t.Errorf("approved = %v, want false", got["approved"])
	}

	updated, err := task.New(filepath.Join(dir, ".bit")).Load(testTrackID)
	if err != nil {
		t.Fatal(err)
	}

	if updated.Body != testUpdateBody {
		t.Errorf("Body = %q, want %q", updated.Body, testUpdateBody)
	}

	if updated.Approved {
		t.Error("Approved = true, want false")
	}
}

func TestServeMCPCmd_TaskUpdateLeavesOmittedFieldsAlone(t *testing.T) {
	seed := task.Task{
		ID:         testBarID,
		Title:      testSeedBarTitle,
		Status:     task.StatusTodo,
		Phase:      testSeedPhase,
		PhaseLabel: testSeedPhaseLabel,
		Body:       testSeedBarBody,
	}

	tests := []struct {
		name string
		args map[string]any
		want task.Task
	}{
		{
			name: "status only leaves title, body and phase alone",
			args: map[string]any{"id": testBarID, testStatusKey: task.StatusDoing},
			want: task.Task{
				ID: testBarID, Title: testSeedBarTitle, Status: task.StatusDoing,
				Phase: testSeedPhase, PhaseLabel: testSeedPhaseLabel, Body: testSeedBarBody,
			},
		},
		{
			name: "title only leaves body and status alone",
			args: map[string]any{"id": testBarID, testTitleKey: testRenamedBarTitle},
			want: task.Task{
				ID: testBarID, Title: testRenamedBarTitle, Status: task.StatusTodo,
				Phase: testSeedPhase, PhaseLabel: testSeedPhaseLabel, Body: testSeedBarBody,
			},
		},
		{
			name: "phase tag only leaves title and body alone",
			args: map[string]any{"id": testBarID, testPhaseKey: testRetaggedPhase, testPhaseLabelKey: testRetaggedPhaseLabel},
			want: task.Task{
				ID: testBarID, Title: testSeedBarTitle, Status: task.StatusTodo,
				Phase: testRetaggedPhase, PhaseLabel: testRetaggedPhaseLabel, Body: testSeedBarBody,
			},
		},
		{
			name: "id alone is a no-op",
			args: map[string]any{"id": testBarID},
			want: seed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			seeded := seed
			seedTasks(t, dir, &seeded)

			callTool(t, mcpSession(t, dir), taskUpdateTool, tt.args)

			got, err := task.New(filepath.Join(dir, ".bit")).Load(testBarID)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(*got, tt.want) {
				t.Errorf("task = %+v, want %+v", *got, tt.want)
			}
		})
	}
}

func TestServeMCPCmd_TaskUpdateRefusesAnUnknownStatus(t *testing.T) {
	tests := []struct {
		name    string
		status  string
		wantErr bool
	}{
		{name: "a misspelled status is refused", status: testMisspelledStatus, wantErr: true},
		{name: "todo is accepted", status: task.StatusTodo},
		{name: "doing is accepted", status: task.StatusDoing},
		{name: "done is accepted", status: task.StatusDone},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			seedTasks(t, dir, &task.Task{ID: testTrackID, Title: testTitle, Status: task.StatusTodo})

			result := callToolResult(t, mcpSession(t, dir), taskUpdateTool, map[string]any{
				"id": testTrackID, testStatusKey: tt.status,
			})

			if result.IsError != tt.wantErr {
				t.Fatalf("IsError = %v, want %v (content %v)", result.IsError, tt.wantErr, result.Content)
			}

			got, err := task.New(filepath.Join(dir, ".bit")).Load(testTrackID)
			if err != nil {
				t.Fatal(err)
			}

			want := tt.status
			if tt.wantErr {
				want = task.StatusTodo
			}

			if got.Status != want {
				t.Errorf("Status = %q, want %q", got.Status, want)
			}
		})
	}
}
