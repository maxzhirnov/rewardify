package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/maxzhirnov/rewardify/internal/api/handlers"
	"github.com/maxzhirnov/rewardify/internal/api/middlewares"
	"github.com/maxzhirnov/rewardify/internal/logger"
	"github.com/maxzhirnov/rewardify/internal/service"
)

type Server struct {
	service     *service.Service
	handlers    *handlers.Handlers
	middlewares *middlewares.Middlewares
	r           *chi.Mux
	logger      *logger.Logger
}

func NewServer(s *service.Service, h *handlers.Handlers, m *middlewares.Middlewares, l *logger.Logger) *Server {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	return &Server{
		service:     s,
		handlers:    h,
		middlewares: m,
		r:           r,
		logger:      l,
	}
}

func (api *Server) Start(addr string) error {
	api.logger.Log.Info("Starting Server server on: ", addr)
	api.r.Get("/ping", api.handlers.HandlePing)

	api.r.Route("/api/user", func(r chi.Router) {
		r.Use()
		r.Post("/register", api.handlers.HandleRegister)
		r.Post("/login", api.handlers.HandleLogin)
		r.Post("/orders", api.handlers.HandleOrdersUpload)
		r.Get("/orders", api.handlers.HandleOrdersList)
		r.Get("/balance", api.handlers.HandleGetBalance)
		r.Post("/balance/withdraw", api.handlers.HandleWithdraw)
		r.Get("/withdrawals", api.handlers.HandleListAllWithdraws)
	})

	err := http.ListenAndServe(addr, api.r)
	if err != nil {
		return err
	}
	return nil
}
