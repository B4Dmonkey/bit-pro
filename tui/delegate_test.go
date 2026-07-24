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

func TestDelegate_UnselectedBarRowFollowsTerminalDefault(t *testing.T) {
	t.Parallel()

	l := list.New([]list.Item{
		item{t: &task.Task{ID: "BIT-1", Title: "Track"}},
		item{t: &task.Task{ID: "BIT-1.1", Title: "Bar"}},
	}, delegate{}, 40, 6)
	l.Select(0)

	var buf bytes.Buffer
	delegate{}.Render(&buf, l, 1, l.Items()[1])

	got := buf.String()
	if strings.Contains(got, "38;5;245") {
		t.Errorf("unselected bar row = %q, still contains gray 38;5;245", got)
	}
	if !strings.Contains(got, "BIT-1.1") {
		t.Errorf("unselected bar row = %q, want row text BIT-1.1", got)
	}
}

func TestDelegate_SelectedBoardCardInverted(t *testing.T) {
	t.Parallel()

	l := list.New([]list.Item{item{t: &task.Task{ID: "BIT-1", Title: "Card"}}}, delegate{board: true}, 40, 4)
	var buf bytes.Buffer
	delegate{board: true}.Render(&buf, l, 0, l.Items()[0])

	got := buf.String()
	if !strings.Contains(got, "\x1b[7;32m") {
		t.Errorf("selected board card = %q, want reverse-green SGR \\x1b[7;32m", got)
	}
}

func TestDelegate_SelectedListRowNotInverted(t *testing.T) {
	t.Parallel()

	l := list.New([]list.Item{item{t: &task.Task{ID: "BIT-1", Title: "Row"}}}, delegate{}, 40, 4)
	var buf bytes.Buffer
	delegate{}.Render(&buf, l, 0, l.Items()[0])

	got := buf.String()
	if !strings.Contains(got, "32m") {
		t.Errorf("selected list row = %q, want terminal green 32m", got)
	}
	if strings.Contains(got, "\x1b[7;32m") {
		t.Errorf("selected list row = %q, should not be reverse-inverted \\x1b[7;32m", got)
	}
}

func TestDelegate_TrackVsBarDistinguishedByWeightNotColor(t *testing.T) {
	t.Parallel()

	l := list.New([]list.Item{
		item{t: &task.Task{ID: "BIT-1", Title: "Track"}},
		item{t: &task.Task{ID: "BIT-1.1", Title: "Bar"}},
	}, delegate{}, 40, 6)

	l.Select(1)
	var trackBuf bytes.Buffer
	delegate{}.Render(&trackBuf, l, 0, l.Items()[0])

	l.Select(0)
	var barBuf bytes.Buffer
	delegate{}.Render(&barBuf, l, 1, l.Items()[1])

	track := trackBuf.String()
	bar := barBuf.String()
	if !strings.Contains(track, "\x1b[1m") {
		t.Errorf("unselected track = %q, want bold SGR \\x1b[1m", track)
	}
	if strings.Contains(bar, "\x1b[1m") {
		t.Errorf("unselected bar = %q, should not be bold", bar)
	}
	if strings.Contains(track, "38;5") {
		t.Errorf("unselected track = %q, should carry no foreground color", track)
	}
	if strings.Contains(bar, "38;5") {
		t.Errorf("unselected bar = %q, should carry no foreground color", bar)
	}
}
