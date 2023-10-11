package repo

import (
	"context"
	"errors"

	"github.com/maxzhirnov/rewardify/internal/logger"
	"github.com/maxzhirnov/rewardify/internal/models"
	s "github.com/maxzhirnov/rewardify/internal/store"
)

var (
	ErrUserAlreadyExist  = errors.New("user already exists")
	ErrInsufficientFunds = errors.New("insufficient bonus balance")
)

type store interface {
	Ping(ctx context.Context) error
	Bootstrap() error
	GetUserByUsername(ctx context.Context, username string) (models.User, error)
	InsertNewUser(ctx context.Context, user models.User) error
	InsertNewOrder(ctx context.Context, order models.Order) (bool, string, error)
	InsertNewWithdrawal(ctx context.Context, withdrawal models.Withdrawal) error
	GetUsersOrders(ctx context.Context, userUUID string) ([]models.Order, error)
	GetUsersBalance(ctx context.Context, userUUID string) (models.UsersBalance, error)
	GetAllUnprocessedOrders(ctx context.Context) ([]models.Order, error)
	UpdateOrderAndCreateAccrual(ctx context.Context, order models.Order) error
	GetUsersWithdrawals(ctx context.Context, usrUUID string) ([]models.Withdrawal, error)
}

type Repo struct {
	store  store
	logger *logger.Logger
}

func NewRepo(s store, l *logger.Logger) *Repo {
	return &Repo{
		store:  s,
		logger: l,
	}
}

func (r *Repo) GetUserByUsername(ctx context.Context, username string) (models.User, error) {
	return r.store.GetUserByUsername(ctx, username)
}

func (r *Repo) InsertNewUser(ctx context.Context, user models.User) error {
	err := r.store.InsertNewUser(ctx, user)
	if errors.Is(err, s.ErrUserAlreadyExist) {
		return ErrUserAlreadyExist
	}

	return err
}

func (r *Repo) InsertNewOrder(ctx context.Context, order models.Order) (bool, string, error) {
	return r.store.InsertNewOrder(ctx, order)
}

func (r *Repo) InsertNewWithdrawal(ctx context.Context, withdrawal models.Withdrawal) error {
	err := r.store.InsertNewWithdrawal(ctx, withdrawal)
	if errors.Is(err, s.ErrInsufficientFunds) {
		return ErrInsufficientFunds
	} else if err != nil {
		return err
	}
	return nil
}

func (r *Repo) GetUsersOrders(ctx context.Context, userUUID string) ([]models.Order, error) {
	return r.store.GetUsersOrders(ctx, userUUID)
}

func (r *Repo) GetUsersWithdrawals(ctx context.Context, usrUUID string) ([]models.Withdrawal, error) {
	return r.store.GetUsersWithdrawals(ctx, usrUUID)
}

func (r *Repo) GetUsersBalance(ctx context.Context, userUUID string) (models.UsersBalance, error) {
	return r.store.GetUsersBalance(ctx, userUUID)
}

func (r *Repo) Bootstrap(ctx context.Context) error {
	return r.store.Bootstrap()
}

func (r *Repo) Ping(ctx context.Context) error {
	return r.store.Ping(ctx)
}

func (r *Repo) GetAllUnprocessedOrders(ctx context.Context) ([]models.Order, error) {
	return r.store.GetAllUnprocessedOrders(ctx)
}
func (r *Repo) UpdateOrderAndCreateAccrual(ctx context.Context, order models.Order) error {
	return r.store.UpdateOrderAndCreateAccrual(ctx, order)
}
