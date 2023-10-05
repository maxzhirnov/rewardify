package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/maxzhirnov/rewardify/internal/models"
)

// InsertNewOrder inserts new order into database if order with specified OrderNumber already exists it returns false and error
func (p *Postgres) InsertNewOrder(ctx context.Context, order models.Order) (bool, string, error) {
	var existingUUID string
	err := p.DB.QueryRowContext(ctx, "SELECT user_uuid FROM orders WHERE order_number = $1", order.OrderNumber).Scan(&existingUUID)
	if errors.Is(err, sql.ErrNoRows) {
		_, err := p.DB.ExecContext(ctx, "INSERT INTO orders (order_number, user_uuid, bonus_accrual_status, created_at) VALUES ($1, $2, $3, $4)",
			order.OrderNumber, order.UserUUID, order.BonusAccrualStatus, order.CreatedAt)
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
