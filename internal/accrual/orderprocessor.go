package accrual

import (
	"context"
	"errors"

	"github.com/maxzhirnov/rewardify/internal/logger"
	"github.com/maxzhirnov/rewardify/internal/models"
)

type OrderProcessor struct {
	repo       repo
	apiWrapper api
	logger     *logger.Logger
}

func NewOrderProcessor(repo repo, apiWrapper api, l *logger.Logger) *OrderProcessor {
	return &OrderProcessor{
		repo:       repo,
		apiWrapper: apiWrapper,
		logger:     l,
	}
}

func (p *OrderProcessor) processOrder(ctx context.Context, order models.Order) (int, error) {
	p.logger.Log.Debug("Starting to process order: ", order.OrderNumber)
	response, err := p.apiWrapper.Fetch(ctx, order.OrderNumber)
	retryAfter := 0
	if response != nil {
		retryAfter = response.RetryAfter
	}

	if errors.Is(err, errTooManyRequests) {
		p.logger.Log.Error(err)
		return retryAfter, err
	}

	if errors.Is(err, errNotRegistered) {
		p.logger.Log.Error(err)
		order.BonusAccrualStatus = models.BonusAccrualStatusInvalid
		err := p.repo.UpdateOrderAndCreateAccrual(ctx, order)
		if err != nil {
			p.logger.Log.Error("Error updating response:", err)
		}
		return retryAfter, nil
	}

	if err != nil {
		p.logger.Log.Errorf("Error fetching response: %s", err)
		return retryAfter, err
	}

	if response.Status != models.BonusAccrualStatusProcessed.String() && response.Status != models.BonusAccrualStatusInvalid.String() {
		p.logger.Log.Info("Accrual status not final")
		return retryAfter, nil
	}

	order.BonusAccrualStatus = models.BonusAccrualStatusInvalid
	if response.Status == models.BonusAccrualStatusProcessed.String() {
		order.BonusAccrualStatus = models.BonusAccrualStatusProcessed
		order.BonusesAccrued = response.Accrual
	}
	p.logger.Log.Debugln("processOrder sending order to update", order)
	err = p.repo.UpdateOrderAndCreateAccrual(ctx, order)
	if err != nil {
		p.logger.Log.Error("Error updating response:", err)
	}
	return retryAfter, nil
}
