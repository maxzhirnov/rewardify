package app

import (
	"errors"
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
	CheckAndInsertOrder(order models.Order) (bool, string, error)
	Bootstrap() error
	Ping() error
}

type authService interface {
	Register(username, password string) error
	Authenticate(username, password string) (string, error)
}

type App struct {
	authService authService
	repo        repo
	logger      *logger.Logger
}

func NewApp(a authService, r repo, l *logger.Logger) *App {
	return &App{
		authService: a,
		repo:        r,
		logger:      l,
	}
}

func (app *App) Register(username, password string) error {
	err := app.authService.Register(username, password)
	if errors.Is(err, auth.ErrUserAlreadyExist) {
		return ErrUserAlreadyExist
	}
	return err
}

func (app *App) Authenticate(username, password string) (string, error) {
	return app.authService.Authenticate(username, password)
}

func (app *App) UploadOrder(orderNumber, userUUID string) error {
	order := models.Order{
		OrderNumber:        orderNumber,
		UserUUID:           userUUID,
		BonusAccrualStatus: models.BonusAccrualStatusCreated,
		BonusesAccrued:     0,
		BonusesSpent:       0,
		CreatedAt:          time.Now(),
	}

	if !order.IsValidOrderNumber() {
		return ErrInvalidOrderNumber
	}

	isInserted, insertedUserUUID, err := app.repo.CheckAndInsertOrder(order)
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

func (app *App) Ping() error {
	return app.repo.Ping()
}

func (app *App) Shutdown() {
	app.logger.Log.Info("Shutting down the application")
}
