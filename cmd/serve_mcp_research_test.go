package cmd

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

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
	testEscapingTrack = "../../" + testTrackID

	testHeadSHA         = "9f3c2b7e4d1a0c8b6e5f4a3b2c1d0e9f8a7b6c5d"
	testNextHeadSHA     = "b1e2d3c4b5a6978877665544332211009f8e7d6c"
	testBranch          = "v2"
	testRevParseArgs    = "rev-parse HEAD"
	testSymbolicRefArgs = "symbolic-ref --short -q HEAD"
	testCommitsKey      = "commits"
	testSHAKey          = "sha"
	testAtKey           = "at"
)

func TestResearchWriteHandler(t *testing.T) {
	t.Run("writes a topic", func(t *testing.T) {
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
	})

	t.Run("overwrites a topic", func(t *testing.T) {
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

		topics, err := filepath.Glob(filepath.Join(projectStoreDir(t, dir), testResearchDir, testTrackID, "*.md"))
		if err != nil {
			t.Fatal(err)
		}

		if len(topics) != 1 {
			t.Errorf("topics = %v, want exactly one", topics)
		}
	})

	t.Run("refuses an unknown track", func(t *testing.T) {
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

		_, err := os.Stat(filepath.Join(projectStoreDir(t, dir), testResearchDir, testUnknownTrackID))
		if !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("research folder for %s: stat err = %v, want ErrNotExist", testUnknownTrackID, err)
		}
	})

	t.Run("accepts a completed track", func(t *testing.T) {
		dir := t.TempDir()
		seedDoneTrack(t, dir, task.StatusDone)

		session := mcpSession(t, dir)
		callTool(t, session, taskCompleteTool, map[string]any{"id": testTrackID})

		got := callTool(t, session, researchWriteTool, map[string]any{
			testTrackKey: testTrackID,
			testTopicKey: testIndexTopic,
			testBodyKey:  testResearchBody,
		})

		want := filepath.Join(projectStoreDir(t, dir), testResearchDir, testTrackID, testIndexTopic+".md")
		if got[testPathKey] != want {
			t.Errorf("path = %v, want %s", got[testPathKey], want)
		}

		if _, err := os.Stat(want); err != nil {
			t.Errorf("research topic missing: %v", err)
		}
	})

	t.Run("strips leading dots", func(t *testing.T) {
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

				want := filepath.Join(projectStoreDir(t, dir), testResearchDir, testTrackID, tt.want)
				if got[testPathKey] != want {
					t.Errorf("path = %v, want %s", got[testPathKey], want)
				}
			})
		}
	})

	t.Run("refuses a path like topic", func(t *testing.T) {
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

				_, err := os.Stat(filepath.Join(projectStoreDir(t, dir), testResearchDir))
				if !errors.Is(err, fs.ErrNotExist) {
					t.Errorf("research folder: stat err = %v, want ErrNotExist", err)
				}

				track, err := openProjectStore(t, dir).Load(testTrackID)
				if err != nil {
					t.Fatal(err)
				}

				if track.Title != testTitle {
					t.Errorf("track title = %q, want %q", track.Title, testTitle)
				}
			})
		}
	})

	t.Run("refuses a track id that escapes", func(t *testing.T) {
		dir := t.TempDir()
		seedTasks(t, dir, &task.Task{ID: testTrackID, Title: testTitle, Status: task.StatusDoing})

		result := callToolResult(t, mcpSession(t, dir), researchWriteTool, map[string]any{
			testTrackKey: testEscapingTrack,
			testTopicKey: testIndexTopic,
			testBodyKey:  testResearchBody,
		})

		if !result.IsError {
			t.Errorf("IsError = false, want true (content %v)", result.Content)
		}

		outside := filepath.Join(filepath.Dir(projectStoreDir(t, dir)), testTrackID)
		if _, err := os.Stat(outside); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("folder outside the store: stat err = %v, want ErrNotExist", err)
		}
	})

	t.Run("records the session head", func(t *testing.T) {
		dir := t.TempDir()
		seedTasks(t, dir, &task.Task{ID: testTrackID, Title: testTitle, Status: task.StatusDoing})

		fake := headGit(testHeadSHA)
		start := time.Now().UTC().Truncate(time.Second)

		path := writeIndexTopic(t, mcpSessionWithGit(t, dir, fake.run))

		end := time.Now().UTC()

		commits := recordCommits(t, path)
		if len(commits) != 1 {
			t.Fatalf("commits = %v, want one", commits)
		}

		assertCommit(t, commits[0], testHeadSHA)

		at, err := time.Parse(time.RFC3339, commits[0][testAtKey].(string))
		if err != nil {
			t.Fatalf("at = %v: %v", commits[0][testAtKey], err)
		}

		if at.Before(start) || at.After(end) {
			t.Errorf("at = %v, want between %v and %v", at, start, end)
		}

		assertDirs(t, fake.dirs, dir)
	})

	t.Run("a later write after a new commit appends", func(t *testing.T) {
		dir := t.TempDir()
		seedTasks(t, dir, &task.Task{ID: testTrackID, Title: testTitle, Status: task.StatusDoing})

		fake := headGit(testHeadSHA)
		session := mcpSessionWithGit(t, dir, fake.run)

		writeIndexTopic(t, session)

		fake.replies[testRevParseArgs] = gitReply{out: testNextHeadSHA + "\n"}

		commits := recordCommits(t, writeIndexTopic(t, session))
		if len(commits) != 2 {
			t.Fatalf("commits = %v, want two", commits)
		}

		assertCommit(t, commits[0], testHeadSHA)
		assertCommit(t, commits[1], testNextHeadSHA)
	})

	t.Run("reads git in the session dir, not the project root", func(t *testing.T) {
		dir := t.TempDir()
		seedTasks(t, dir, &task.Task{ID: testTrackID, Title: testTitle, Status: task.StatusDoing})

		wt := filepath.Join(dir, ".claude", "worktrees", "wt")
		if err := os.MkdirAll(wt, 0o755); err != nil {
			t.Fatal(err)
		}

		fake := headGit(testHeadSHA)

		path := writeIndexTopic(t, mcpSessionWithGit(t, wt, fake.run))

		if !strings.HasPrefix(path, projectStoreDir(t, dir)) {
			t.Errorf("path = %s, want it under %s", path, projectStoreDir(t, dir))
		}

		assertDirs(t, fake.dirs, wt)
	})

	t.Run("no git records no commit", func(t *testing.T) {
		dir := t.TempDir()
		seedTasks(t, dir, &task.Task{ID: testTrackID, Title: testTitle, Status: task.StatusDoing})

		commits := recordCommits(t, writeIndexTopic(t, mcpSession(t, dir)))
		if len(commits) != 0 {
			t.Errorf("commits = %v, want none", commits)
		}
	})
}

