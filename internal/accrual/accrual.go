package accrual

import (
	"context"
	"time"

	"github.com/maxzhirnov/rewardify/internal/logger"
	"github.com/maxzhirnov/rewardify/internal/models"
)

type repo interface {
	GetAllUnprocessedOrders(context.Context) ([]models.Order, error)
	UpdateOrderAndCreateAccrual(ctx context.Context, order models.Order) error
}

type api interface {
	Fetch(ctx context.Context, orderID string) (*APIResponse, error)
}

type orderProcessor interface {
	processOrder(ctx context.Context, order models.Order)
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
			s.processOrders(ctx)
		}
	}
}

func (s *Service) processOrders(ctx context.Context) {
	orders, err := s.repo.GetAllUnprocessedOrders(ctx)
	if err != nil {
		s.logger.Log.Error("Error fetching orders:", err)
		return
	}

	for _, order := range orders {
		go s.orderProcessor.processOrder(ctx, order)
	}
}
