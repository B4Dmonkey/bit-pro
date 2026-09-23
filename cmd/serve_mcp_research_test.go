package cmd

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/B4Dmonkey/bit-pro/task"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	testResearchDir  = "research"
	testTopicKey     = "topic"
	testIndexTopic   = "index"
	testResearchBody = "## Findings\n\n- note paths are built with pathologize.Join in task/feedback.go\n" +
		"- see [store](store.md)"
	testRewrittenBody = "## Findings\n\n- rewritten after a re-check"
	testUnsafeTopic   = "../../tasks/" + testTrackID
	testEmptyTopicErr = "is empty"
	testDotDotErr     = `contains ".."`
	testSeparatorErr  = "contains a path separator"
	testStoreTopic    = "store"
	testStoreBody     = "Store.Load normalizes IDs before reading."
	testMissingTopic  = "nope"
	testMCPTopic      = "mcp"
	testTopicsKey     = "topics"
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

func TestServeMCPCmd_ResearchReadReturnsATopic(t *testing.T) {
	dir := t.TempDir()
	seedTasks(t, dir, &task.Task{ID: testTrackID, Title: testTitle, Status: task.StatusDoing})

	session := mcpSession(t, dir)

	callTool(t, session, researchWriteTool, map[string]any{
		testTrackKey: testTrackID,
		testTopicKey: testStoreTopic,
		testBodyKey:  testStoreBody,
	})

	got := callTool(t, session, researchReadTool, map[string]any{
		testTrackKey: testTrackID,
		testTopicKey: testStoreTopic,
	})

	if got[testBodyKey] != testStoreBody {
		t.Errorf("body = %v, want %q", got[testBodyKey], testStoreBody)
	}
}

func TestServeMCPCmd_ResearchReadRefusesAMissingTopic(t *testing.T) {
	dir := t.TempDir()
	seedTasks(t, dir, &task.Task{ID: testTrackID, Title: testTitle, Status: task.StatusDoing})

	result := callToolResult(t, mcpSession(t, dir), researchReadTool, map[string]any{
		testTrackKey: testTrackID,
		testTopicKey: testMissingTopic,
	})

	if !result.IsError {
		t.Errorf("IsError = false, want true (content %v)", result.Content)
	}
}

func TestServeMCPCmd_ResearchReadRefusesAnUnknownTrack(t *testing.T) {
	dir := t.TempDir()
	seedTasks(t, dir, &task.Task{ID: testTrackID, Title: testTitle, Status: task.StatusDoing})

	result := callToolResult(t, mcpSession(t, dir), researchReadTool, map[string]any{
		testTrackKey: testUnknownTrackID,
		testTopicKey: testIndexTopic,
	})

	if !result.IsError {
		t.Errorf("IsError = false, want true (content %v)", result.Content)
	}
}

func TestServeMCPCmd_ResearchReadListsTopics(t *testing.T) {
	dir := t.TempDir()
	seedTasks(t, dir, &task.Task{ID: testTrackID, Title: testTitle, Status: task.StatusDoing})

	session := mcpSession(t, dir)

	for _, topic := range []string{testIndexTopic, testStoreTopic, testMCPTopic} {
		callTool(t, session, researchWriteTool, map[string]any{
			testTrackKey: testTrackID,
			testTopicKey: topic,
			testBodyKey:  testResearchBody,
		})
	}

	got := callTool(t, session, researchReadTool, map[string]any{testTrackKey: testTrackID})

	want := []any{testIndexTopic, testMCPTopic, testStoreTopic}
	if !reflect.DeepEqual(got[testTopicsKey], want) {
		t.Errorf("topics = %v, want %v", got[testTopicsKey], want)
	}

	if body, _ := got[testBodyKey].(string); body != "" {
		t.Errorf("body = %q, want empty", body)
	}
}

func TestServeMCPCmd_ResearchReadListsNothingForATrackWithNoResearch(t *testing.T) {
	dir := t.TempDir()
	seedTasks(t, dir, &task.Task{ID: testTrackID, Title: testTitle, Status: task.StatusDoing})

	result := callToolResult(t, mcpSession(t, dir), researchReadTool, map[string]any{testTrackKey: testTrackID})

	if result.IsError {
		t.Fatalf("IsError = true, want false (content %v)", result.Content)
	}

	got, _ := result.StructuredContent.(map[string]any)
	if topics, _ := got[testTopicsKey].([]any); len(topics) != 0 {
		t.Errorf("topics = %v, want none", topics)
	}
}

type pathLikeTopic struct {
	name    string
	topic   string
	wantErr string
}

func pathLikeTopics() []pathLikeTopic {
	return []pathLikeTopic{
		{name: "traversal", topic: testUnsafeTopic, wantErr: testDotDotErr},
		{name: "embedded dot dot", topic: "a..b", wantErr: testDotDotErr},
		{name: "slash", topic: "notes/store", wantErr: testSeparatorErr},
		{name: "backslash", topic: `notes\store`, wantErr: testSeparatorErr},
		{name: "dots only", topic: "...", wantErr: testEmptyTopicErr},
		{name: "dots and spaces", topic: ". .", wantErr: testEmptyTopicErr},
	}
}

func assertToolErrorNames(t *testing.T, result *mcp.CallToolResult, want string) {
	t.Helper()

	if !result.IsError {
		t.Fatalf("IsError = false, want true (content %v)", result.Content)
	}

	text, ok := result.Content[0].(*mcp.TextContent)
	if !ok || !strings.Contains(text.Text, want) {
		t.Errorf("error = %v, want it to contain %q", result.Content[0], want)
	}
}

func TestServeMCPCmd_ResearchWriteRefusesAPathLikeTopic(t *testing.T) {
	topics := append(pathLikeTopics(), pathLikeTopic{name: "empty", topic: "", wantErr: testEmptyTopicErr})

	for _, tt := range topics {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			seedTasks(t, dir, &task.Task{ID: testTrackID, Title: testTitle, Status: task.StatusDoing})

			result := callToolResult(t, mcpSession(t, dir), researchWriteTool, map[string]any{
				testTrackKey: testTrackID,
				testTopicKey: tt.topic,
				testBodyKey:  testResearchBody,
			})

			assertToolErrorNames(t, result, tt.wantErr)

			_, err := os.Stat(filepath.Join(dir, ".bit", testResearchDir))
			if !errors.Is(err, fs.ErrNotExist) {
				t.Errorf("research folder: stat err = %v, want ErrNotExist", err)
			}

			track, err := task.New(filepath.Join(dir, ".bit")).Load(testTrackID)
			if err != nil {
				t.Fatal(err)
			}

			if track.Title != testTitle {
				t.Errorf("track title = %q, want %q", track.Title, testTitle)
			}
		})
	}
}

func TestServeMCPCmd_ResearchReadRefusesAPathLikeTopic(t *testing.T) {
	for _, tt := range pathLikeTopics() {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			seedTasks(t, dir, &task.Task{ID: testTrackID, Title: testTitle, Status: task.StatusDoing})

			result := callToolResult(t, mcpSession(t, dir), researchReadTool, map[string]any{
				testTrackKey: testTrackID,
				testTopicKey: tt.topic,
			})

			assertToolErrorNames(t, result, tt.wantErr)
		})
	}
}
