package usage

import (
	"fmt"
	"strconv"

	"github.com/B4Dmonkey/bit-pro/db/orm"
)

type Priced struct {
	Row     orm.Usage
	Price   orm.ModelPrice
	Missing bool
}

func Render(rows []Priced) string {
	var cost Amount

	var in, out int64

	for _, r := range rows {
		cost = cost.Add(r.Amount())
		in += r.Row.InputTokens + r.Row.CacheWrite5mTokens + r.Row.CacheWrite1hTokens + r.Row.CacheReadTokens
		out += r.Row.OutputTokens
	}

	return fmt.Sprintf("| Total cost | %s |\n| Total tokens | %s (in %s · out %s) |\n",
		cost, formatTokens(in+out), formatTokens(in), formatTokens(out))
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