func headGit(sha string) *fakeGit {
	return &fakeGit{replies: map[string]gitReply{
		testRevParseArgs:    {out: sha + "\n"},
		testSymbolicRefArgs: {out: testBranch + "\n"},
	}}
}

func writeIndexTopic(t *testing.T, s *mcp.ClientSession) string {
	t.Helper()

	got := callTool(t, s, researchWriteTool, map[string]any{
		testTrackKey: testTrackID,
		testTopicKey: testIndexTopic,
		testBodyKey:  testResearchBody,
	})

	path, ok := got[testPathKey].(string)
	if !ok {
		t.Fatalf("path = %v, want a string", got[testPathKey])
	}

	return path
}

func recordCommits(t *testing.T, mdPath string) []map[string]any {
	t.Helper()

	raw, ok := readRecord(t, mdPath)[testCommitsKey].([]any)
	if !ok {
		t.Fatalf("commits = %v, want a list", readRecord(t, mdPath)[testCommitsKey])
	}

	commits := make([]map[string]any, 0, len(raw))
	for _, c := range raw {
		m, ok := c.(map[string]any)
		if !ok {
			t.Fatalf("commit = %v, want an object", c)
		}

		commits = append(commits, m)
	}

	return commits
}

func assertCommit(t *testing.T, c map[string]any, sha string) {
	t.Helper()

	if c[testSHAKey] != sha || c[testBranchKey] != testBranch {
		t.Errorf("commit = %v, want sha %s on branch %s", c, sha, testBranch)
	}
}

func assertDirs(t *testing.T, dirs []string, want string) {
	t.Helper()

	if len(dirs) == 0 {
		t.Fatalf("git saw no dir, want %s", want)
	}

	for _, d := range dirs {
		if d != want {
			t.Errorf("git dir = %s, want %s", d, want)
		}
	}
}

func TestResearchReadHandler(t *testing.T) {
	t.Run("returns a topic", func(t *testing.T) {
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
	})

	t.Run("refuses a missing topic", func(t *testing.T) {
		dir := t.TempDir()
		seedTasks(t, dir, &task.Task{ID: testTrackID, Title: testTitle, Status: task.StatusDoing})

		result := callToolResult(t, mcpSession(t, dir), researchReadTool, map[string]any{
			testTrackKey: testTrackID,
			testTopicKey: testMissingTopic,
		})

		if !result.IsError {
			t.Errorf("IsError = false, want true (content %v)", result.Content)
		}
	})

	t.Run("refuses an unknown track", func(t *testing.T) {
		dir := t.TempDir()
		seedTasks(t, dir, &task.Task{ID: testTrackID, Title: testTitle, Status: task.StatusDoing})

		result := callToolResult(t, mcpSession(t, dir), researchReadTool, map[string]any{
			testTrackKey: testUnknownTrackID,
			testTopicKey: testIndexTopic,
		})

		if !result.IsError {
			t.Errorf("IsError = false, want true (content %v)", result.Content)
		}
	})

	t.Run("lists topics", func(t *testing.T) {
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
	})

	t.Run("lists nothing for a track with no research", func(t *testing.T) {
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
	})

	t.Run("refuses a path like topic", func(t *testing.T) {
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
	})

	t.Run("refuses a track id that escapes", func(t *testing.T) {
		tests := []struct {
			name string
			args map[string]any
		}{
			{name: "topic", args: map[string]any{testTrackKey: testEscapingTrack, testTopicKey: testIndexTopic}},
			{name: "listing", args: map[string]any{testTrackKey: testEscapingTrack}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				dir := t.TempDir()
				seedTasks(t, dir, &task.Task{ID: testTrackID, Title: testTitle, Status: task.StatusDoing})
				seedEscapedResearch(t, dir)

				result := callToolResult(t, mcpSession(t, dir), researchReadTool, tt.args)

				if !result.IsError {
					t.Errorf("IsError = false, want true (content %v)", result.Content)
				}
			})
		}
	})
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

func seedEscapedResearch(t *testing.T, dir string) {
	t.Helper()

	s := openProjectStore(t, dir)
	if _, err := s.WriteResearch(testTrackID, testIndexTopic, testResearchBody, task.Commit{}); err != nil {
		t.Fatal(err)
	}

	sd := projectStoreDir(t, dir)

	escaped := filepath.Join(filepath.Dir(sd), testTrackID)
	if err := os.Rename(filepath.Join(sd, testResearchDir, testTrackID), escaped); err != nil {
		t.Fatal(err)
	}
}
