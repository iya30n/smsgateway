package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	mysqlAdapter "smsgateway/adapter/mysql"
	"smsgateway/config"
	"smsgateway/delivery/httpserver/smshandler"
	"smsgateway/delivery/httpserver/userhandler"
	"smsgateway/repository/migrator"
	operatormysqlrepository "smsgateway/repository/operatorrepository/mysql"
	smsnumbermysqlrepository "smsgateway/repository/smsnumberrepository/mysql"
	smsmysqlrepository "smsgateway/repository/smsrepository/mysql"
	usermysqlrepository "smsgateway/repository/userrepository/mysql"
	"smsgateway/service"
	"smsgateway/service/smsservice"
	"smsgateway/validator/smsvalidator"
	"smsgateway/validator/uservalidator"
	"syscall"
	"time"

	"github.com/labstack/echo/v5"
)

func main() {
	cnf := config.Load()

	// adapters
	mysqlAdapter := mysqlAdapter.New(cnf.Mysql)

	// migrate db tables
	migrator := migrator.NewMigrator(mysqlAdapter, "mysql", "repository/mysql/migrations")
	migrator.Up()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// repositories
	userRepo := usermysqlrepository.NewMysql(mysqlAdapter)
	smsRepo := smsmysqlrepository.NewMysql(mysqlAdapter)
	operatorRepo := operatormysqlrepository.NewMysql(mysqlAdapter)
	smsNumberRepo := smsnumbermysqlrepository.NewMysql(mysqlAdapter)

	// services
	userSvc := service.NewUserService(userRepo)
	operatorSvc := service.NewOperatorService(operatorRepo, smsNumberRepo)
	smsSvc := smsservice.NewSMSService(smsRepo, operatorSvc)

	e := echo.New()

	// user routes
	userHandler := userhandler.New(e, userSvc, uservalidator.New(userRepo))
	userHandler.SetupRoutes()

	// sms routes
	smsHandler := smshandler.New(e, smsSvc, smsvalidator.New(userRepo, smsNumberRepo))
	smsHandler.SetupRoutes()

	sc := echo.StartConfig{
		Address:         fmt.Sprintf("%s:%s", cnf.HttpServer.Host, cnf.HttpServer.Port),
		GracefulTimeout: time.Second * 30,
	}

	if err := sc.Start(ctx, e); err != nil {
		panic(err)
	}
}
