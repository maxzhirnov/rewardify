package repo

import (
	"context"

	"github.com/maxzhirnov/rewardify/internal/models"
)

func (p *Postgres) UpdateOrderAndCreateAccrual(ctx context.Context, order models.Order) error {
	tx, err := p.db.Begin()
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
		_, err = tx.ExecContext(ctx, sqlInsertAccrual, order.UserUUID, order.OrderNumber, order.BonusesAccrued)
		if err != nil {
			return err
		}

		// Проверка существования записи
		var exists bool
		checkExistence := `SELECT exists(SELECT 1 FROM balances WHERE user_uuid=$1)`
		err = tx.QueryRowContext(ctx, checkExistence, order.UserUUID).Scan(&exists)
		if err != nil {
			return err
		}

		if exists {
			sqlUpdateBalance := `
            UPDATE balances
            SET total_bonus = total_bonus + $1
            WHERE user_uuid = $2
            `
			_, err = tx.ExecContext(ctx, sqlUpdateBalance, order.BonusesAccrued, order.UserUUID)
			if err != nil {
				return err
			}
		} else {
			sqlInsertBalance := `
            INSERT INTO balances(user_uuid, total_bonus, redeemed_bonus)
            VALUES ($1, $2, 0)
            `
			_, err = tx.ExecContext(ctx, sqlInsertBalance, order.UserUUID, order.BonusesAccrued)
			if err != nil {
				return err
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
