package store

import (
	"context"
	"time"

	"github.com/maxzhirnov/rewardify/internal/models"
)

func (p *Postgres) GetUserByUsername(username string) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user := models.User{}

	sql := `SELECT uuid, username, password, created_at  FROM users WHERE username=$1`

	row := p.DB.QueryRowContext(ctx, sql, username)

	err := row.Scan(&user.UUID, &user.Username, &user.Password, &user.CreateAt)
	if err != nil {
		p.logger.Log.Error(err)
		return models.User{}, err
	}

	return user, nil
}
