package store

import (
	"context"

	"github.com/maxzhirnov/rewardify/internal/models"
)

// GetUsersBalance queries users bonus amount from database by userUUID and returns UsersBalance object or error
func (p *Postgres) GetUsersBalance(ctx context.Context, userUUID string) (models.UsersBalance, error) {
	userBalance := models.UsersBalance{
		UserUUID: userUUID,
	}

	sql := `
SELECT total_bonus, redeemed_bonus, (total_bonus - redeemed_bonus) as current from balances WHERE user_uuid=$1
`

	row := p.DB.QueryRowContext(ctx, sql, userUUID)

	err := row.Scan(&userBalance.Earned, &userBalance.Withdrawn, &userBalance.Current)
	if err != nil {
		return userBalance, err
	}

	return userBalance, nil
}
