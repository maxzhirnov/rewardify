package repo

import (
	"context"

	"github.com/maxzhirnov/rewardify/internal/models"
)

// GetUserByUsername queries user from database by username and returns user object or error if occurred
func (p *Postgres) GetUserByUsername(ctx context.Context, username string) (models.User, error) {
	user := models.User{}

	sql := `SELECT uuid, username, password, created_at FROM users WHERE username=$1`

	row := p.db.QueryRowContext(ctx, sql, username)

	err := row.Scan(&user.UUID, &user.Username, &user.Password, &user.CreateAt)
	if err != nil {
		p.logger.Log.Error(err)
		return models.User{}, err
	}

	return user, nil
}
