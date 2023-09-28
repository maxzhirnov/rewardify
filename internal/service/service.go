package service

import (
	"github.com/maxzhirnov/rewardify/internal/logger"
)

type repo interface {
}

type Service struct {
	repo   repo
	logger *logger.Logger
}

func NewService(r repo, l *logger.Logger) *Service {
	return &Service{
		repo:   r,
		logger: l,
	}
}

func (s *Service) Shutdown() {
	s.logger.Log.Info("Shutting down the application")
}
