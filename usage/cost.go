package usage

import (
	"fmt"
	"strings"

	"github.com/B4Dmonkey/bit-pro/db/orm"
)

const (
	perMTok         = 1_000_000
	haiku55Prefix   = "claude-haiku-5-5"
	haiku55TierEdge = 100_000
)

type Amount struct {
	USD     float64
	Unknown bool
	Priced  bool
}

func (a Amount) Add(b Amount) Amount {
	return Amount{USD: a.USD + b.USD, Unknown: a.Unknown || b.Unknown, Priced: a.Priced || b.Priced}
}

func (a Amount) String() string {
	switch {
	case !a.Priced:
		return "unknown"
	case a.Unknown:
		return fmt.Sprintf("$%.2f + unknown", a.USD)
	default:
		return fmt.Sprintf("$%.2f", a.USD)
	}
}

func (p Priced) Amount() Amount {
	r := p.Row
	prompt := r.InputTokens + r.CacheWrite5mTokens + r.CacheWrite1hTokens + r.CacheReadTokens

	if p.Missing || (strings.HasPrefix(r.Model, haiku55Prefix) && prompt > haiku55TierEdge) {
		return Amount{Unknown: true}
	}

	return Amount{USD: Cost(r, p.Price), Priced: true}
}

func Cost(row orm.Usage, price orm.ModelPrice) float64 {
	return (float64(row.InputTokens)*price.Input.Float64 +
		float64(row.CacheWrite5mTokens)*price.CacheWrite5m.Float64 +
		float64(row.CacheWrite1hTokens)*price.CacheWrite1h.Float64 +
		float64(row.CacheReadTokens)*price.CacheRead.Float64 +
		float64(row.OutputTokens)*price.Output.Float64) / perMTok
}
