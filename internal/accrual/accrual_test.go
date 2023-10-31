package accrual

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/maxzhirnov/rewardify/internal/logger"
	"github.com/maxzhirnov/rewardify/internal/models"
)

type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) GetAllUnprocessedOrders(ctx context.Context) ([]models.Order, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.Order), args.Error(1)
}

func (m *MockRepo) UpdateOrderAndCreateAccrual(ctx context.Context, order models.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

type MockAPI struct {
	mock.Mock
}

func (m *MockAPI) Fetch(ctx context.Context, orderID string) (*APIResponse, error) {
	args := m.Called(ctx, orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*APIResponse), args.Error(1)
}

type MockOrderProcessor struct {
	mock.Mock
	called      bool
	calledTimes int
}

func (m *MockOrderProcessor) processOrder(ctx context.Context, order models.Order) (int, error) {
	m.called = true
	m.calledTimes++
	return 0, nil
}

func TestService_processOrder(t *testing.T) {
	repo := new(MockRepo)
	api := new(MockAPI)
	l, err := logger.NewLogger(logger.ErrorLevel, false)
	if err != nil {
		t.Fatal(err)
	}

	op := NewOrderProcessor(repo, api, l)

	// Setup expectations
	api.On("Fetch", mock.Anything, "123456").Return(&APIResponse{Status: string(models.BonusAccrualStatusProcessed), Accrual: 100}, nil)
	api.On("Fetch", mock.Anything, "666").Return(&APIResponse{Status: string(models.BonusAccrualStatusInvalid), Accrual: 0}, nil)
	api.On("Fetch", mock.Anything, "777").Return(nil, errTooManyRequests).Once()
	api.On("Fetch", mock.Anything, "777").Return(&APIResponse{Status: string(models.BonusAccrualStatusProcessed), Accrual: 100}, nil)
	api.On("Fetch", mock.Anything, "888").Return(nil, errNotRegistered)
	api.On("Fetch", mock.Anything, "999").Return(nil, errBadRequest)

	repo.On("UpdateOrderAndCreateAccrual", mock.Anything, models.Order{OrderNumber: "123456", BonusesAccrued: 100, BonusAccrualStatus: models.BonusAccrualStatusProcessed}).Return(nil)
	repo.On("UpdateOrderAndCreateAccrual", mock.Anything, models.Order{OrderNumber: "666", BonusesAccrued: 0, BonusAccrualStatus: models.BonusAccrualStatusInvalid}).Return(nil)
	repo.On("UpdateOrderAndCreateAccrual", mock.Anything, models.Order{OrderNumber: "777", BonusesAccrued: 100, BonusAccrualStatus: models.BonusAccrualStatusProcessed}).Return(nil)
	repo.On("UpdateOrderAndCreateAccrual", mock.Anything, models.Order{OrderNumber: "888", BonusesAccrued: 0, BonusAccrualStatus: models.BonusAccrualStatusInvalid}).Return(nil)

	type input struct {
		order models.Order
	}

	tests := []struct {
		name  string
		input input
	}{
		{
			name: "happy path",
			input: input{
				order: models.Order{OrderNumber: "123456"},
			},
		},
		{
			name: "invalid order",
			input: input{
				order: models.Order{OrderNumber: "666"},
			},
		},
		{
			name: "api err too many request",
			input: input{
				order: models.Order{OrderNumber: "777"},
			},
		},
		{
			name: "api err not registered",
			input: input{
				order: models.Order{OrderNumber: "888"},
			},
		},
		{
			name: "api err bad request",
			input: input{
				order: models.Order{OrderNumber: "999"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op.processOrder(context.TODO(), tt.input.order)
		})
	}

}

func TestService_MonitorAndUpdateOrders(t *testing.T) {
	repo := new(MockRepo)
	op := new(MockOrderProcessor)
	l, err := logger.NewLogger(logger.ErrorLevel, false)
	if err != nil {
		t.Fatal(err)
	}

	mockOrders := []models.Order{
		{OrderNumber: "1"},
		{OrderNumber: "2"},
	}

	repo.On("GetAllUnprocessedOrders", mock.Anything).Return(mockOrders, nil)

	mockService := NewService(repo, op, l)

	// Создание контекста с отменой
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel() // Отменить контекст после завершения теста

	// Запуск метода в горутине
	go mockService.MonitorAndUpdateOrders(ctx, 90*time.Millisecond)

	// Ожидание достаточно времени, чтобы ticker сработал
	time.Sleep(110 * time.Millisecond)

	// Проверка, что processOrders был вызван
	assert.True(t, op.called)
	assert.Equal(t, 2, op.calledTimes)
}
