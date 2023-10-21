package store

import (
	"context"
)

func (p *Postgres) Ping(ctx context.Context) error {
	return p.DB.PingContext(ctx)
}
