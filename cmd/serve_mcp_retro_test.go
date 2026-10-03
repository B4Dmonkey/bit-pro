package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	testRetroBody  = "## Proposal 1\n\n**Pattern:** ...\n"
	testRetroName  = "album-proposals"
	testNameKey    = "name"
	testProjectKey = "project"
)

func TestRetroWriteHandler(t *testing.T) {
	t.Run("stores a proposal under the code-prefixed name", func(t *testing.T) {
		dir := t.TempDir()
		seedCodedProject(t, dir, testOwnCode, testOwnTrack)

		got := callTool(t, mcpSession(t, dir), retroWriteTool, map[string]any{
			testNameKey: testRetroName,
			testBodyKey: testRetroBody,
		})

		const want = "BIT-album-proposals"
		if got[testNameKey] != want {
			t.Errorf("name = %v, want %q", got[testNameKey], want)
		}

		retroDir := filepath.Join(dataDir(t), "retro")

		body, err := os.ReadFile(filepath.Join(retroDir, want+".md"))
		if err != nil {
			t.Fatal(err)
		}

		if string(body) != testRetroBody {
			t.Errorf("body = %q, want %q", body, testRetroBody)
		}

		raw, err := os.ReadFile(filepath.Join(retroDir, want+".json"))
		if err != nil {
			t.Fatal(err)
		}

		var rec map[string]any
		if err := json.Unmarshal(raw, &rec); err != nil {
			t.Fatal(err)
		}

		for key, wantVal := range map[string]any{
			testProjectKey: testOwnCode,
			testNameKey:    want,
			"commits":      []any{},
			"content":      want + ".md",
		} {
			if !reflect.DeepEqual(rec[key], wantVal) {
				t.Errorf("record[%q] = %#v, want %#v", key, rec[key], wantVal)
			}
		}
	})

	t.Run("records the session head", func(t *testing.T) {
		dir := t.TempDir()
		seedCodedProject(t, dir, testOwnCode, testOwnTrack)

		fake := headGit(testHeadSHA)
		start := time.Now().UTC().Truncate(time.Second)

		path := writeRetro(t, mcpSessionWithGit(t, dir, fake.run))

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

	t.Run("a re-run after a new commit appends", func(t *testing.T) {
		dir := t.TempDir()
		seedCodedProject(t, dir, testOwnCode, testOwnTrack)

		fake := headGit(testHeadSHA)
		session := mcpSessionWithGit(t, dir, fake.run)

		path := writeRetro(t, session)
		created := readRecord(t, path)["created_at"]

		fake.replies[testRevParseArgs] = gitReply{out: testNextHeadSHA + "\n"}

		writeRetro(t, session)

		commits := recordCommits(t, path)
		if len(commits) != 2 {
			t.Fatalf("commits = %v, want two", commits)
		}

		assertCommit(t, commits[0], testHeadSHA)
		assertCommit(t, commits[1], testNextHeadSHA)

		if got := readRecord(t, path)["created_at"]; got != created {
			t.Errorf("created_at = %v, want %v", got, created)
		}
	})

	t.Run("no git records no commit", func(t *testing.T) {
		dir := t.TempDir()
		seedCodedProject(t, dir, testOwnCode, testOwnTrack)

		commits := recordCommits(t, writeRetro(t, mcpSession(t, dir)))
		if len(commits) != 0 {
			t.Errorf("commits = %v, want none", commits)
		}
	})
}

func writeRetro(t *testing.T, s *mcp.ClientSession) string {
	t.Helper()

	got := callTool(t, s, retroWriteTool, map[string]any{
		testNameKey: testRetroName,
		testBodyKey: testRetroBody,
	})

	name, ok := got[testNameKey].(string)
	if !ok {
		t.Fatalf("name = %v, want a string", got[testNameKey])
	}

	return filepath.Join(dataDir(t), "retro", name+".md")
}

func TestRetroListHandler(t *testing.T) {
	t.Run("lists every project's proposals", func(t *testing.T) {
		dir := t.TempDir()
		other := t.TempDir()

		seedCodedProject(t, dir, testOwnCode, testOwnTrack)
		seedCodedProject(t, other, testOtherCode, testOtherTrack)

		session := mcpSession(t, dir)
		callTool(t, session, retroWriteTool, map[string]any{testNameKey: "BIT-49-proposals", testBodyKey: testRetroBody})
		callTool(t, mcpSession(t, other), retroWriteTool, map[string]any{
			testNameKey: testRetroName,
			testBodyKey: testRetroBody,
		})

		var got struct {
			Proposals []map[string]string `json:"proposals"`
		}

		decodeToolResult(t, session, retroListTool, map[string]any{}, &got)

		want := []map[string]string{
			{testNameKey: "BIT-49-proposals", testProjectKey: testOwnCode},
			{testNameKey: "EX-album-proposals", testProjectKey: testOtherCode},
		}
		if !reflect.DeepEqual(got.Proposals, want) {
			t.Errorf("proposals = %v, want %v", got.Proposals, want)
		}
	})

	t.Run("no proposals lists none", func(t *testing.T) {
		dir := t.TempDir()
		seedCodedProject(t, dir, testOwnCode, testOwnTrack)

		got := callTool(t, mcpSession(t, dir), retroListTool, map[string]any{})

		proposals, ok := got["proposals"].([]any)
		if !ok || len(proposals) != 0 {
			t.Errorf("proposals = %#v, want []", got["proposals"])
		}
	})
}

func TestRetroReadHandler(t *testing.T) {
	t.Run("reads another project's proposal", func(t *testing.T) {
		dir := t.TempDir()
		other := t.TempDir()

		seedCodedProject(t, dir, testOwnCode, testOwnTrack)
		seedCodedProject(t, other, testOtherCode, testOtherTrack)

		callTool(t, mcpSession(t, other), retroWriteTool, map[string]any{
			testNameKey: testRetroName,
			testBodyKey: testRetroBody,
		})

		const want = "EX-album-proposals"

		got := callTool(t, mcpSession(t, dir), retroReadTool, map[string]any{testNameKey: want})

		if got[testNameKey] != want {
			t.Errorf("name = %v, want %q", got[testNameKey], want)
		}

		if got[testProjectKey] != testOtherCode {
			t.Errorf("project = %v, want %q", got[testProjectKey], testOtherCode)
		}

		if got[testBodyKey] != testRetroBody {
			t.Errorf("body = %q, want %q", got[testBodyKey], testRetroBody)
		}
	})

	t.Run("refuses an unknown or path-like name", func(t *testing.T) {
		tests := []struct {
			name  string
			retro string
		}{
			{name: "unknown", retro: "BIT-none-proposals"},
			{name: "path-like", retro: "../x"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				dir := t.TempDir()
				seedCodedProject(t, dir, testOwnCode, testOwnTrack)

				result := callToolResult(t, mcpSession(t, dir), retroReadTool, map[string]any{testNameKey: tt.retro})
				if !result.IsError {
					t.Fatalf("IsError = false, want true (content %v)", result.Content)
				}
			})
		}
	})
}
