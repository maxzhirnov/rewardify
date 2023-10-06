package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/maxzhirnov/rewardify/internal/accrual"
	"github.com/maxzhirnov/rewardify/internal/api"
	"github.com/maxzhirnov/rewardify/internal/api/handlers"
	"github.com/maxzhirnov/rewardify/internal/api/middlewares"
	"github.com/maxzhirnov/rewardify/internal/app"
	"github.com/maxzhirnov/rewardify/internal/auth"
	"github.com/maxzhirnov/rewardify/internal/config"
	l "github.com/maxzhirnov/rewardify/internal/logger"
	"github.com/maxzhirnov/rewardify/internal/repo"
	"github.com/maxzhirnov/rewardify/internal/store"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	// Logging
	logger, err := l.NewLogger(l.DebugLevel, true)
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Close() // Закрыть файл куда пишутся логи

	// Config
	cfg := config.NewFromFlagsOrEnv(logger)
	logger.Log.Debug(cfg)

	// Инициализируем все зависимости и создаем сервис
	storage, err := store.NewPostgres(cfg.DatabaseURI(), logger)
	if err != nil {
		logger.Log.Fatal(err)
	}

	repository := repo.NewRepo(storage, logger)
	err = repository.Bootstrap(context.TODO())
	if err != nil {
		logger.Log.Fatal(err)
	}

	authService := auth.NewAuthService(repository, cfg.AuthSecretKey())

	httpClient := &http.Client{}
	accrualAPIWrapper := accrual.NewAPIWrapper(cfg.AccrualSystemAddress(), httpClient, logger)
	accrualService := accrual.NewService(repository, accrualAPIWrapper, logger)
	appInstance := app.NewApp(authService, accrualService, repository, logger)

	go appInstance.StartAccrualService(ctx)
	go appInstance.WaitForShutdown(ctx)

	// Создаем Server со всеми его зависимостями
	hd := handlers.NewHandlers(appInstance, logger)
	mw := middlewares.NewMiddlewares(authService, logger)
	server := api.NewServer(hd, mw, logger)

	// Запускаем сервер
	err = server.Start(cfg.RunAddress())
	if err != nil {
		return
	}
}
