package accrual

import (
	"context"
	"time"

	"golang.org/x/sync/errgroup"

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
	processOrder(ctx context.Context, order models.Order) (int, error)
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
			ticker.Reset(checkInterval)
			if retryAfter, err := s.processOrders(ctx); err != nil {
				// Если есть ошибка, увеличиваем интервал тикера
				s.logger.Log.Errorln("Error processing orders:", err)
				retryAfterDuration := time.Duration(retryAfter) * time.Second
				newInterval := checkInterval + retryAfterDuration
				ticker.Reset(newInterval)
			}
		}
	}
}

func (s *Service) processOrders(ctx context.Context) (int, error) {
	retryAfter := retryAfterDefault
	orders, err := s.repo.GetAllUnprocessedOrders(ctx)
	if err != nil {
		s.logger.Log.Error("Error fetching orders:", err)
		return retryAfter, err
	}

	g, ctx := errgroup.WithContext(ctx)

	for _, order := range orders {
		order := order
		g.Go(func() error {
			retryAfter, err = s.orderProcessor.processOrder(ctx, order)
			return err
		})
	}

	if err := g.Wait(); err != nil {
		s.logger.Log.Error("Error processing order:", err)
		return retryAfter, err
	}

	return retryAfter, nil
}
