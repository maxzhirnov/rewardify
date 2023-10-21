package store

import (
	"context"

	"github.com/maxzhirnov/rewardify/internal/models"
)

func (p *Postgres) UpdateOrderAndCreateAccrual(ctx context.Context, order models.Order) error {
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
	_, err = tx.ExecContext(ctx, sqlUpdateOrder, order.BonusAccrualStatus, order.OrderNumber)
	if err != nil {
		return err
	}

	sqlInsertAccrual := `
INSERT INTO accruals_calculated (user_uuid, order_number, accrued) 
VALUES ($1, $2, $3)
`
	if order.BonusAccrualStatus == models.BonusAccrualStatusProcessed {
		p.logger.Log.Debugln("storage inserting accrual with following args: ", order.UserUUID, order.OrderNumber, order.BonusesAccrued)
		_, err = tx.ExecContext(ctx, sqlInsertAccrual, order.UserUUID, order.OrderNumber, order.BonusesAccrued)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
