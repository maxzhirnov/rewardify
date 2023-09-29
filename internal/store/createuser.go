package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/maxzhirnov/rewardify/internal/models"
)

func (p *Postgres) CreateUser(user models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sql := `INSERT INTO users (uuid, username, password, created_at) values ($1, $2, $3, now())`

	if _, err := p.DB.ExecContext(ctx, sql, user.UUID, user.Username, user.Password); err != nil {
		p.logger.Log.Debug(err)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return ErrUserAlreadyExist
			}
		}
		return err
	}

	return nil
}
