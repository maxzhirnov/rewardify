package middlewares

import (
	"github.com/maxzhirnov/rewardify/internal/logger"
	"github.com/maxzhirnov/rewardify/internal/models"
)

type authService interface {
	ValidateToken(tokenString string) (models.User, error)
}

type Middlewares struct {
	authService authService
	logger      *logger.Logger
}

func NewMiddlewares(authService authService, l *logger.Logger) *Middlewares {
	return &Middlewares{
		authService: authService,
		logger:      l,
	}
}
