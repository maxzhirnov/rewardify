package repo

import (
	"errors"

	"github.com/maxzhirnov/rewardify/internal/logger"
	"github.com/maxzhirnov/rewardify/internal/models"
	s "github.com/maxzhirnov/rewardify/internal/store"
)

var (
	ErrUserAlreadyExist = errors.New("user already exists")
)

type store interface {
	Ping() error
	Bootstrap() error
	GetUserByUsername(username string) (models.User, error)
	CreateUser(user models.User) error
	CheckAndInsertOrder(order models.Order) (bool, string, error)
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

func (r *Repo) GetUserByUsername(username string) (models.User, error) {
	return r.store.GetUserByUsername(username)
}

func (r *Repo) CreateUser(user models.User) error {
	err := r.store.CreateUser(user)
	if errors.Is(err, s.ErrUserAlreadyExist) {
		return ErrUserAlreadyExist
	}

	return err
}

func (r *Repo) CheckAndInsertOrder(order models.Order) (bool, string, error) {
	return r.store.CheckAndInsertOrder(order)
}

func (r *Repo) Bootstrap() error {
	return r.store.Bootstrap()
}

func (r *Repo) Ping() error {
	return r.store.Ping()
}
