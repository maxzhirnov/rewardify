package handlers

import (
	"time"

	"github.com/maxzhirnov/rewardify/internal/logger"
)

const (
	cookieExpirationTime = 72 * time.Hour
)

type app interface {
	Register(username, password string) error
	Authenticate(username, password string) (string, error)
	UploadOrder(orderNumber, userUUID string) error
	Ping() error
}

type Handlers struct {
	app    app
	logger *logger.Logger
}

func NewHandlers(app app, l *logger.Logger) *Handlers {
	return &Handlers{
		app:    app,
		logger: l,
	}
}
