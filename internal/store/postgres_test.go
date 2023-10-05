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
