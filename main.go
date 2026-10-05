package main

import (
	mysqlAdapter "smsgateway/adapter/mysql"
	"smsgateway/config"
	"smsgateway/repository/migrator"
)

func main() {
	cnf := config.Load()

	// adapters
	mysqlAdapter := mysqlAdapter.New(cnf.Mysql)

	// migrate db tables
	migrator := migrator.NewMigrator(mysqlAdapter, "mysql", "repository/mysql/migrations")
	migrator.Up()
}
