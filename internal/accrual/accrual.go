package accrual

import (
	"context"
	"errors"
	"time"

	"github.com/maxzhirnov/rewardify/internal/logger"
	"github.com/maxzhirnov/rewardify/internal/models"
)

const (
	monitorAndUpdateInterval       = 10 * time.Second
	resubmitRequestForStatusUpdate = 10 * time.Second
)

type repo interface {
	GetAllUnprocessedOrders(context.Context) ([]models.Order, error)
	UpdateOrderAndCreateAccrual(ctx context.Context, order models.Order, newStatus string) error
}

type api interface {
	Fetch(ctx context.Context, orderID string) (*APIResponse, error)
}

type Service struct {
	repo       repo
	apiWrapper api
	logger     *logger.Logger
}

func NewService(repo repo, apiWrapper api, logger *logger.Logger) *Service {
	return &Service{
		repo:       repo,
		apiWrapper: apiWrapper,
		logger:     logger,
	}
}

func (s *Service) MonitorAndUpdateOrders(ctx context.Context) {
	ticker := time.NewTicker(monitorAndUpdateInterval)
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
		go s.processOrder(ctx, order)
	}
}

func (s *Service) processOrder(ctx context.Context, order models.Order) {
	for {
		response, err := s.apiWrapper.Fetch(ctx, order.OrderNumber)
		if errors.Is(err, errTooManyRequests) {
			time.Sleep(resubmitRequestForStatusUpdate)
			continue
		} else if errors.Is(err, errNotRegistered) {
			err := s.repo.UpdateOrderAndCreateAccrual(ctx, order, string(models.BonusAccrualStatusInvalid))
			if err != nil {
				s.logger.Log.Error("Error updating response:", err)
			}

			return
		} else if errors.Is(err, errBadRequest) {

		} else if err != nil {
			s.logger.Log.Error("Error fetching response:", err)
			return
		}

		if response.Status == "PROCESSED" || response.Status == "INVALID" {
			order.BonusesAccrued = response.Accrual
			err := s.repo.UpdateOrderAndCreateAccrual(ctx, order, response.Status)
			if err != nil {
				s.logger.Log.Error("Error updating response:", err)
			}
			return
		}

		time.Sleep(resubmitRequestForStatusUpdate)
	}
}
