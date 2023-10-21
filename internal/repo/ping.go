package repo

import (
	"context"
)

func (p *Postgres) Ping(ctx context.Context) error {
	return p.db.PingContext(ctx)
}
