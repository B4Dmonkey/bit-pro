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

	return fmt.Sprintf("%s| | |\n|---|---|\n| Total cost | %s |\n| Total tokens | %s (in %s · out %s) |\n\n%s%s\n%s\n",
		asOf, cost, formatTokens(in+out), formatTokens(in), formatTokens(out),
		renderAgents(rows), renderBars(rows, bars), closingNote)
}

const barsNote = "_bar-level work only — the agent totals above include it too_"

func renderBars(rows []Priced, bars []*task.Task) string {
	byBar, labelsByBar := barSpend(rows)

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
			fmt.Fprintf(&sb, "| ↳ Bar %s |%s| %s |\n", barNumber(b.ID), labelCell(labelsByBar[b.ID]), byBar[b.ID])
		}
	}

	return sb.String() + "\n" + barsNote + "\n\n"
}

func barSpend(rows []Priced) (map[string]Amount, map[string]map[string]bool) {
	byBar := map[string]Amount{}
	labels := map[string]map[string]bool{}

	for _, r := range rows {
		if r.Row.Bar == "" {
			continue
		}

		byBar[r.Row.Bar] = byBar[r.Row.Bar].Add(r.Amount())

		if labels[r.Row.Bar] == nil {
			labels[r.Row.Bar] = map[string]bool{}
		}

		if r.Row.Skill != "" {
			labels[r.Row.Bar][r.Row.Skill] = true
		}

		if r.Row.AgentID != "" {
			labels[r.Row.Bar][r.Row.AgentType+" (subagent)"] = true
		}
	}

	return byBar, labels
}

func labelCell(set map[string]bool) string {
	if len(set) == 0 {
		return " "
	}

	labels := make([]string, 0, len(set))
	for l := range set {
		labels = append(labels, l)
	}

	sort.Strings(labels)

	return " " + strings.Join(labels, ", ") + " "
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

	sb.WriteString("### By agent\n\n| Agent / skill | Model | Calls | Cost |\n|---|---|---:|---:|\n")

	for _, a := range agents {
		fmt.Fprintf(&sb, "| **%s** | | | %s |\n%s", a.name, a.cost, renderSkills(rowsByAgent[a.name]))
	}

	return sb.String() + "\n"
}

type skillLine struct {
	skill, model string
	cost         Amount
	calls        map[string]bool
}

type subagentLine struct {
	typ, model string
	cost       Amount
	ids        []string
}

type agentTree struct {
	rowsByID map[string][]Priced
	children map[string][]string
}

func newAgentTree(rows []Priced) agentTree {
	t := agentTree{rowsByID: map[string][]Priced{}, children: map[string][]string{}}

	for _, r := range rows {
		t.rowsByID[r.Row.AgentID] = append(t.rowsByID[r.Row.AgentID], r)
	}

	var ids []string

	for id := range t.rowsByID {
		if id != "" {
			ids = append(ids, id)
		}
	}

	sort.Strings(ids)

	for _, id := range ids {
		parent := t.rowsByID[id][0].Row.ParentAgentID
		if _, ok := t.rowsByID[parent]; !ok || parent == id {
			parent = ""
		}

		t.children[parent] = append(t.children[parent], id)
	}

	return t
}

func (t agentTree) subtreeCost(id string) Amount {
	var total Amount

	for _, r := range t.rowsByID[id] {
		total = total.Add(r.Amount())
	}

	for _, c := range t.children[id] {
		total = total.Add(t.subtreeCost(c))
	}

	return total
}

func renderSkills(rows []Priced) string {
	t := newAgentTree(rows)

	return t.render([]string{""}, 0)
}

func (t agentTree) render(ids []string, depth int) string {
	indent := strings.Repeat("\u00a0\u00a0", depth) + "↳ "

	lines := map[[2]string]*skillLine{}

	var order []*skillLine

	var rows []Priced

	var kids []string

	for _, id := range ids {
		rows = append(rows, t.rowsByID[id]...)
		kids = append(kids, t.children[id]...)
	}

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

	var out, own strings.Builder

	for _, l := range order {
		if l.skill == "" {
			fmt.Fprintf(&own, "| %sown work | %s | | %s |\n", indent, l.model, l.cost)

			continue
		}

		fmt.Fprintf(&out, "| %s%s | %s | %d | %s |\n", indent, l.skill, l.model, len(l.calls), l.cost)
	}

	for _, sl := range t.subagentLines(kids) {
		fmt.Fprintf(&out, "| %s%s (subagent) | %s | %d | %s |\n", indent, sl.typ, sl.model, len(sl.ids), sl.cost)

		out.WriteString(t.render(sl.ids, depth+1))
	}

	return out.String() + own.String()
}

func (t agentTree) subagentLines(kids []string) []*subagentLine {
	groups := map[[2]string]*subagentLine{}

	var order []*subagentLine

	for _, c := range kids {
		first := t.rowsByID[c][0].Row
		key := [2]string{first.AgentType, modelFamily(first.Model)}

		g := groups[key]
		if g == nil {
			g = &subagentLine{typ: key[0], model: key[1]}
			groups[key] = g
			order = append(order, g)
		}

		g.cost = g.cost.Add(t.subtreeCost(c))
		g.ids = append(g.ids, c)
	}

	sort.SliceStable(order, func(i, j int) bool { return order[i].cost.USD > order[j].cost.USD })

	return order
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
