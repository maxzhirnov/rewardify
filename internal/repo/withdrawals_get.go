package repo

import (
	"context"

	"github.com/maxzhirnov/rewardify/internal/models"
)

func (p *Postgres) GetUsersWithdrawals(ctx context.Context, usrUUID string) ([]models.Withdrawal, error) {
	sql := `
SELECT order_number, withdrew, created_at FROM withdrawals
WHERE user_uuid=$1
ORDER BY created_at
`
	rows, err := p.db.QueryContext(ctx, sql, usrUUID)
	if err != nil {
		return nil, err
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	withdrawals := make([]models.Withdrawal, 0)
	for rows.Next() {
		w := models.Withdrawal{UserUUID: usrUUID}
		err := rows.Scan(&w.OrderNumber, &w.Amount, &w.CreatedAt)
		if err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, w)
	}
	return withdrawals, nil

}
