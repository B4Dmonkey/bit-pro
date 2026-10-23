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
			row(rulerAgent, 750_000),
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
				InvokedAgent: rulerAgent, Model: opusModel, Skill: skill, SkillInvocation: inv, InputTokens: tokens,
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

	t.Run("subagents nest under their parent", func(t *testing.T) {
		const haikuModel = "claude-haiku-4-5"

		haiku := orm.ModelPrice{Input: rate(1)}

		got := Render(task.StatusDoing, []Priced{
			{Row: orm.Usage{InvokedAgent: rulerAgent, Model: opusModel, InputTokens: 250_000}, Price: opus},
			{Row: orm.Usage{
				InvokedAgent: rulerAgent, AgentID: "g1", AgentType: "general-purpose", Model: opusModel,
				Skill: "bit:analyze", SkillInvocation: "i1", InputTokens: 500_000,
			}, Price: opus},
			{Row: orm.Usage{
				InvokedAgent: rulerAgent, AgentID: "e1", AgentType: "Explore", ParentAgentID: "g1",
				Model: haikuModel, InputTokens: 1_000_000,
			}, Price: haiku},
			{Row: orm.Usage{
				InvokedAgent: rulerAgent, AgentID: "e2", AgentType: "Explore", ParentAgentID: "g1",
				Model: haikuModel, InputTokens: 1_000_000,
			}, Price: haiku},
		}, nil)

		prev := -1

		for _, want := range []string{
			"**bit:ruler** — $5.00",
			"  general-purpose (subagent, opus) — 1 × $4.00 = $4.00",
			"    bit:analyze (opus) — 1 × $2.00 = $2.00",
			"    Explore (subagent, haiku) — 2 × $1.00 = $2.00",
			"  own work (opus) — $1.00",
		} {
			i := strings.Index(got, want)
			if i < 0 || i < prev {
				t.Fatalf("Render() = %q, want %q after the previous line", got, want)
			}

			prev = i
		}
	})

	t.Run("bar skills cell", func(t *testing.T) {
		bar := func(skill, agentID, agentType string) Priced {
			return Priced{Row: orm.Usage{
				InvokedAgent: rulerAgent, Bar: barOne, Skill: skill, AgentID: agentID, AgentType: agentType, InputTokens: 250_000,
			}, Price: opus}
		}

		bars := []*task.Task{{ID: barOne, Phase: 1, PhaseLabel: "first"}}

		got := Render(task.StatusDoing, []Priced{
			bar("bit:plan", "", ""), bar("bit:do", "", ""), bar("", "e1", "Explore"),
		}, bars)

		if want := "| ↳ Bar 1.1 | Explore (subagent), bit:do, bit:plan | $3.00 |"; !strings.Contains(got, want) {
			t.Errorf("Render() = %q, want it to contain %q", got, want)
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
