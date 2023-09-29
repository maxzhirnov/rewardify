package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/maxzhirnov/rewardify/internal/models"
)

func (p *Postgres) CheckAndInsertOrder(order models.Order) (bool, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Hour)
	defer cancel()

	var existingUUID string
	err := p.DB.QueryRowContext(ctx, "SELECT user_uuid FROM orders WHERE order_number = $1", order.OrderNumber).Scan(&existingUUID)
	var errPgx = pgx.ErrNoRows
	if errors.As(err, &errPgx) {
		_, err := p.DB.ExecContext(ctx, "INSERT INTO orders (order_number, user_uuid, bonus_accrual_status, bonuses_accrued, bonuses_spent, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
			order.OrderNumber, order.UserUUID, order.BonusAccrualStatus, order.BonusesAccrued, order.BonusesSpent, order.CreatedAt)
		if err != nil {
			return false, "", err
		}
		return true, "", nil
	} else if err != nil {
		// Другая ошибка
		return false, "", err
	}
	// Заказ с таким номером уже существует
	if existingUUID == order.UserUUID {
		// Существующий заказ принадлежит тому же пользователю
		return false, existingUUID, nil
	} else {
		// Существующий заказ принадлежит другому пользователю
		return false, existingUUID, nil
	}
}
