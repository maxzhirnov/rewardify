package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/maxzhirnov/rewardify/internal/api/handlers"
	mw "github.com/maxzhirnov/rewardify/internal/api/middlewares"
	"github.com/maxzhirnov/rewardify/internal/logger"
)

type Server struct {
	handlers    *handlers.Handlers
	middlewares *mw.Middlewares
	r           *chi.Mux
	logger      *logger.Logger
}

func NewServer(h *handlers.Handlers, m *mw.Middlewares, l *logger.Logger) *Server {
	r := chi.NewRouter()

	// Встроенные middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Собственные middleware
	r.Use(mw.GzipMiddleware)

	return &Server{
		handlers:    h,
		middlewares: m,
		r:           r,
		logger:      l,
	}
}

func (api *Server) Start(addr string) error {
	api.logger.Log.Info("Starting Server server on: ", addr)

	// Routes
	api.r.Get("/ping", api.handlers.HandlePing)
	api.r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", api.handlers.HandleRegister)
		r.Post("/login", api.handlers.HandleLogin)
		// Группа доступная только auth пользователям
		r.Group(func(r chi.Router) {
			r.Use(api.middlewares.AuthMiddleware)
			r.Get("/ping", api.handlers.HandlePing)
			r.Post("/orders", api.handlers.HandleUploadOrder)
			r.Get("/orders", api.handlers.HandleGetOrders)
			r.Get("/balance", api.handlers.HandleGetBalance)
			r.Post("/balance/withdraw", api.handlers.HandleWithdraw)
			r.Get("/withdrawals", api.handlers.HandleGetWithdrawals)
		})
	})

	err := http.ListenAndServe(addr, api.r)
	if err != nil {
		return err
	}
	return nil
}
