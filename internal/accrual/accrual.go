package accrual

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
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

func (s *Service) RunBinary(a, d string) error {
	s.logger.Log.Debugf("Starting accrual binary with params a: %s, d: %s", a, d)
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd/accrual/accrual_windows_amd64", "-a", "-a", a, "-d", d)
	case "linux":
		cmd = exec.Command("cmd/accrual/accrual_linux_amd64", "-a", "-a", a, "-d", d)
	case "darwin":
		switch runtime.GOARCH {
		case "arm64":
			cmd = exec.Command("cmd/accrual/accrual_darwin_arm64", "-a", a, "-d", d)
			s.logger.Log.Debug(cmd)
		case "amd64":
			cmd = exec.Command("cmd/accrual/accrual_darwin_amd64", "-a", a, "-d", d)
		}
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
	// Запуск команды
	o, err := cmd.CombinedOutput()
	//err := cmd.Start()
	if err != nil {
		return err
	}

	s.logger.Log.Warn(string(o))
	return nil
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
