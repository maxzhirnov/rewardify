package store

import (
	"database/sql"
	"errors"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/maxzhirnov/rewardify/internal/logger"
)

var (
	ErrUserAlreadyExist  = errors.New("user already exists")
	ErrInsufficientFunds = errors.New("insufficient bonus balance")
)

type Postgres struct {
	DB     *sql.DB
	logger *logger.Logger
}

func NewPostgres(conn string, l *logger.Logger) (*Postgres, error) {
	db, err := sql.Open("pgx", conn)
	if err != nil {
		return nil, err
	}

	return &Postgres{
		DB:     db,
		logger: l,
	}, nil
}
