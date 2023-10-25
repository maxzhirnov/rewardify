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

	var uBalance models.UsersBalance

	// Получаем текущий баланс пользователя
	sqlGetUserBalance := `
    SELECT total_bonus, redeemed_bonus, (total_bonus - redeemed_bonus) as current from balances WHERE user_uuid=$1 FOR UPDATE
    `
	row := tx.QueryRowContext(ctx, sqlGetUserBalance, withdrawal.UserUUID)
	err = row.Scan(&uBalance.Earned, &uBalance.Withdrawn, &uBalance.Current)
	if err != nil {
		return err
	}

	// Проверяем, достаточно ли бонусов на балансе
	if uBalance.Current < withdrawal.Amount {
		return ErrInsufficientFunds
	}

	// Вставляем новый вывод
	sqlInsertWithdrawal := `
    INSERT INTO withdrawals (user_uuid, order_number, withdrew, created_at) VALUES ($1, $2, $3, now())
    `
	_, err = tx.ExecContext(ctx, sqlInsertWithdrawal, withdrawal.UserUUID, withdrawal.OrderNumber, withdrawal.Amount)
	if err != nil {
		return err
	}

	// Обновляем redeemed_bonus для пользователя
	sqlUpdateBalance := `
    UPDATE balances
    SET redeemed_bonus = redeemed_bonus + $1
    WHERE user_uuid = $2
    `
	_, err = tx.ExecContext(ctx, sqlUpdateBalance, withdrawal.Amount, withdrawal.UserUUID)
	if err != nil {
		return err
	}

	// Проверка существования записи
	var exists bool
	checkExistence := `SELECT exists(SELECT 1 FROM balances WHERE user_uuid=$1) FOR UPDATE`
	err = tx.QueryRowContext(ctx, checkExistence, withdrawal.UserUUID).Scan(&exists)
	if err != nil {
		return err
	}

	if !exists {
		sqlInsertBalance := `
        INSERT INTO balances(user_uuid, total_bonus, redeemed_bonus)
        VALUES ($1, 0, $2)
        `
		_, err = tx.ExecContext(ctx, sqlInsertBalance, withdrawal.UserUUID, withdrawal.Amount)
		if err != nil {
			return err
		}
	}

	// Подтверждаем транзакцию
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
