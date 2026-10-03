package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/B4Dmonkey/bit-pro/db/orm"
	"github.com/B4Dmonkey/bit-pro/git"
	"github.com/B4Dmonkey/bit-pro/project"
	"github.com/B4Dmonkey/bit-pro/store"
	"github.com/B4Dmonkey/bit-pro/task"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var sandboxed sync.Map

func mcpSandbox(t *testing.T) {
	t.Helper()

	if _, done := sandboxed.LoadOrStore(t, struct{}{}); done {
		return
	}

	t.Cleanup(func() { sandboxed.Delete(t) })

	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_DATA_HOME", t.TempDir())
}

func registerProject(t *testing.T, dir string) {
	t.Helper()

	mcpSandbox(t)

	if _, err := project.Find(t.Context(), dir); err == nil {
		return
	}

	path, err := project.CanonicalPath(dir)
	if err != nil {
		t.Fatalf("CanonicalPath(%q) returned error: %v", dir, err)
	}

	seedProject(t, orm.CreateProjectParams{Path: path, Code: testCode})
}

func openProjectStore(t *testing.T, dir string) *task.Store {
	t.Helper()

	s, err := project.OpenStore(t.Context(), dir)
	if err != nil {
		t.Fatalf("project.OpenStore(%q) returned error: %v", dir, err)
	}

	return s
}

func projectStoreDir(t *testing.T, dir string) string {
	t.Helper()

	p, err := project.Find(t.Context(), dir)
	if err != nil {
		t.Fatalf("project.Find(%q) returned error: %v", dir, err)
	}

	sd, err := store.ProjectDir(p.Code)
	if err != nil {
		t.Fatalf("store.ProjectDir(%q) returned error: %v", p.Code, err)
	}

	return sd
}

var errNoGit = errors.New("no git in this test")

func noGit(context.Context, string, ...string) (string, error) {
	return "", errNoGit
}

type gitReply struct {
	out string
	err error
}

type fakeGit struct {
	replies map[string]gitReply
	dirs    []string
}

func (f *fakeGit) run(_ context.Context, dir string, args ...string) (string, error) {
	f.dirs = append(f.dirs, dir)

	r, ok := f.replies[strings.Join(args, " ")]
	if !ok {
		return "", fmt.Errorf("fakeGit: no reply for %q", args)
	}

	return strings.TrimSpace(r.out), r.err
}

func readRecord(t *testing.T, mdPath string) map[string]any {
	t.Helper()

	data, err := os.ReadFile(strings.TrimSuffix(mdPath, filepath.Ext(mdPath)) + ".json")
	if err != nil {
		t.Fatal(err)
	}

	var rec map[string]any
	if err := json.Unmarshal(data, &rec); err != nil {
		t.Fatal(err)
	}

	return rec
}

func mcpSession(t *testing.T, root string) *mcp.ClientSession {
	t.Helper()

	return mcpSessionWithGit(t, root, noGit)
}

func mcpSessionWithGit(t *testing.T, root string, run git.Runner) *mcp.ClientSession {
	t.Helper()

	mcpSandbox(t)

	ctx, cancel := context.WithCancel(t.Context())

	serverT, clientT := mcp.NewInMemoryTransports()

	errCh := make(chan error, 1)
	go func() { errCh <- runMCPServer(ctx, root, run, serverT) }()

	t.Cleanup(func() {
		cancel()
		<-errCh
	})

	session, err := mcp.NewClient(&mcp.Implementation{Name: testClient, Version: "1"}, nil).Connect(ctx, clientT, nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { session.Close() })

	return session
}

func callTool(t *testing.T, s *mcp.ClientSession, name string, args map[string]any) map[string]any {
	t.Helper()

	var got map[string]any

	decodeToolResult(t, s, name, args, &got)

	return got
}

func callToolList(t *testing.T, s *mcp.ClientSession, name string, args map[string]any) []map[string]any {
	t.Helper()

	var got struct {
		Tasks []map[string]any `json:"tasks"`
	}

	decodeToolResult(t, s, name, args, &got)

	return got.Tasks
}

func callToolResult(t *testing.T, s *mcp.ClientSession, name string, args map[string]any) *mcp.CallToolResult {
	t.Helper()

	result, err := s.CallTool(t.Context(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatal(err)
	}

	return result
}

func decodeToolResult(t *testing.T, s *mcp.ClientSession, name string, args map[string]any, into any) {
	t.Helper()

	result, err := s.CallTool(t.Context(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatal(err)
	}

	if result.IsError {
		t.Fatalf("%s returned error: %v", name, result.Content)
	}

	b, err := json.Marshal(result.StructuredContent)
	if err != nil {
		t.Fatal(err)
	}

	if err := json.Unmarshal(b, into); err != nil {
		t.Fatal(err)
	}
}

func seedTasks(t *testing.T, dir string, tasks ...*task.Task) {
	t.Helper()

	registerProject(t, dir)

	s := openProjectStore(t, dir)

	for _, tk := range tasks {
		if err := s.Save(tk); err != nil {
			t.Fatal(err)
		}
	}
}
