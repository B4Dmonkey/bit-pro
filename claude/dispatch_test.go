package claude

import (
	"context"
	"os"
	"slices"
	"testing"
)

func TestAgents_ParsesTheRealPayload(t *testing.T) {
	payload, err := os.ReadFile("testdata/agents.json")
	if err != nil {
		t.Fatalf("ReadFile() returned error: %v", err)
	}

	var got call

	run := func(_ context.Context, dir, name string, args ...string) (string, int, error) {
		got = call{dir: dir, name: name, args: args}

		return string(payload), 0, nil
	}

	agents, err := Agents(t.Context(), run)
	if err != nil {
		t.Fatalf("Agents() returned error: %v", err)
	}

	want := []Agent{
		{Name: "acme-7b", Cwd: "/tmp/acme"},
		{Name: "ACME-1-a-track", Cwd: "/tmp/acme/.claude/worktrees/ACME-1-a-track"},
		{Name: "6a4a7973", Cwd: "/tmp/other"},
	}

	if !slices.Equal(agents, want) {
		t.Errorf("Agents() = %+v, want %+v", agents, want)
	}

	wantCall := call{dir: "", name: "claude", args: []string{"agents", "--json"}}

	if got.dir != wantCall.dir || got.name != wantCall.name || !slices.Equal(got.args, wantCall.args) {
		t.Errorf("Agents() ran %+v, want %+v", got, wantCall)
	}
}

type call struct {
	dir  string
	name string
	args []string
}

func TestWorktreeName(t *testing.T) {
	tests := []struct {
		trackID string
		title   string
		want    string
	}{
		{"ACME-1", "a track", "ACME-1-a-track"},
		{
			"BIT-39",
			"Dispatch — the daemon works queued bars unattended",
			"BIT-39-dispatch-the-daemon-works-queued-bars-unattended",
		},
		{"BIT-7", "bp init registers the MCP server", "BIT-7-bp-init-registers-the-mcp-server"},
		{"BIT-8", "  spaced  out  ", "BIT-8-spaced-out"},
		{"BIT-9", "slash/and.dot", "BIT-9-slash-and-dot"},
	}

	for _, tt := range tests {
		t.Run(tt.trackID, func(t *testing.T) {
			if got := WorktreeName(tt.trackID, tt.title); got != tt.want {
				t.Errorf("WorktreeName(%q, %q) = %q, want %q", tt.trackID, tt.title, got, tt.want)
			}
		})
	}
}
