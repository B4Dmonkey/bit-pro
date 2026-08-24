package cmd

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/B4Dmonkey/bit-pro/task"
	"github.com/modelcontextprotocol/go-sdk/mcp"
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

	testFirstBarTitle    = "Store.Create owns ID minting"
	testCreatePhase      = 1
	testCreatePhaseLabel = "Scope writes"

	testThirdBarID       = "FOO-1.3"
	testThirdBarTitle    = "a third bar"
	testFourthBarID      = "FOO-1.4"
	testInsertedBarTitle = "Contradiction forces the pointer patch"

	testFeedbackDir    = "feedback"
	testTrackKey       = "track"
	testPathKey        = "path"
	testUnknownTrackID = "FOO-9"
	testNoteBody       = "## What the plan said\n\nThe handler validates the anchor pair.\n\n" +
		"## What the work required\n\nThe rule belongs in the store."
	testSecondNoteBody = "## What the plan said\n\nA second correction."

	testCompletedDir = "completed"
	testTasksDir     = "tasks"

	testAfterKey  = "after"
	testBarKey    = "bar"
	testBeforeKey = "before"
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
		testBodyKey:  testCreateBody,
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
		"id":        testTrackID,
		testBodyKey: testUpdateBody,
	})

	if got["id"] != testTrackID {
		t.Errorf("id = %v, want %s", got["id"], testTrackID)
	}

	if got[testApprovedKey] != false {
		t.Errorf("approved = %v, want false", got[testApprovedKey])
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

func TestServeMCPCmd_TaskCreateMintsABarUnderATrack(t *testing.T) {
	dir := t.TempDir()
	seedConfig(t, dir)
	seedTasks(t, dir, &task.Task{ID: testTrackID, Title: testTitle, Status: task.StatusTodo})

	session := mcpSession(t, dir)

	first := callTool(t, session, taskCreateTool, map[string]any{
		testTitleKey:      testFirstBarTitle,
		testParentKey:     testTrackID,
		testPhaseKey:      testCreatePhase,
		testPhaseLabelKey: testCreatePhaseLabel,
		testBodyKey:       testSeedBarBody,
	})

	if first["id"] != testBarID {
		t.Fatalf("id = %v, want %s", first["id"], testBarID)
	}

	second := callTool(t, session, taskCreateTool, map[string]any{
		testTitleKey:      testSecondBarTitle,
		testParentKey:     testTrackID,
		testPhaseKey:      testCreatePhase,
		testPhaseLabelKey: testCreatePhaseLabel,
	})

	if second["id"] != testSecondBarID {
		t.Fatalf("id = %v, want %s", second["id"], testSecondBarID)
	}

	store := task.New(filepath.Join(dir, ".bit"))

	bar, err := store.Load(testBarID)
	if err != nil {
		t.Fatal(err)
	}

	want := task.Task{
		ID: testBarID, Title: testFirstBarTitle, Status: task.StatusTodo,
		Phase: testCreatePhase, PhaseLabel: testCreatePhaseLabel, Body: testSeedBarBody,
	}

	if !reflect.DeepEqual(*bar, want) {
		t.Errorf("task = %+v, want %+v", *bar, want)
	}

	bodyless, err := store.Load(testSecondBarID)
	if err != nil {
		t.Fatal(err)
	}

	if bodyless.Body != "" {
		t.Errorf("Body = %q, want empty", bodyless.Body)
	}

	bars := callToolList(t, session, taskListTool, map[string]any{testParentKey: testTrackID})

	wantBars := []map[string]any{
		{
			"id": testBarID, testTitleKey: testFirstBarTitle, testStatusKey: task.StatusTodo,
			testApprovedKey: false, testPhaseKey: float64(testCreatePhase),
			testPhaseLabelKey: testCreatePhaseLabel, testParentKey: testTrackID,
		},
		{
			"id": testSecondBarID, testTitleKey: testSecondBarTitle, testStatusKey: task.StatusTodo,
			testApprovedKey: false, testPhaseKey: float64(testCreatePhase),
			testPhaseLabelKey: testCreatePhaseLabel, testParentKey: testTrackID,
		},
	}

	if !reflect.DeepEqual(bars, wantBars) {
		t.Errorf("bars = %+v, want %+v", bars, wantBars)
	}
}

func TestServeMCPCmd_TaskCreateAfterPlacesABarMidTrack(t *testing.T) {
	dir := t.TempDir()
	seedConfig(t, dir)
	seedTasks(t, dir,
		&task.Task{
			ID: testTrackID, Title: testTitle, Status: task.StatusTodo,
			Order: []string{testBarID, testSecondBarID, testThirdBarID},
		},
		&task.Task{ID: testBarID, Title: testBarTitle, Status: task.StatusTodo},
		&task.Task{ID: testSecondBarID, Title: testSecondBarTitle, Status: task.StatusTodo},
		&task.Task{ID: testThirdBarID, Title: testThirdBarTitle, Status: task.StatusTodo},
	)

	session := mcpSession(t, dir)

	got := callTool(t, session, taskCreateTool, map[string]any{
		testTitleKey:      testInsertedBarTitle,
		testParentKey:     testTrackID,
		testAfterKey:      testBarID,
		testPhaseKey:      testCreatePhase,
		testPhaseLabelKey: testCreatePhaseLabel,
	})

	if got["id"] != testFourthBarID {
		t.Fatalf("id = %v, want %s", got["id"], testFourthBarID)
	}

	track, err := task.New(filepath.Join(dir, ".bit")).Load(testTrackID)
	if err != nil {
		t.Fatal(err)
	}

	wantOrder := []string{testBarID, testFourthBarID, testSecondBarID, testThirdBarID}
	if !reflect.DeepEqual(track.Order, wantOrder) {
		t.Fatalf("Order = %v, want %v", track.Order, wantOrder)
	}

	bars := callToolList(t, session, taskListTool, map[string]any{testParentKey: testTrackID})

	gotIDs := make([]string, len(bars))
	for i, bar := range bars {
		id, _ := bar["id"].(string)
		gotIDs[i] = id
	}

	if !reflect.DeepEqual(gotIDs, wantOrder) {
		t.Errorf("listed IDs = %v, want %v", gotIDs, wantOrder)
	}
}

func TestServeMCPCmd_TaskMoveResequencesABar(t *testing.T) {
	tests := []struct {
		name      string
		args      map[string]any
		wantOrder []string
	}{
		{
			name:      "before moves a bar to the front",
			args:      map[string]any{testBarKey: testThirdBarID, testBeforeKey: testBarID},
			wantOrder: []string{testThirdBarID, testBarID, testSecondBarID},
		},
		{
			name:      "after moves a bar to the back",
			args:      map[string]any{testBarKey: testBarID, testAfterKey: testThirdBarID},
			wantOrder: []string{testSecondBarID, testThirdBarID, testBarID},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			seedOrderedTrack(t, dir)

			session := mcpSession(t, dir)

			got := callTool(t, session, taskMoveTool, tt.args)
			if len(got) != 0 {
				t.Errorf("structured content = %v, want empty", got)
			}

			track, err := task.New(filepath.Join(dir, ".bit")).Load(testTrackID)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(track.Order, tt.wantOrder) {
				t.Fatalf("Order = %v, want %v", track.Order, tt.wantOrder)
			}

			if gotIDs := listedBarIDs(t, session); !reflect.DeepEqual(gotIDs, tt.wantOrder) {
				t.Errorf("listed IDs = %v, want %v", gotIDs, tt.wantOrder)
			}
		})
	}
}

