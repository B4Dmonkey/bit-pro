package claude

import "testing"

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
