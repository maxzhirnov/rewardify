package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/maxzhirnov/rewardify/internal/api"
	h "github.com/maxzhirnov/rewardify/internal/api/handlers"
	m "github.com/maxzhirnov/rewardify/internal/api/middlewares"
	l "github.com/maxzhirnov/rewardify/internal/logger"
	r "github.com/maxzhirnov/rewardify/internal/repo"
	s "github.com/maxzhirnov/rewardify/internal/service"
	st "github.com/maxzhirnov/rewardify/internal/store"
)

func main() {
	// Logging
	logger, err := l.NewLogger(logrus.DebugLevel)
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Close() // Закрыть файл куда пишутся логи

	// Инициализируем все зависимости и создаем сервис
	store := st.Postgres{}
	repo := r.NewRepo(store, logger)
	service := s.NewService(repo, logger)

	// Предусматриваем Gracefully shutdown
	go waitForShutdown(service)

	// Создаем Server со всеми его зависимостями
	handlers := h.NewHandlers(logger)
	middlewares := m.NewMiddlewares(logger)
	server := api.NewServer(service, handlers, middlewares, logger)

	// Запускаем сервер
	err = server.Start(":8080")
	if err != nil {
		return
	}
}

func waitForShutdown(s *s.Service) {
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)
	<-stopChan

	s.Shutdown()
	os.Exit(0)
}
