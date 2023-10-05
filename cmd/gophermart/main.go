package main

import (
	"context"
	"log"
	"net/http"
	"os"
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

	// Предусматриваем Gracefully shutdown
	go waitForShutdown(appInstance)

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

func waitForShutdown(s *app.App) {
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)
	<-stopChan

	s.Shutdown()
	os.Exit(0)
}
