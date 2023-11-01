package models

import (
	"time"
)

type Withdrawal struct {
	UserUUID    string
	OrderNumber string
	Amount      float32
	CreatedAt   time.Time
}
