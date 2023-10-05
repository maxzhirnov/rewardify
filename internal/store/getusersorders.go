package store

import (
	"context"

	"github.com/maxzhirnov/rewardify/internal/models"
)

// GetUsersOrders queries all orders by specific UserUUID and returns them as a slice of objects
func (p *Postgres) GetUsersOrders(ctx context.Context, userUUID string) ([]models.Order, error) {
	ctx = context.TODO()
	sql := `
SELECT
    orders.order_number,
    orders.user_uuid,
    orders.bonus_accrual_status,
    COALESCE(accruals.accrued, 0) AS accrued,
    orders.created_at
FROM orders
LEFT JOIN accruals ON orders.order_number=accruals.order_number
WHERE orders.user_uuid=$1
ORDER BY orders.created_at
	`

	rows, err := p.DB.QueryContext(ctx, sql, userUUID)
	if err != nil {
		return nil, err
	}
	orders := make([]models.Order, 0)
	for rows.Next() {
		order := models.Order{}
		err := rows.Scan(&order.OrderNumber, &order.UserUUID, &order.BonusAccrualStatus, &order.BonusesAccrued, &order.CreatedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}