func TestServeMCPCmd_TaskMoveRefusesABadAnchorPair(t *testing.T) {
	tests := []struct {
		name string
		args map[string]any
	}{
		{
			name: "both anchors is refused",
			args: map[string]any{testBarKey: testThirdBarID, testBeforeKey: testBarID, testAfterKey: testSecondBarID},
		},
		{
			name: "neither anchor is refused",
			args: map[string]any{testBarKey: testThirdBarID},
		},
	}

	wantOrder := []string{testBarID, testSecondBarID, testThirdBarID}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			seedOrderedTrack(t, dir)

			result := callToolResult(t, mcpSession(t, dir), taskMoveTool, tt.args)
			if !result.IsError {
				t.Fatalf("IsError = false, want true (content %v)", result.Content)
			}

			track, err := task.New(filepath.Join(dir, ".bit")).Load(testTrackID)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(track.Order, wantOrder) {
				t.Errorf("Order = %v, want %v", track.Order, wantOrder)
			}
		})
	}
}

func seedOrderedTrack(t *testing.T, dir string) {
	t.Helper()

	seedTasks(t, dir,
		&task.Task{
			ID: testTrackID, Title: testTitle, Status: task.StatusTodo,
			Order: []string{testBarID, testSecondBarID, testThirdBarID},
		},
		&task.Task{ID: testBarID, Title: testBarTitle, Status: task.StatusTodo},
		&task.Task{ID: testSecondBarID, Title: testSecondBarTitle, Status: task.StatusTodo},
		&task.Task{ID: testThirdBarID, Title: testThirdBarTitle, Status: task.StatusTodo},
	)
}

func listedBarIDs(t *testing.T, session *mcp.ClientSession) []string {
	t.Helper()

	bars := callToolList(t, session, taskListTool, map[string]any{testParentKey: testTrackID})

	ids := make([]string, len(bars))
	for i, bar := range bars {
		id, _ := bar["id"].(string)
		ids[i] = id
	}

	return ids
}

