package store

import (
	"context"

	"github.com/maxzhirnov/rewardify/internal/models"
)

func (p *Postgres) InsertNewWithdrawal(ctx context.Context, withdrawal models.Withdrawal) error {
	tx, err := p.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var uBalance models.UsersBalance

	sqlGetUserBalance := `
SELECT total_bonus, redeemed_bonus, (total_bonus - redeemed_bonus) as current from balances WHERE user_uuid=$1
`
	row := tx.QueryRowContext(ctx, sqlGetUserBalance, withdrawal.UserUUID)

	err = row.Scan(&uBalance.Earned, &uBalance.Withdrawn, &uBalance.Current)
	if err != nil {
		return err
	}

	// If not enough bonuses on balance
	if uBalance.Current < withdrawal.Amount {
		return ErrInsufficientFunds
	}

	sqlInsertWithdrawal := `
INSERT INTO withdrawals (user_uuid, order_number, withdrew, created_at) VALUES ($1, $2, $3, now())
`
	_, err = tx.ExecContext(ctx, sqlInsertWithdrawal, withdrawal.UserUUID, withdrawal.OrderNumber, withdrawal.Amount)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
