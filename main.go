package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	mysqlAdapter "smsgateway/adapter/mysql"
	"smsgateway/adapter/rabbitmq"
	"smsgateway/config"
	"smsgateway/delivery/httpserver/smshandler"
	"smsgateway/delivery/httpserver/userhandler"
	"smsgateway/pkg/logger"
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
	"go.uber.org/zap"
)

func main() {
	cnf := config.Load()

	logger.Logger.Info("starting sms gateway api",
		zap.String("host", cnf.HttpServer.Host),
		zap.String("port", cnf.HttpServer.Port),
	)

	// adapters
	mysqlAdapter := mysqlAdapter.New(cnf.Mysql)
	rabbitmqAdapter, err := rabbitmq.New(cnf.RabbitMQ)
	if err != nil {
		logger.Logger.Fatal("failed to connect to rabbitmq", zap.Error(err))
	}
	defer rabbitmqAdapter.Close()

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
	smsSvc := smsservice.NewSMSService(smsRepo, rabbitmqAdapter, operatorSvc)

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
		logger.Logger.Fatal("http server stopped", zap.Error(err))
	}

	logger.Logger.Info("http server stopped")
}