func TestServeMCPCmd_FeedbackAddWritesANote(t *testing.T) {
	dir := t.TempDir()
	seedTasks(t, dir, &task.Task{ID: testTrackID, Title: testTitle, Status: task.StatusDoing})

	session := mcpSession(t, dir)

	first := callTool(t, session, feedbackAddTool, map[string]any{
		testTrackKey: testTrackID,
		testBodyKey:  testNoteBody,
	})

	wantFirst := filepath.Join(testFeedbackDir, testTrackID+"-001.md")
	if gotPath, ok := first[testPathKey].(string); !ok || !strings.HasSuffix(gotPath, wantFirst) {
		t.Fatalf("path = %v, want suffix %s", first[testPathKey], wantFirst)
	}

	second := callTool(t, session, feedbackAddTool, map[string]any{
		testTrackKey: testTrackID,
		testBodyKey:  testSecondNoteBody,
	})

	wantSecond := filepath.Join(testFeedbackDir, testTrackID+"-002.md")
	if gotPath, ok := second[testPathKey].(string); !ok || !strings.HasSuffix(gotPath, wantSecond) {
		t.Fatalf("path = %v, want suffix %s", second[testPathKey], wantSecond)
	}

	written, err := os.ReadFile(first[testPathKey].(string))
	if err != nil {
		t.Fatal(err)
	}

	if string(written) != testNoteBody {
		t.Errorf("note body = %q, want %q", string(written), testNoteBody)
	}
}

func TestServeMCPCmd_FeedbackAddRefusesAnUnknownTrack(t *testing.T) {
	dir := t.TempDir()
	seedTasks(t, dir, &task.Task{ID: testTrackID, Title: testTitle, Status: task.StatusDoing})

	result := callToolResult(t, mcpSession(t, dir), feedbackAddTool, map[string]any{
		testTrackKey: testUnknownTrackID,
		testBodyKey:  testNoteBody,
	})

	if !result.IsError {
		t.Fatalf("IsError = false, want true (content %v)", result.Content)
	}

	notes, err := filepath.Glob(filepath.Join(dir, ".bit", testFeedbackDir, "*.md"))
	if err != nil {
		t.Fatal(err)
	}

	if len(notes) != 0 {
		t.Errorf("notes = %v, want none", notes)
	}
}

func TestServeMCPCmd_TaskCompleteFilesATrackAndItsBars(t *testing.T) {
	dir := t.TempDir()
	seedDoneTrack(t, dir, task.StatusDone)

	session := mcpSession(t, dir)

	got := callTool(t, session, taskCompleteTool, map[string]any{"id": testTrackID})
	if len(got) != 0 {
		t.Errorf("structured content = %v, want empty", got)
	}

	for _, id := range []string{testTrackID, testBarID, testSecondBarID} {
		assertRelocated(t, dir, id, testCompletedDir)
	}

	if listed := callToolList(t, session, taskListTool, map[string]any{}); len(listed) != 0 {
		t.Errorf("listed tasks = %v, want none", listed)
	}
}

func TestServeMCPCmd_TaskCompleteRefusesUnfinishedBars(t *testing.T) {
	dir := t.TempDir()
	seedDoneTrack(t, dir, task.StatusDoing)

	result := callToolResult(t, mcpSession(t, dir), taskCompleteTool, map[string]any{"id": testTrackID})
	if !result.IsError {
		t.Fatalf("IsError = false, want true (content %v)", result.Content)
	}

	for _, id := range []string{testTrackID, testBarID, testSecondBarID} {
		if _, err := os.Stat(filepath.Join(dir, ".bit", testTasksDir, id+".md")); err != nil {
			t.Errorf("%s missing from %s: %v", id, testTasksDir, err)
		}
	}

	completed, err := filepath.Glob(filepath.Join(dir, ".bit", testCompletedDir, "*.md"))
	if err != nil {
		t.Fatal(err)
	}

	if len(completed) != 0 {
		t.Errorf("completed = %v, want none", completed)
	}
}

func seedDoneTrack(t *testing.T, dir, lastBarStatus string) {
	t.Helper()

	seedTasks(t, dir,
		&task.Task{
			ID: testTrackID, Title: testTitle, Status: task.StatusDone,
			Order: []string{testBarID, testSecondBarID},
		},
		&task.Task{ID: testBarID, Title: testBarTitle, Status: task.StatusDone},
		&task.Task{ID: testSecondBarID, Title: testSecondBarTitle, Status: lastBarStatus},
	)
}

func assertRelocated(t *testing.T, dir, id, into string) {
	t.Helper()

	if _, err := os.Stat(filepath.Join(dir, ".bit", into, id+".md")); err != nil {
		t.Errorf("%s missing from %s: %v", id, into, err)
	}

	if _, err := os.Stat(filepath.Join(dir, ".bit", testTasksDir, id+".md")); !os.IsNotExist(err) {
		t.Errorf("%s still under %s (err %v)", id, testTasksDir, err)
	}
}
