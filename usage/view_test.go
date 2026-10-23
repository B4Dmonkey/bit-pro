package usage

import (
	"strings"
	"testing"

	"github.com/B4Dmonkey/bit-pro/db/orm"
	"github.com/B4Dmonkey/bit-pro/task"
)

func TestRender(t *testing.T) {
	opus := orm.ModelPrice{Input: rate(4)}
	sonnet := orm.ModelPrice{CacheRead: rate(0.20), Output: rate(10)}

	t.Run("totals across models", func(t *testing.T) {
		got := Render(task.StatusDoing, []Priced{
			{Row: orm.Usage{InputTokens: 1_000_000}, Price: opus},
			{Row: orm.Usage{CacheReadTokens: 1_000_000, OutputTokens: 100_000}, Price: sonnet},
		}, nil)

		for _, want := range []string{"| Total cost | $5.20 |", "| Total tokens | 2.10M (in 2.00M · out 100K) |"} {
			if !strings.Contains(got, want) {
				t.Errorf("Render() = %q, want it to contain %q", got, want)
			}
		}
	})

	t.Run("unknown model marks the total", func(t *testing.T) {
		got := Render(task.StatusDoing, []Priced{
			{Row: orm.Usage{InputTokens: 1_000_000}, Price: opus},
			{Row: orm.Usage{Model: "claude-nextgen-9", InputTokens: 10, OutputTokens: 10}, Missing: true},
		}, nil)

		for _, want := range []string{"| Total cost | $4.00 + unknown |", "| Total tokens | 1.00M (in 1.00M · out 10) |"} {
			if !strings.Contains(got, want) {
				t.Errorf("Render() = %q, want it to contain %q", got, want)
			}
		}
	})

	t.Run("only unknown", func(t *testing.T) {
		got := Render(task.StatusDoing, []Priced{{Row: orm.Usage{InputTokens: 10}, Missing: true}}, nil)

		if want := "| Total cost | unknown |"; !strings.Contains(got, want) {
			t.Errorf("Render() = %q, want it to contain %q", got, want)
		}
	})

	t.Run("as of the latest row", func(t *testing.T) {
		got := Render(task.StatusDoing, []Priced{
			{Row: orm.Usage{Ts: "2026-10-14T23:30:00Z", InputTokens: 1}, Price: opus},
			{Row: orm.Usage{Ts: "2026-10-20T09:00:00Z", InputTokens: 1}, Price: opus},
		}, nil)

		lines := strings.Split(strings.TrimSpace(got), "\n")
		if want := "_as of 2026-10-20_"; lines[0] != want {
			t.Errorf("first line = %q, want %q", lines[0], want)
		}

		if last := lines[len(lines)-1]; last != closingNote {
			t.Errorf("last line = %q, want %q", last, closingNote)
		}
	})

	t.Run("done track is marked complete", func(t *testing.T) {
		row := Priced{Row: orm.Usage{Ts: "2026-10-20T09:00:00Z", InputTokens: 1}, Price: opus}
		got := Render(task.StatusDone, []Priced{row}, nil)

		if want := "_as of 2026-10-20 (track complete)_"; !strings.Contains(got, want) {
			t.Errorf("Render() = %q, want it to contain %q", got, want)
		}
	})

	t.Run("agents grouped then by cost", func(t *testing.T) {
		row := func(agent string, tokens int64) Priced {
			return Priced{Row: orm.Usage{InvokedAgent: agent, InputTokens: tokens}, Price: opus}
		}

		got := Render(task.StatusDoing, []Priced{
			row("bit:theCreator", 250_000),
			row("main", 500_000),
			row("bit:ruler", 750_000),
			row("bit:bot", 50_000),
			row("other", 1_000_000),
		}, nil)

		heads := []string{
			"**bit:bot** — $0.20",
			"**bit:ruler** — $3.00",
			"**bit:theCreator** — $1.00",
			"**other** — $4.00",
			"**main** — $2.00",
		}

		if !strings.Contains(got, "### By agent") {
			t.Fatalf("Render() = %q, want a By agent section", got)
		}

		prev := -1

		for _, h := range heads {
			i := strings.Index(got, h)
			if i < 0 || i < prev {
				t.Fatalf("Render() = %q, want %q after the previous head", got, h)
			}

			prev = i
		}

		if want := "| Total cost | $10.20 |"; !strings.Contains(got, want) {
			t.Errorf("Render() = %q, want it to contain %q", got, want)
		}
	})

	t.Run("agent block with skills and own work", func(t *testing.T) {
		row := func(skill, inv string, tokens int64) Priced {
			return Priced{Row: orm.Usage{
				InvokedAgent: "bit:ruler", Model: opusModel, Skill: skill, SkillInvocation: inv, InputTokens: tokens,
			}, Price: opus}
		}

		got := Render(task.StatusDoing, []Priced{
			row("bit:scope", "i1", 5_000_000),
			row("bit:scope", "i2", 5_000_000),
			row("", "", 2_500_000),
		}, nil)

		prev := -1

		for _, want := range []string{
			"**bit:ruler** — $50.00",
			"bit:scope (opus) — 2 × $20.00 = $40.00",
			"own work (opus) — $10.00",
		} {
			i := strings.Index(got, want)
			if i < 0 || i < prev {
				t.Fatalf("Render() = %q, want %q after the previous line", got, want)
			}

			prev = i
		}
	})

	t.Run("token formats", func(t *testing.T) {
		tests := []struct {
			n    int64
			want string
		}{{999, "999"}, {1_000, "1K"}, {100_000, "100K"}, {1_000_000, "1.00M"}, {2_100_000, "2.10M"}}

		for _, tt := range tests {
			if got := formatTokens(tt.n); got != tt.want {
				t.Errorf("formatTokens(%d) = %q, want %q", tt.n, got, tt.want)
			}
		}
	})
}
