package usage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/B4Dmonkey/bit-pro/db/orm"
	"github.com/B4Dmonkey/bit-pro/task"
)

func Build(
	ctx context.Context,
	q *orm.Queries,
	projectID int64,
	track, status string,
	bars []*task.Task,
) (string, int, error) {
	rows, err := q.UsageForTrack(ctx, orm.UsageForTrackParams{ProjectID: projectID, Track: track})
	if err != nil {
		return "", 0, fmt.Errorf("loading usage for %s: %w", track, err)
	}

	priced := make([]Priced, 0, len(rows))

	for _, r := range rows {
		price, err := q.PriceAt(ctx, orm.PriceAtParams{Model: r.Model, EffectiveFrom: r.Ts})
		missing := errors.Is(err, sql.ErrNoRows)

		if err != nil && !missing {
			return "", 0, fmt.Errorf("pricing %s: %w", r.Model, err)
		}

		priced = append(priced, Priced{Row: r, Price: price, Missing: missing})
	}

	return Render(status, priced, bars), len(rows), nil
}
