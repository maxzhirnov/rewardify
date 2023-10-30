package repo

import (
	"context"

	"github.com/maxzhirnov/rewardify/internal/models"
)

func (p *Postgres) InsertNewWithdrawal(ctx context.Context, withdrawal models.Withdrawal) error {
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Пытаемся обновить баланс пользователя и проверяем что он не опускается ниже нуля
	sqlUpdateBalance := `
 UPDATE balances
 SET redeemed_bonus = redeemed_bonus + $1
 WHERE user_uuid = $2
 RETURNING (total_bonus - redeemed_bonus)
 `
	var newCurrentBalance int
	err = tx.QueryRowContext(ctx, sqlUpdateBalance, withdrawal.Amount, withdrawal.UserUUID).Scan(&newCurrentBalance)
	if err != nil {
		return err
	}

	// Проверяем что новый баланс не ниже нул
	if newCurrentBalance < 0 {
		return ErrInsufficientFunds
	}

	// Вставляем новое списание
	sqlInsertWithdrawal := `
 INSERT INTO withdrawals (user_uuid, order_number, withdrew, created_at) VALUES ($1, $2, $3, now())
 `
	_, err = tx.ExecContext(ctx, sqlInsertWithdrawal, withdrawal.UserUUID, withdrawal.OrderNumber, withdrawal.Amount)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
