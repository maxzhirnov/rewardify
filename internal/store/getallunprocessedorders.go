package store

import (
	"context"

	"github.com/maxzhirnov/rewardify/internal/models"
)

func (p *Postgres) GetAllUnprocessedOrders(ctx context.Context) ([]models.Order, error) {
	sql := `
SELECT order_number, user_uuid, bonus_accrual_status, created_at 
FROM orders
WHERE bonus_accrual_status NOT IN ('PROCESSED', 'INVALID')
`

	rows, err := p.DB.QueryContext(ctx, sql)
	if err != nil {
		return nil, err
	}

	unprocessedOrders := make([]models.Order, 0)

	for rows.Next() {
		o := models.Order{}
		err := rows.Scan(&o.OrderNumber, &o.UserUUID, &o.BonusAccrualStatus, &o.CreatedAt)
		if err != nil {
			return nil, err
		}
		unprocessedOrders = append(unprocessedOrders, o)
	}

	return unprocessedOrders, nil
}
