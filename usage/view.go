package usage

import (
	"fmt"
	"strconv"
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

func Render(status string, rows []Priced) string {
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

	return fmt.Sprintf("%s| Total cost | %s |\n| Total tokens | %s (in %s · out %s) |\n\n%s\n",
		asOf, cost, formatTokens(in+out), formatTokens(in), formatTokens(out), closingNote)
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
