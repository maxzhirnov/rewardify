package models

import (
	"strconv"
	"time"
)

type BonusAccrualStatus string

const (
	BonusAccrualStatusCreated        BonusAccrualStatus = "created"
	BonusAccrualStatusCalculating    BonusAccrualStatus = "calculating"
	BonusAccrualStatusAccrued        BonusAccrualStatus = "accrued"
	BonusAccrualStatusPartiallySpent BonusAccrualStatus = "partially spent"
	BonusAccrualStatusSpent          BonusAccrualStatus = "spent"
)

type Order struct {
	OrderNumber        string             `json:"order_number"`
	UserUUID           string             `json:"user_uuid"`
	BonusAccrualStatus BonusAccrualStatus `json:"bonus_accrual_status"`
	BonusesAccrued     float64            `json:"bonuses_accrued"`
	BonusesSpent       float64            `json:"bonuses_spent"`
	CreatedAt          time.Time          `json:"created_at"`
}

func (o Order) IsValidOrderNumber() bool {
	// Удаление пробелов и проверка, что строка не пустая
	inputLen := len(o.OrderNumber)
	if inputLen == 0 {
		return false
	}

	var sum int
	isSecondDigit := false

	// Обход строки справа налево
	for i := inputLen - 1; i >= 0; i-- {
		digit, err := strconv.Atoi(string(o.OrderNumber[i]))
		if err != nil {
			// Невалидный символ
			return false
		}

		if isSecondDigit {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		isSecondDigit = !isSecondDigit
	}

	// Число валидно, если сумма кратна 10
	return sum%10 == 0
}
