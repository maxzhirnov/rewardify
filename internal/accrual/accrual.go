package accrual

import (
	"context"
	"time"

	"github.com/maxzhirnov/rewardify/internal/logger"
	"github.com/maxzhirnov/rewardify/internal/models"
)

const errorDelay = 5 * time.Second

type repo interface {
	GetAllUnprocessedOrders(context.Context) ([]models.Order, error)
	UpdateOrderAndCreateAccrual(ctx context.Context, order models.Order) error
}

type api interface {
	Fetch(ctx context.Context, orderID string) (*APIResponse, error)
}

type orderProcessor interface {
	processOrder(ctx context.Context, order models.Order) error
}

type Service struct {
	repo           repo
	orderProcessor orderProcessor
	logger         *logger.Logger
}

func NewService(repo repo, orderProcessor orderProcessor, logger *logger.Logger) *Service {
	return &Service{
		repo:           repo,
		orderProcessor: orderProcessor,
		logger:         logger,
	}
}

func (s *Service) MonitorAndUpdateOrders(ctx context.Context, checkInterval time.Duration) {
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.processOrders(ctx); err != nil {
				// Если есть ошибка, увеличиваем интервал тикера
				s.logger.Log.Errorln("Error processing orders:", err)
				newInterval := checkInterval + errorDelay
				ticker.Reset(newInterval)
			} else {
				// В случае успеха, вернем интервал тикера к исходному значению
				ticker.Reset(checkInterval)
			}
		}
	}
}

func (s *Service) processOrders(ctx context.Context) error {
	orders, err := s.repo.GetAllUnprocessedOrders(ctx)
	if err != nil {
		s.logger.Log.Error("Error fetching orders:", err)
		return err
	}

	errCh := make(chan error, len(orders))

	for _, order := range orders {
		go func(ord models.Order) { // предположим, что Order — это тип вашего заказа
			if err := s.orderProcessor.processOrder(ctx, ord); err != nil {
				errCh <- err
			} else {
				errCh <- nil
			}
		}(order)
	}

	// Ждем завершения всех горутин и собираем ошибки
	for i := 0; i < len(orders); i++ {
		if err := <-errCh; err != nil {
			s.logger.Log.Error("Error processing order:", err)
			return err
		}
	}

	close(errCh)

	return nil
}
