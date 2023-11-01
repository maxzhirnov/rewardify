package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/maxzhirnov/rewardify/internal/logger"
	"github.com/maxzhirnov/rewardify/internal/models"
)

const (
	cookieExpirationTime = 72 * time.Hour
)

type app interface {
	Register(ctx context.Context, username, password string) error
	Authenticate(ctx context.Context, username, password string) (string, error)
	UploadOrder(ctx context.Context, orderNumber, userUUID string) error
	CreateWithdrawal(ctx context.Context, userUUID, orderNumber string, amount float32) error
	GetAllOrders(ctx context.Context, userUUID string) ([]models.Order, error)
	GetAllWithdrawals(ctx context.Context, usrUUID string) ([]models.Withdrawal, error)
	GetBalance(ctx context.Context, userUUID string) (models.UsersBalance, error)
	Ping(ctx context.Context) error
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

func JSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}
}
