package app

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/maxzhirnov/rewardify/internal/auth"
	"github.com/maxzhirnov/rewardify/internal/logger"
	"github.com/maxzhirnov/rewardify/internal/models"
	r "github.com/maxzhirnov/rewardify/internal/repo"
)

var (
	ErrUserAlreadyExist            = errors.New("user already exists")
	ErrInvalidOrderNumber          = errors.New("invalid order number")
	ErrAlreadyCreatedByUser        = errors.New("order already exist")
	ErrAlreadyCreatedByAnotherUser = errors.New("order already have been uploaded by another user")
	ErrInsufficientFunds           = errors.New("insufficient bonus balance")
)

type repo interface {
	InsertNewOrder(ctx context.Context, order models.Order) (bool, string, error)
	InsertNewWithdrawal(ctx context.Context, withdrawal models.Withdrawal) error
	GetUsersOrders(ctx context.Context, userUUID string) ([]models.Order, error)
	GetUsersWithdrawals(ctx context.Context, usrUUID string) ([]models.Withdrawal, error)
	GetUsersBalance(ctx context.Context, userUUID string) (models.UsersBalance, error)
	Bootstrap() error
	Ping(ctx context.Context) error
}

type authService interface {
	Register(ctx context.Context, username, password string) error
	Authenticate(ctx context.Context, username, password string) (string, error)
}

type accrualService interface {
	MonitorAndUpdateOrders(ctx context.Context, checkInterval time.Duration)
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

func (app *App) GetAllWithdrawals(ctx context.Context, usrUUID string) ([]models.Withdrawal, error) {
	return app.repo.GetUsersWithdrawals(ctx, usrUUID)
}

func (app *App) CreateWithdrawal(ctx context.Context, userUUID, orderNumber string, amount float32) error {
	// check if valid luhn number
	o := models.Order{OrderNumber: orderNumber}
	if !o.IsValidOrderNumber() {
		return ErrInvalidOrderNumber
	}

	// create and process withdrawal
	w := models.Withdrawal{
		UserUUID:    userUUID,
		OrderNumber: orderNumber,
		Amount:      amount,
		CreatedAt:   time.Now(),
	}

	err := app.repo.InsertNewWithdrawal(ctx, w)
	if errors.Is(err, r.ErrInsufficientFunds) {
		return ErrInsufficientFunds
	} else if err != nil {
		return err
	}
	return nil
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
