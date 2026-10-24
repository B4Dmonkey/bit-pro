package usage

import (
	"database/sql"
	"math"
	"testing"

	"github.com/B4Dmonkey/bit-pro/db/orm"
)

func rate(v float64) sql.NullFloat64 {
	return sql.NullFloat64{Float64: v, Valid: true}
}

func TestCost(t *testing.T) {
	t.Run("each class at its rate", func(t *testing.T) {
		const mtok = 1_000_000

		row := orm.Usage{
			InputTokens: mtok, CacheWrite5mTokens: mtok, CacheWrite1hTokens: mtok,
			CacheReadTokens: mtok, OutputTokens: mtok,
		}
		price := orm.ModelPrice{
			Input: rate(4), CacheWrite5m: rate(5), CacheWrite1h: rate(8), CacheRead: rate(0.20), Output: rate(20),
		}

		got := Cost(row, price)

		if want := 37.20; math.Abs(got-want) > 1e-9 {
			t.Errorf("Cost() = %v, want %v", got, want)
		}
	})
}

func TestPricedAmount(t *testing.T) {
	haiku := orm.ModelPrice{CacheRead: rate(1)}

	tests := []struct {
		name        string
		row         orm.Usage
		wantUnknown bool
	}{
		{"haiku 5.5 over 100K", orm.Usage{Model: "claude-haiku-5-5", CacheReadTokens: 100_001}, true},
		{"haiku 5.5 at 100K", orm.Usage{Model: "claude-haiku-5-5", CacheReadTokens: 100_000}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Priced{Row: tt.row, Price: haiku}.Amount()

			if got.Unknown != tt.wantUnknown {
				t.Errorf("Amount().Unknown = %v, want %v", got.Unknown, tt.wantUnknown)
			}
		})
	}
}
