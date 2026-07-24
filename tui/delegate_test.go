package tui

import (
	"bytes"
	"strings"
	"testing"

	"github.com/B4Dmonkey/bit-pro/task"
	"charm.land/bubbles/v2/list"
)

func TestDelegate_SelectedRowUsesTerminalGreen(t *testing.T) {
	t.Parallel()

	l := list.New([]list.Item{item{t: &task.Task{ID: "BIT-1", Title: "Track"}}}, delegate{}, 40, 4)
	var buf bytes.Buffer
	delegate{}.Render(&buf, l, 0, l.Items()[0])

	got := buf.String()
	if !strings.Contains(got, "32m") {
		t.Errorf("selected row = %q, want terminal green SGR 32m", got)
	}
	if strings.Contains(got, "38;5;99") {
		t.Errorf("selected row = %q, still contains 256-purple 38;5;99", got)
	}
}
