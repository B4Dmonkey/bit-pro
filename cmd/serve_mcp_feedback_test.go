package cmd

import (
	"reflect"
	"testing"

	"github.com/B4Dmonkey/bit-pro/db/orm"
	"github.com/B4Dmonkey/bit-pro/project"
	"github.com/B4Dmonkey/bit-pro/task"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	testNotesKey    = "notes"
	testOwnCode     = "BIT"
	testOtherCode   = "EX"
	testOwnTrack    = "BIT-1"
	testOwnTrack2   = "BIT-2"
	testOtherTrack  = "EX-1"
	testMissingOwn  = "BIT-9"
	testLowerTrack  = "bit-1"
	testOwnNote1    = "BIT-1-001"
	testOwnNote2    = "BIT-1-002"
	testOwnNoteBIT2 = "BIT-2-001"
	testOtherNote   = "EX-1-001"
	testNoteIDKey   = "id"
)

func seedCodedProject(t *testing.T, dir, code string, tracks ...string) {
	t.Helper()

	mcpSandbox(t)

	path, err := project.CanonicalPath(dir)
	if err != nil {
		t.Fatalf("CanonicalPath(%q) returned error: %v", dir, err)
	}

	seedProject(t, orm.CreateProjectParams{Path: path, Code: code})

	if _, err := project.OpenStore(t.Context(), dir); err != nil {
		t.Fatalf("project.OpenStore(%q) returned error: %v", dir, err)
	}

	for _, id := range tracks {
		seedTasks(t, dir, &task.Task{ID: id, Title: testTitle, Status: task.StatusDoing})
	}
}

func addNote(t *testing.T, s *mcp.ClientSession, track string) {
	t.Helper()

	callTool(t, s, feedbackAddTool, map[string]any{testTrackKey: track, testBodyKey: testNoteBody})
}

func seedTwoProjectsNotes(t *testing.T) *mcp.ClientSession {
	t.Helper()

	dir := t.TempDir()
	other := t.TempDir()

	seedCodedProject(t, dir, testOwnCode, testOwnTrack, testOwnTrack2)
	seedCodedProject(t, other, testOtherCode, testOtherTrack)

	session := mcpSession(t, dir)
	addNote(t, session, testOwnTrack)
	addNote(t, session, testOwnTrack)
	addNote(t, session, testOwnTrack2)
	addNote(t, mcpSession(t, other), testOtherTrack)

	return session
}

func listNotes(t *testing.T, s *mcp.ClientSession, args map[string]any) []string {
	t.Helper()

	var got struct {
		Notes []string `json:"notes"`
	}

	decodeToolResult(t, s, feedbackListTool, args, &got)

	return got.Notes
}

func TestFeedbackListHandler(t *testing.T) {
	t.Run("lists only this project's notes", func(t *testing.T) {
		got := listNotes(t, seedTwoProjectsNotes(t), map[string]any{})

		want := []string{testOwnNote1, testOwnNote2, testOwnNoteBIT2}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("notes = %v, want %v", got, want)
		}
	})

	t.Run("filters to one track", func(t *testing.T) {
		got := listNotes(t, seedTwoProjectsNotes(t), map[string]any{testTrackKey: testLowerTrack})

		want := []string{testOwnNote1, testOwnNote2}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("notes = %v, want %v", got, want)
		}
	})

	t.Run("a project with no notes lists none", func(t *testing.T) {
		dir := t.TempDir()
		seedCodedProject(t, dir, testOwnCode, testOwnTrack)

		got := callTool(t, mcpSession(t, dir), feedbackListTool, map[string]any{})

		notes, ok := got[testNotesKey].([]any)
		if !ok || len(notes) != 0 {
			t.Errorf("notes = %#v, want []", got[testNotesKey])
		}
	})

	t.Run("refuses an unknown track", func(t *testing.T) {
		dir := t.TempDir()
		seedCodedProject(t, dir, testOwnCode, testOwnTrack)

		result := callToolResult(t, mcpSession(t, dir), feedbackListTool, map[string]any{testTrackKey: testMissingOwn})
		if !result.IsError {
			t.Fatalf("IsError = false, want true (content %v)", result.Content)
		}
	})
}

func TestFeedbackReadHandler(t *testing.T) {
	t.Run("returns the note body", func(t *testing.T) {
		dir := t.TempDir()
		seedCodedProject(t, dir, testOwnCode, testOwnTrack)

		session := mcpSession(t, dir)
		addNote(t, session, testOwnTrack)

		got := callTool(t, session, feedbackReadTool, map[string]any{testNoteIDKey: testOwnNote1})
		if got[testBodyKey] != testNoteBody {
			t.Errorf("body = %q, want %q", got[testBodyKey], testNoteBody)
		}
	})

	t.Run("refuses another project's note", func(t *testing.T) {
		result := callToolResult(t, seedTwoProjectsNotes(t), feedbackReadTool, map[string]any{testNoteIDKey: testOtherNote})

		assertToolErrorNames(t, result, "another project")
	})

	t.Run("refuses an unknown or path-like id", func(t *testing.T) {
		tests := []struct {
			name string
			id   string
		}{
			{name: "unknown", id: "BIT-1-099"},
			{name: "path-like", id: "../" + testOwnNote1},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := callToolResult(t, seedTwoProjectsNotes(t), feedbackReadTool, map[string]any{testNoteIDKey: tt.id})
				if !result.IsError {
					t.Fatalf("IsError = false, want true (content %v)", result.Content)
				}
			})
		}
	})
}
