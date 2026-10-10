package mysql

import (
	"database/sql"
	"fmt"
	"smsgateway/pkg/logger"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
)

type Config struct {
	Username string
	Password string
	Host     string
	Port     string
	DBName   string
}

type Adapter struct {
	client *sql.DB
}

func New(config Config) Adapter {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@(%s:%s)/%s?parseTime=true", config.Username, config.Password, config.Host, config.Port, config.DBName))
	if err != nil {
		logger.Logger.Fatal("failed to open mysql connection",
			zap.String("host", config.Host),
			zap.String("port", config.Port),
			zap.String("database", config.DBName),
			zap.Error(err),
		)
	}

	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	return Adapter{db}
}

func (a Adapter) Client() *sql.DB {
	return a.client
}
