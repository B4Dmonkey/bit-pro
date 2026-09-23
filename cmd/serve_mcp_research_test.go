package cmd

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/B4Dmonkey/bit-pro/task"
)

const (
	testResearchDir  = "research"
	testTopicKey     = "topic"
	testIndexTopic   = "index"
	testResearchBody = "## Findings\n\n- note paths are built with pathologize.Join in task/feedback.go\n" +
		"- see [store](store.md)"
	testRewrittenBody = "## Findings\n\n- rewritten after a re-check"
	testUnsafeTopic   = "../../tasks/" + testTrackID
	testDotsOnlyTopic = "..."
)

func TestServeMCPCmd_ResearchWriteWritesATopic(t *testing.T) {
	dir := t.TempDir()
	seedTasks(t, dir, &task.Task{ID: testTrackID, Title: testTitle, Status: task.StatusDoing})

	got := callTool(t, mcpSession(t, dir), researchWriteTool, map[string]any{
		testTrackKey: testTrackID,
		testTopicKey: testIndexTopic,
		testBodyKey:  testResearchBody,
	})

	want := filepath.Join(testResearchDir, testTrackID, testIndexTopic+".md")

	gotPath, ok := got[testPathKey].(string)
	if !ok || !strings.HasSuffix(gotPath, want) {
		t.Fatalf("path = %v, want suffix %s", got[testPathKey], want)
	}

	written, err := os.ReadFile(gotPath)
	if err != nil {
		t.Fatal(err)
	}

	if string(written) != testResearchBody {
		t.Errorf("research body = %q, want %q", string(written), testResearchBody)
	}
}

func TestServeMCPCmd_ResearchWriteOverwritesATopic(t *testing.T) {
	dir := t.TempDir()
	seedTasks(t, dir, &task.Task{ID: testTrackID, Title: testTitle, Status: task.StatusDoing})

	session := mcpSession(t, dir)

	callTool(t, session, researchWriteTool, map[string]any{
		testTrackKey: testTrackID,
		testTopicKey: testIndexTopic,
		testBodyKey:  testResearchBody,
	})

	got := callTool(t, session, researchWriteTool, map[string]any{
		testTrackKey: testTrackID,
		testTopicKey: testIndexTopic,
		testBodyKey:  testRewrittenBody,
	})

	written, err := os.ReadFile(got[testPathKey].(string))
	if err != nil {
		t.Fatal(err)
	}

	if string(written) != testRewrittenBody {
		t.Errorf("research body = %q, want %q", string(written), testRewrittenBody)
	}

	topics, err := filepath.Glob(filepath.Join(dir, ".bit", testResearchDir, testTrackID, "*.md"))
	if err != nil {
		t.Fatal(err)
	}

	if len(topics) != 1 {
		t.Errorf("topics = %v, want exactly one", topics)
	}
}

func TestServeMCPCmd_ResearchWriteKeepsAnUnsafeTopicInTheTrackFolder(t *testing.T) {
	dir := t.TempDir()
	seedTasks(t, dir, &task.Task{ID: testTrackID, Title: testTitle, Status: task.StatusDoing})

	got := callTool(t, mcpSession(t, dir), researchWriteTool, map[string]any{
		testTrackKey: testTrackID,
		testTopicKey: testUnsafeTopic,
		testBodyKey:  testResearchBody,
	})

	rel, err := filepath.Rel(filepath.Join(dir, ".bit", testResearchDir, testTrackID), got[testPathKey].(string))
	if err != nil {
		t.Fatal(err)
	}

	if strings.HasPrefix(rel, "..") || strings.ContainsRune(rel, filepath.Separator) {
		t.Errorf("topic path %q escapes the track's research folder", rel)
	}

	track, err := task.New(filepath.Join(dir, ".bit")).Load(testTrackID)
	if err != nil {
		t.Fatal(err)
	}

	if track.Title != testTitle {
		t.Errorf("track title = %q, want %q", track.Title, testTitle)
	}
}

func TestServeMCPCmd_ResearchWriteRefusesAnUnknownTrack(t *testing.T) {
	dir := t.TempDir()
	seedTasks(t, dir, &task.Task{ID: testTrackID, Title: testTitle, Status: task.StatusDoing})

	result := callToolResult(t, mcpSession(t, dir), researchWriteTool, map[string]any{
		testTrackKey: testUnknownTrackID,
		testTopicKey: testIndexTopic,
		testBodyKey:  testResearchBody,
	})

	if !result.IsError {
		t.Fatalf("IsError = false, want true (content %v)", result.Content)
	}

	_, err := os.Stat(filepath.Join(dir, ".bit", testResearchDir, testUnknownTrackID))
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("research folder for %s: stat err = %v, want ErrNotExist", testUnknownTrackID, err)
	}
}

func TestServeMCPCmd_ResearchWriteAcceptsACompletedTrack(t *testing.T) {
	dir := t.TempDir()
	seedDoneTrack(t, dir, task.StatusDone)

	session := mcpSession(t, dir)
	callTool(t, session, taskCompleteTool, map[string]any{"id": testTrackID})

	got := callTool(t, session, researchWriteTool, map[string]any{
		testTrackKey: testTrackID,
		testTopicKey: testIndexTopic,
		testBodyKey:  testResearchBody,
	})

	want := filepath.Join(dir, ".bit", testResearchDir, testTrackID, testIndexTopic+".md")
	if got[testPathKey] != want {
		t.Errorf("path = %v, want %s", got[testPathKey], want)
	}

	if _, err := os.Stat(want); err != nil {
		t.Errorf("research topic missing: %v", err)
	}
}

func TestServeMCPCmd_ResearchWriteStripsLeadingDots(t *testing.T) {
	tests := []struct {
		name  string
		topic string
		want  string
	}{
		{name: "hidden", topic: ".hidden", want: "hidden.md"},
		{name: "traversal", topic: testUnsafeTopic, want: "tasks" + testTrackID + ".md"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			seedTasks(t, dir, &task.Task{ID: testTrackID, Title: testTitle, Status: task.StatusDoing})

			got := callTool(t, mcpSession(t, dir), researchWriteTool, map[string]any{
				testTrackKey: testTrackID,
				testTopicKey: tt.topic,
				testBodyKey:  testResearchBody,
			})

			want := filepath.Join(dir, ".bit", testResearchDir, testTrackID, tt.want)
			if got[testPathKey] != want {
				t.Errorf("path = %v, want %s", got[testPathKey], want)
			}
		})
	}
}

func TestServeMCPCmd_ResearchWriteRefusesADotsOnlyTopic(t *testing.T) {
	dir := t.TempDir()
	seedTasks(t, dir, &task.Task{ID: testTrackID, Title: testTitle, Status: task.StatusDoing})

	result := callToolResult(t, mcpSession(t, dir), researchWriteTool, map[string]any{
		testTrackKey: testTrackID,
		testTopicKey: testDotsOnlyTopic,
		testBodyKey:  testResearchBody,
	})

	if !result.IsError {
		t.Fatalf("IsError = false, want true (content %v)", result.Content)
	}

	topics, err := filepath.Glob(filepath.Join(dir, ".bit", testResearchDir, testTrackID, "*"))
	if err != nil {
		t.Fatal(err)
	}

	if len(topics) != 0 {
		t.Errorf("topics = %v, want none", topics)
	}
}
