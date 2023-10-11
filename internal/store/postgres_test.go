package store

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"github.com/maxzhirnov/rewardify/internal/logger"
	"github.com/maxzhirnov/rewardify/internal/models"
)

func TestPostgres_GetUserByUsername(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	l, _ := logger.NewLogger(logger.DebugLevel, false)

	p := &Postgres{
		DB:     db,
		logger: l,
	}

	mockUser := models.User{
		UUID:     "some-uuid",
		Username: "username",
		Password: "password",
		CreateAt: time.Now(),
	}

	rows := sqlmock.NewRows([]string{"uuid", "username", "password", "created_at"}).
		AddRow(mockUser.UUID, mockUser.Username, mockUser.Password, mockUser.CreateAt)
	sql := `SELECT uuid, username, password, created_at FROM users WHERE username=$1`
	mock.ExpectQuery(regexp.QuoteMeta(sql)).
		WithArgs("username").
		WillReturnRows(rows)

	// Вызов функции, которую мы тестируем
	user, err := p.GetUserByUsername(context.Background(), "username")

	// Проверки
	assert.Nil(t, err)
	assert.Equal(t, mockUser, user)

	// Проверка, что все ожидания mock были выполнены
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
func TestPostgres_GetUsersBalance(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	l, _ := logger.NewLogger(logger.DebugLevel, false)

	p := &Postgres{
		DB:     db,
		logger: l,
	}

	mockBalance := models.UsersBalance{
		UserUUID:  "123",
		Earned:    210,
		Withdrawn: 45.5,
		Current:   169.5,
	}

	rows := sqlmock.NewRows([]string{"total_bonus", "redeemed_bonus", "current"}).
		AddRow(mockBalance.Earned, mockBalance.Withdrawn, mockBalance.Current)

	sql := `
SELECT total_bonus, redeemed_bonus, (total_bonus - redeemed_bonus) as current from balances WHERE user_uuid=$1
`
	mock.ExpectQuery(regexp.QuoteMeta(sql)).
		WithArgs("123").
		WillReturnRows(rows)

	// Вызов функции, которую мы тестируем
	balance, err := p.GetUsersBalance(context.Background(), "123")

	// Проверки
	assert.Nil(t, err)
	assert.Equal(t, mockBalance, balance)

	// Проверка, что все ожидания mock были выполнены
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
func TestPostgres_GetAllUnprocessedOrders(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	l, _ := logger.NewLogger(logger.DebugLevel, false)

	p := &Postgres{
		DB:     db,
		logger: l,
	}

	mockDate := time.Date(0001, 01, 01, 00, 00, 00, 00, time.UTC)

	rows := sqlmock.NewRows([]string{"order_number", "user_uuid", "bonus_accrual_status", "created_at"}).
		AddRow("1", "uuid1", "NEW", mockDate).
		AddRow("2", "uuid1", "NEW", mockDate).
		AddRow("3", "uuid1", "NEW", mockDate)

	sql := `
SELECT order_number, user_uuid, bonus_accrual_status, created_at 
FROM orders
WHERE bonus_accrual_status NOT IN ('PROCESSED', 'INVALID')
`
	mock.ExpectQuery(regexp.QuoteMeta(sql)).
		WillReturnRows(rows)

	// Вызов функции, которую мы тестируем
	orders, err := p.GetAllUnprocessedOrders(context.Background())

	expectedOrders := []models.Order{
		{
			OrderNumber:        "1",
			UserUUID:           "uuid1",
			BonusAccrualStatus: "NEW",
			BonusesAccrued:     0,
			CreatedAt:          mockDate,
		},
		{
			OrderNumber:        "2",
			UserUUID:           "uuid1",
			BonusAccrualStatus: "NEW",
			BonusesAccrued:     0,
			CreatedAt:          mockDate,
		},
		{
			OrderNumber:        "3",
			UserUUID:           "uuid1",
			BonusAccrualStatus: "NEW",
			BonusesAccrued:     0,
			CreatedAt:          mockDate,
		},
	}
	// Проверки
	assert.Nil(t, err)
	assert.Equal(t, orders, expectedOrders)

	// Проверка, что все ожидания mock были выполнены
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
func TestPostgres_GetUsersOrders(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	l, _ := logger.NewLogger(logger.DebugLevel, false)

	p := &Postgres{
		DB:     db,
		logger: l,
	}

	mockDate := time.Date(0001, 01, 01, 00, 00, 00, 00, time.UTC)
	const mockUUID = "uuid1"
	const newStatus = "NEW"
	rows := sqlmock.NewRows([]string{"order_number", "user_uuid", "bonus_accrual_status", "accrued", "created_at"}).
		AddRow("1", mockUUID, newStatus, 123.3, mockDate).
		AddRow("2", mockUUID, newStatus, 150.5, mockDate).
		AddRow("3", mockUUID, newStatus, 130, mockDate)

	sql := `
SELECT
    orders.order_number,
    orders.user_uuid,
    orders.bonus_accrual_status,
    COALESCE(accruals_calculated.accrued, 0) AS accrued,
    orders.created_at
FROM orders
LEFT JOIN accruals_calculated ON orders.order_number=accruals_calculated.order_number
WHERE orders.user_uuid=$1
ORDER BY orders.created_at
	`
	mock.ExpectQuery(regexp.QuoteMeta(sql)).
		WithArgs(mockUUID).
		WillReturnRows(rows)

	expectedOrders := []models.Order{
		{
			OrderNumber:        "1",
			UserUUID:           mockUUID,
			BonusAccrualStatus: newStatus,
			BonusesAccrued:     123.3,
			CreatedAt:          mockDate,
		},
		{
			OrderNumber:        "2",
			UserUUID:           mockUUID,
			BonusAccrualStatus: newStatus,
			BonusesAccrued:     150.5,
			CreatedAt:          mockDate,
		},
		{
			OrderNumber:        "3",
			UserUUID:           mockUUID,
			BonusAccrualStatus: newStatus,
			BonusesAccrued:     130,
			CreatedAt:          mockDate,
		},
	}

	orders, err := p.GetUsersOrders(context.TODO(), mockUUID)
	if err != nil {
		return
	}
	// Проверки
	assert.Nil(t, err)
	assert.Equal(t, orders, expectedOrders)

	// Проверка, что все ожидания mock были выполнены
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
func TestPostgres_GetUsersWithdrawals(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	l, _ := logger.NewLogger(logger.DebugLevel, false)

	p := &Postgres{
		DB:     db,
		logger: l,
	}

	mockDate := time.Date(0001, 01, 01, 00, 00, 00, 00, time.UTC)
	const mockUUID = "uuid1"
	rows := sqlmock.NewRows([]string{"order_number", "withdrew", "created_at"}).
		AddRow("1", 123.3, mockDate).
		AddRow("2", 150.5, mockDate).
		AddRow("3", 130, mockDate)

	sql := `
SELECT order_number, withdrew, created_at FROM withdrawals
WHERE user_uuid=$1
ORDER BY created_at
`
	mock.ExpectQuery(regexp.QuoteMeta(sql)).
		WithArgs(mockUUID).
		WillReturnRows(rows)

	expectedWithdrawals := []models.Withdrawal{
		{
			UserUUID:    mockUUID,
			OrderNumber: "1",
			Amount:      123.3,
			CreatedAt:   mockDate,
		},
		{
			UserUUID:    mockUUID,
			OrderNumber: "2",
			Amount:      150.5,
			CreatedAt:   mockDate,
		},
		{
			UserUUID:    mockUUID,
			OrderNumber: "3",
			Amount:      130,
			CreatedAt:   mockDate,
		},
	}
	w, err := p.GetUsersWithdrawals(context.TODO(), mockUUID)
	if err != nil {
		return
	}
	// Проверки
	assert.Nil(t, err)
	assert.Equal(t, w, expectedWithdrawals)

	// Проверка, что все ожидания mock были выполнены
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
