package app

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/maxzhirnov/rewardify/internal/auth"
	"github.com/maxzhirnov/rewardify/internal/logger"
	"github.com/maxzhirnov/rewardify/internal/models"
)

var (
	ErrUserAlreadyExist            = errors.New("user already exists")
	ErrInvalidOrderNumber          = errors.New("invalid order number")
	ErrAlreadyCreatedByUser        = errors.New("order already exist")
	ErrAlreadyCreatedByAnotherUser = errors.New("order already have been uploaded by another user")
)

type repo interface {
	InsertNewOrder(ctx context.Context, order models.Order) (bool, string, error)
	GetUsersOrders(ctx context.Context, userUUID string) ([]models.Order, error)
	GetUsersBalance(ctx context.Context, userUUID string) (models.UsersBalance, error)
	Bootstrap(ctx context.Context) error
	Ping(ctx context.Context) error
}

type authService interface {
	Register(ctx context.Context, username, password string) error
	Authenticate(ctx context.Context, username, password string) (string, error)
}

type accrualService interface {
	MonitorAndUpdateOrders(ctx context.Context)
}

type App struct {
	authService    authService
	accrualService accrualService
	repo           repo
	logger         *logger.Logger
}

func NewApp(auth authService, accrual accrualService, repo repo, l *logger.Logger) *App {
	return &App{
		authService:    auth,
		accrualService: accrual,
		repo:           repo,
		logger:         l,
	}
}

func (app *App) StartAccrualUpdater(ctx context.Context) {
	app.accrualService.MonitorAndUpdateOrders(ctx)
}

func (app *App) Register(ctx context.Context, username, password string) error {
	err := app.authService.Register(ctx, username, password)
	if errors.Is(err, auth.ErrUserAlreadyExist) {
		return ErrUserAlreadyExist
	}
	return err
}

func (app *App) Authenticate(ctx context.Context, username, password string) (string, error) {
	return app.authService.Authenticate(ctx, username, password)
}

func (app *App) UploadOrder(ctx context.Context, orderNumber, userUUID string) error {
	order := models.Order{
		OrderNumber:        orderNumber,
		UserUUID:           userUUID,
		BonusAccrualStatus: models.BonusAccrualStatusNew,
		BonusesAccrued:     0,
		CreatedAt:          time.Now(),
	}

	if !order.IsValidOrderNumber() {
		return ErrInvalidOrderNumber
	}

	isInserted, insertedUserUUID, err := app.repo.InsertNewOrder(ctx, order)
	if !isInserted {
		if insertedUserUUID == userUUID {
			return ErrAlreadyCreatedByUser
		} else {
			return ErrAlreadyCreatedByAnotherUser
		}
	}
	if err != nil {
		return err
	}

	return nil
}

func (app *App) GetAllOrders(ctx context.Context, userUUID string) ([]models.Order, error) {
	return app.repo.GetUsersOrders(ctx, userUUID)
}

func (app *App) GetBalance(ctx context.Context, userUUID string) (models.UsersBalance, error) {
	return app.repo.GetUsersBalance(ctx, userUUID)
}

func (app *App) Ping(ctx context.Context) error {
	return app.repo.Ping(ctx)
}

func (app *App) WaitForShutdown(ctx context.Context) {
	<-ctx.Done()
	app.shutdown()

}

func (app *App) shutdown() {
	app.logger.Log.Info("Shutting down the application")
	os.Exit(0)
}
