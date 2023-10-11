package accrual

import (
	"context"
	"errors"
	"time"

	"github.com/maxzhirnov/rewardify/internal/logger"
	"github.com/maxzhirnov/rewardify/internal/models"
)

const resubmitRequestForStatusUpdate = 2 * time.Second

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

func (p *OrderProcessor) processOrder(ctx context.Context, order models.Order) {
	p.logger.Log.Debug("Starting to process order: ", order.OrderNumber)
	for {
		response, err := p.apiWrapper.Fetch(ctx, order.OrderNumber)
		if errors.Is(err, errTooManyRequests) {
			p.logger.Log.Error(err)
			time.Sleep(resubmitRequestForStatusUpdate)
			continue
		} else if errors.Is(err, errNotRegistered) {
			p.logger.Log.Error(err)
			order.BonusAccrualStatus = models.BonusAccrualStatusInvalid
			err := p.repo.UpdateOrderAndCreateAccrual(ctx, order)
			if err != nil {
				p.logger.Log.Error("Error updating response:", err)
			}
			return
		} else if errors.Is(err, errBadRequest) {
			p.logger.Log.Error(err)
			return
		} else if err != nil {
			p.logger.Log.Error("Error fetching response:", err)
			return
		}

		if response.Status == "PROCESSED" || response.Status == "INVALID" {
			switch response.Status {
			case "PROCESSED":
				order.BonusAccrualStatus = models.BonusAccrualStatusProcessed
			case "INVALID":
				order.BonusAccrualStatus = models.BonusAccrualStatusInvalid
			}
			order.BonusesAccrued = response.Accrual
			p.logger.Log.Debugln("processOrder sending order to update", order)
			err := p.repo.UpdateOrderAndCreateAccrual(ctx, order)
			if err != nil {
				p.logger.Log.Error("Error updating response:", err)
			}
			return
		}
	}
}
