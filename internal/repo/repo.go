package repo

import (
	"github.com/maxzhirnov/rewardify/internal/logger"
)

type store interface {
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
