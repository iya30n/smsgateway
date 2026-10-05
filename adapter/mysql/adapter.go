package mysql

import (
	"database/sql"
	"fmt"
	"time"
	_ "github.com/go-sql-driver/mysql"
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
		panic(err)
	}

	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	return Adapter{db}
}

func (a Adapter) Client() *sql.DB {
	return a.client
}
