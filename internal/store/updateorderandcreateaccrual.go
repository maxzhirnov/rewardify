package store

import (
	"context"

	"github.com/maxzhirnov/rewardify/internal/models"
)

func (p *Postgres) UpdateOrderAndCreateAccrual(ctx context.Context, order models.Order, newStatus string) error {
	tx, err := p.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	sqlUpdateOrder := `
UPDATE orders 
SET bonus_accrual_status=$1
WHERE order_number=$2
`
	_, err = tx.ExecContext(ctx, sqlUpdateOrder, newStatus, order.OrderNumber)
	if err != nil {
		return err
	}

	sqlInsertAccrual := `
INSERT INTO accruals_calculated (user_uuid, order_number, accrued) 
VALUES ($1, $2, $3)
`
	_, err = tx.ExecContext(ctx, sqlInsertAccrual, order.UserUUID, order.OrderNumber, order.BonusesAccrued)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
