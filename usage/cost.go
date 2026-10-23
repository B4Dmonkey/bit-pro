package usage

import "github.com/B4Dmonkey/bit-pro/db/orm"

const perMTok = 1_000_000

func Cost(row orm.Usage, price orm.ModelPrice) float64 {
	return (float64(row.InputTokens)*price.Input.Float64 +
		float64(row.CacheWrite5mTokens)*price.CacheWrite5m.Float64 +
		float64(row.CacheWrite1hTokens)*price.CacheWrite1h.Float64 +
		float64(row.CacheReadTokens)*price.CacheRead.Float64 +
		float64(row.OutputTokens)*price.Output.Float64) / perMTok
}
