package models

import (
	"time"
)

type User struct {
	UUID     string
	Username string
	Password string
	CreateAt time.Time
}
