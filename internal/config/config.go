package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/joho/godotenv"

	"github.com/maxzhirnov/rewardify/internal/logger"
	"github.com/maxzhirnov/rewardify/internal/utils"
)

const (
	nameRunAddress           = "a"
	nameDatabaseURI          = "d"
	nameAccrualSystemAddress = "r"

	defaultRunAddress           = ":8181"
	defaultDatabaseURI          = ""
	defaultAccrualSystemAddress = "http://localhost:8123"

	usageRunAddress           = "Provide address where app will run, e.g :8181"
	usageDatabaseURI          = "Provide URI string of Postgresql database"
	usageAccrualSystemAddress = "Provide address of accrual system"
)

type Config struct {
	runAddress           string
	databaseURI          string
	accrualSystemAddress string
	authSecretKey        string
	logger               *logger.Logger
}

func NewFromFlagsOrEnv(l *logger.Logger) *Config {
	l.Log.Info("Creating config from flags or environment variables")
	// Парсим флаги
	var runAddress string
	flag.StringVar(&runAddress, nameRunAddress, defaultRunAddress, usageRunAddress)

	var databaseURI string
	flag.StringVar(&runAddress, nameDatabaseURI, defaultDatabaseURI, usageDatabaseURI)

	var accrualSystemAddress string
	flag.StringVar(&runAddress, nameAccrualSystemAddress, defaultAccrualSystemAddress, usageAccrualSystemAddress)

	flag.Parse()

	// Попробуем загрузить .env файл, но если его нет, это ок
	if err := godotenv.Load(".env"); err != nil {
		l.Log.Info(".env file is not exist, running with system environment")
	}

	// Если находим переменные в окружении, то используем их
	if envRunAddress, ok := os.LookupEnv("RUN_ADDRESS"); ok {
		runAddress = envRunAddress
	}

	if envDatabaseURI, ok := os.LookupEnv("DATABASE_URI"); ok {
		databaseURI = envDatabaseURI
	}

	if envAccrualSystemAddress, ok := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS"); ok {
		accrualSystemAddress = envAccrualSystemAddress
	}

	authSecretKey, ok := os.LookupEnv("AUTH_SECRET_KEY")
	if !ok {
		l.Log.Warn("auth secret key parsed unsuccessful")
	}

	return &Config{
		runAddress:           runAddress,
		databaseURI:          databaseURI,
		accrualSystemAddress: accrualSystemAddress,
		authSecretKey:        authSecretKey,
		logger:               l,
	}
}

func (c Config) RunAddress() string {
	return c.runAddress
}

func (c Config) DatabaseURI() string {
	return c.databaseURI
}

func (c Config) AccrualSystemAddress() string {
	return c.accrualSystemAddress
}

func (c Config) AuthSecretKey() string {
	return c.authSecretKey
}

func (c Config) String() string {
	hiddenPasswordDatabaseURI := utils.HidePassword(c.databaseURI)
	return fmt.Sprintf("(configs: (runAddress: %s, databaseURI: %s, accrualSyrtemAddress: %s))", c.runAddress, hiddenPasswordDatabaseURI, c.accrualSystemAddress)
}
