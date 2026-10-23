package usage

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/B4Dmonkey/bit-pro/db/orm"
	"github.com/B4Dmonkey/bit-pro/task"
)

const closingNote = "_Updated after each skill/agent run against this track; " +
	"figures are cumulative totals, not per-run deltas._"

type Priced struct {
	Row     orm.Usage
	Price   orm.ModelPrice
	Missing bool
}

func Render(status string, rows []Priced, bars []*task.Task) string {
	var cost Amount

	var in, out int64

	var latest time.Time

	for _, r := range rows {
		cost = cost.Add(r.Amount())
		in += r.Row.InputTokens + r.Row.CacheWrite5mTokens + r.Row.CacheWrite1hTokens + r.Row.CacheReadTokens
		out += r.Row.OutputTokens

		if ts, err := time.Parse(time.RFC3339, r.Row.Ts); err == nil && ts.After(latest) {
			latest = ts
		}
	}

	asOf := ""

	if !latest.IsZero() {
		suffix := ""
		if status == task.StatusDone {
			suffix = " (track complete)"
		}

		asOf = fmt.Sprintf("_as of %s%s_\n\n", latest.UTC().Format(time.DateOnly), suffix)
	}

	return fmt.Sprintf("%s| Total cost | %s |\n| Total tokens | %s (in %s · out %s) |\n\n%s%s\n%s\n",
		asOf, cost, formatTokens(in+out), formatTokens(in), formatTokens(out),
		renderAgents(rows), renderBars(rows, bars), closingNote)
}

const barsNote = "_bar-level work only — the agent totals above include it too_"

func renderBars(rows []Priced, bars []*task.Task) string {
	byBar := map[string]Amount{}

	for _, r := range rows {
		if r.Row.Bar != "" {
			byBar[r.Row.Bar] = byBar[r.Row.Bar].Add(r.Amount())
		}
	}

	type verse struct {
		phase int
		label string
		cost  Amount
		bars  []*task.Task
	}

	var verses []*verse

	byPhase := map[int]*verse{}

	for _, b := range bars {
		cost, ok := byBar[b.ID]
		if !ok {
			continue
		}

		v := byPhase[b.Phase]
		if v == nil {
			v = &verse{phase: b.Phase, label: b.PhaseLabel}
			byPhase[b.Phase] = v
			verses = append(verses, v)
		}

		v.cost = v.cost.Add(cost)
		v.bars = append(v.bars, b)
	}

	if len(verses) == 0 {
		return ""
	}

	sort.SliceStable(verses, func(i, j int) bool { return verses[i].phase < verses[j].phase })

	var sb strings.Builder

	sb.WriteString("### By verse / bar\n\n| Verse / bar | Skills | Cost |\n|---|---|---:|\n")

	for _, v := range verses {
		fmt.Fprintf(&sb, "| **Verse %d — %s** | | **%s** |\n", v.phase, v.label, v.cost)

		for _, b := range v.bars {
			fmt.Fprintf(&sb, "| ↳ Bar %s | | %s |\n", barNumber(b.ID), byBar[b.ID])
		}
	}

	return sb.String() + "\n" + barsNote + "\n\n"
}

func barNumber(id string) string {
	if _, n, ok := strings.Cut(id, "-"); ok {
		return n
	}

	return id
}

const unknownAgent = "unknown"

type agentCost struct {
	name string
	cost Amount
}

func renderAgents(rows []Priced) string {
	byAgent := map[string]Amount{}
	rowsByAgent := map[string][]Priced{}

	for _, r := range rows {
		name := r.Row.InvokedAgent
		if name == "" {
			name = unknownAgent
		}

		byAgent[name] = byAgent[name].Add(r.Amount())
		rowsByAgent[name] = append(rowsByAgent[name], r)
	}

	agents := make([]agentCost, 0, len(byAgent))
	for name, cost := range byAgent {
		agents = append(agents, agentCost{name, cost})
	}

	sort.Slice(agents, func(i, j int) bool {
		a, b := agents[i], agents[j]
		if ra, rb := agentRank(a.name), agentRank(b.name); ra != rb {
			return ra < rb
		}

		if a.cost.USD != b.cost.USD {
			return a.cost.USD > b.cost.USD
		}

		return a.name < b.name
	})

	var sb strings.Builder

	sb.WriteString("### By agent\n\n")

	for _, a := range agents {
		fmt.Fprintf(&sb, "**%s** — %s\n\n%s", a.name, a.cost, renderSkills(rowsByAgent[a.name]))
	}

	return sb.String()
}

type skillLine struct {
	skill, model string
	cost         Amount
	calls        map[string]bool
}

func renderSkills(rows []Priced) string {
	lines := map[[2]string]*skillLine{}

	var order []*skillLine

	for _, r := range rows {
		key := [2]string{r.Row.Skill, modelFamily(r.Row.Model)}

		l := lines[key]
		if l == nil {
			l = &skillLine{skill: key[0], model: key[1], calls: map[string]bool{}}
			lines[key] = l
			order = append(order, l)
		}

		l.cost = l.cost.Add(r.Amount())
		l.calls[r.Row.SkillInvocation] = true
	}

	sort.SliceStable(order, func(i, j int) bool { return order[i].cost.USD > order[j].cost.USD })

	var skills, own strings.Builder

	for _, l := range order {
		if l.skill == "" {
			fmt.Fprintf(&own, "  own work (%s) — %s\n", l.model, l.cost)

			continue
		}

		n := len(l.calls)
		perCall := Amount{USD: l.cost.USD / float64(n), Unknown: l.cost.Unknown, Priced: l.cost.Priced}
		fmt.Fprintf(&skills, "  %s (%s) — %d × %s = %s\n", l.skill, l.model, n, perCall, l.cost)
	}

	if skills.Len()+own.Len() == 0 {
		return ""
	}

	return skills.String() + own.String() + "\n"
}

func modelFamily(model string) string {
	for _, f := range []string{"opus", "sonnet", "haiku"} {
		if strings.Contains(model, f) {
			return f
		}
	}

	return model
}

func agentRank(name string) int {
	switch strings.ToLower(strings.TrimPrefix(strings.ToLower(name), "bit:")) {
	case "bot":
		return 0
	case "ruler", "theruler":
		return 1
	case "creator", "thecreator":
		return 2
	case unknownAgent:
		return 4
	default:
		return 3
	}
}

func formatTokens(n int64) string {
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%.2fM", float64(n)/1_000_000)
	case n >= 1_000:
		return fmt.Sprintf("%dK", n/1_000)
	default:
		return strconv.FormatInt(n, 10)
	}
}
