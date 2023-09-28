package middlewares

import (
	"net/http"

	"github.com/maxzhirnov/rewardify/internal/logger"
)

type Middlewares struct {
	logger *logger.Logger
}

func NewMiddlewares(l *logger.Logger) *Middlewares {
	return &Middlewares{logger: l}
}

func (m Middlewares) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.logger.Log.Info("Auth middleware not implemented")
		next.ServeHTTP(w, r)
	})
}
