package tui

import (
	"bytes"
	"strings"
	"testing"

	"charm.land/bubbles/v2/list"
	"github.com/B4Dmonkey/bit-pro/task"
)

func TestDelegate(t *testing.T) {
	t.Parallel()

	t.Run("selected row uses terminal green", func(t *testing.T) {
		t.Parallel()

		l := list.New([]list.Item{item{t: &task.Task{ID: ttid1, Title: ttTrack}}}, delegate{}, 40, 4)

		var buf bytes.Buffer

		delegate{}.Render(&buf, l, 0, l.Items()[0])

		got := buf.String()
		if !strings.Contains(got, "32m") {
			t.Errorf("selected row = %q, want terminal green SGR 32m", got)
		}

		if strings.Contains(got, "38;5;99") {
			t.Errorf("selected row = %q, still contains 256-purple 38;5;99", got)
		}
	})

	t.Run("unselected bar row follows terminal default", func(t *testing.T) {
		t.Parallel()

		l := list.New([]list.Item{
			item{t: &task.Task{ID: ttid1, Title: ttTrack}},
			item{t: &task.Task{ID: ttid1_1, Title: ttBar}},
		}, delegate{}, 40, 6)
		l.Select(0)

		var buf bytes.Buffer
		delegate{}.Render(&buf, l, 1, l.Items()[1])

		got := buf.String()
		if strings.Contains(got, "38;5;245") {
			t.Errorf("unselected bar row = %q, still contains gray 38;5;245", got)
		}

		if !strings.Contains(got, ttid1_1) {
			t.Errorf("unselected bar row = %q, want row text BIT-1.1", got)
		}
	})

	t.Run("selected board card inverted", func(t *testing.T) {
		t.Parallel()

		l := list.New([]list.Item{item{t: &task.Task{ID: ttid1, Title: "Card"}}}, delegate{board: true}, 40, 4)

		var buf bytes.Buffer
		delegate{board: true}.Render(&buf, l, 0, l.Items()[0])

		got := buf.String()
		if !strings.Contains(got, "\x1b[7;32m") {
			t.Errorf("selected board card = %q, want reverse-green SGR \\x1b[7;32m", got)
		}
	})

	t.Run("selected list row not inverted", func(t *testing.T) {
		t.Parallel()

		l := list.New([]list.Item{item{t: &task.Task{ID: ttid1, Title: "Row"}}}, delegate{}, 40, 4)

		var buf bytes.Buffer

		delegate{}.Render(&buf, l, 0, l.Items()[0])

		got := buf.String()
		if !strings.Contains(got, "32m") {
			t.Errorf("selected list row = %q, want terminal green 32m", got)
		}

		if strings.Contains(got, "\x1b[7;32m") {
			t.Errorf("selected list row = %q, should not be reverse-inverted \\x1b[7;32m", got)
		}
	})

	t.Run("done row shows marker", func(t *testing.T) {
		t.Parallel()

		l := list.New([]list.Item{item{t: &task.Task{ID: ttid1, Title: ttTrack, Status: task.StatusDone}}}, delegate{}, 40, 4)

		var buf bytes.Buffer

		delegate{}.Render(&buf, l, 0, l.Items()[0])

		got := buf.String()
		if !strings.Contains(got, "✓") {
			t.Errorf("done row = %q, want ✓ marker", got)
		}
	})

	t.Run("unfinished row has no marker", func(t *testing.T) {
		t.Parallel()

		l := list.New([]list.Item{item{t: &task.Task{ID: ttid1, Title: ttTrack, Status: task.StatusTodo}}}, delegate{}, 40, 4)

		var buf bytes.Buffer

		delegate{}.Render(&buf, l, 0, l.Items()[0])

		got := buf.String()
		if strings.Contains(got, "✓") {
			t.Errorf("unfinished row = %q, should have no ✓ marker", got)
		}
	})

	t.Run("doing row shows in progress marker", func(t *testing.T) {
		t.Parallel()

		doing := item{t: &task.Task{ID: ttid1, Title: ttTrack, Status: task.StatusDoing}}
		l := list.New([]list.Item{doing}, delegate{}, 40, 4)

		var buf bytes.Buffer

		delegate{}.Render(&buf, l, 0, l.Items()[0])

		got := buf.String()
		if !strings.Contains(got, "→") {
			t.Errorf("doing row = %q, want → marker", got)
		}

		if strings.Contains(got, "✓") {
			t.Errorf("doing row = %q, should not have ✓ marker", got)
		}
	})

	t.Run("board card has no in progress marker", func(t *testing.T) {
		t.Parallel()

		l := list.New(
			[]list.Item{item{t: &task.Task{ID: ttid1, Title: "Card", Status: task.StatusDoing}}},
			delegate{board: true}, 40, 4,
		)

		var buf bytes.Buffer

		delegate{board: true}.Render(&buf, l, 0, l.Items()[0])

		got := buf.String()
		if strings.Contains(got, "→") {
			t.Errorf("board card = %q, should not have → marker", got)
		}

		if !strings.Contains(got, ttid1) {
			t.Errorf("board card = %q, want card text BIT-1", got)
		}
	})

	t.Run("unapproved item renders yellow", func(t *testing.T) {
		t.Parallel()

		l := list.New([]list.Item{
			item{t: &task.Task{ID: ttid1, Title: ttTrack, Approved: false}},
			item{t: &task.Task{ID: ttid1_1, Title: ttBar, Approved: true}},
		}, delegate{}, 40, 6)
		l.Select(1)

		var buf bytes.Buffer
		delegate{}.Render(&buf, l, 0, l.Items()[0])

		got := buf.String()
		if !strings.Contains(got, "33m") {
			t.Errorf("unapproved unselected row = %q, want terminal yellow SGR 33m", got)
		}
	})

	t.Run("approved item does not render yellow", func(t *testing.T) {
		t.Parallel()

		l := list.New([]list.Item{
			item{t: &task.Task{ID: ttid1, Title: ttTrack, Approved: true}},
			item{t: &task.Task{ID: ttid1_1, Title: ttBar, Approved: true}},
		}, delegate{}, 40, 6)
		l.Select(1)

		var buf bytes.Buffer
		delegate{}.Render(&buf, l, 0, l.Items()[0])

		got := buf.String()
		if strings.Contains(got, "33m") {
			t.Errorf("approved unselected row = %q, should not contain terminal yellow SGR 33m", got)
		}
	})

	t.Run("selected unapproved item keeps yellow text", func(t *testing.T) {
		t.Parallel()

		l := list.New([]list.Item{item{t: &task.Task{ID: ttid1, Title: "T", Approved: false}}}, delegate{}, 40, 4)

		var buf bytes.Buffer
		delegate{}.Render(&buf, l, 0, l.Items()[0])

		got := buf.String()
		if !strings.Contains(got, "33m") {
			t.Errorf("selected unapproved row = %q, want terminal yellow SGR 33m (approval color survives focus)", got)
		}
	})

	t.Run("selected approved item cursor stays green", func(t *testing.T) {
		t.Parallel()

		l := list.New([]list.Item{item{t: &task.Task{ID: ttid1, Title: "T", Approved: true}}}, delegate{}, 40, 4)

		var buf bytes.Buffer
		delegate{}.Render(&buf, l, 0, l.Items()[0])

		got := buf.String()
		if !strings.Contains(got, "32m") {
			t.Errorf("selected approved row = %q, want terminal green SGR 32m (cursor)", got)
		}

		if strings.Contains(got, "33m") {
			t.Errorf("selected approved row = %q, should not contain terminal yellow SGR 33m", got)
		}
	})

	t.Run("selected unapproved item cursor stays green", func(t *testing.T) {
		t.Parallel()

		l := list.New([]list.Item{item{t: &task.Task{ID: ttid1, Title: ttTrack, Approved: false}}}, delegate{}, 40, 4)

		var buf bytes.Buffer
		delegate{}.Render(&buf, l, 0, l.Items()[0])

		got := buf.String()
		if !strings.Contains(got, "32m") {
			t.Errorf("selected unapproved row = %q, want terminal green SGR 32m (cursor)", got)
		}
	})

	t.Run("track vs bar distinguished by weight not color", func(t *testing.T) {
		t.Parallel()

		l := list.New([]list.Item{
			item{t: &task.Task{ID: ttid1, Title: ttTrack, Approved: true}},
			item{t: &task.Task{ID: ttid1_1, Title: ttBar, Approved: true}},
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
	})
}
