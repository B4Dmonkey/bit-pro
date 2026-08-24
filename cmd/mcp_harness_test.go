package cmd

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/B4Dmonkey/bit-pro/task"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func mcpSession(t *testing.T, root string) *mcp.ClientSession {
	t.Helper()

	ctx, cancel := context.WithCancel(t.Context())

	serverT, clientT := mcp.NewInMemoryTransports()

	errCh := make(chan error, 1)
	go func() { errCh <- runMCPServer(ctx, root, serverT) }()

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

	store := task.New(filepath.Join(dir, ".bit"))

	for _, tk := range tasks {
		if err := store.Save(tk); err != nil {
			t.Fatal(err)
		}
	}
}
