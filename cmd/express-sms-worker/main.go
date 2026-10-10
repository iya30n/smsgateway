package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	mysqlAdapter "smsgateway/adapter/mysql"
	"smsgateway/adapter/rabbitmq"
	smsoperator "smsgateway/adapter/sms_operator"
	"smsgateway/config"
	"smsgateway/internal/worker/smsworker"
	"smsgateway/pkg/logger"
	operatorRepository "smsgateway/repository/operatorrepository/mysql"
	smsNumberRepository "smsgateway/repository/smsnumberrepository/mysql"
	smsRepository "smsgateway/repository/smsrepository/mysql"
	"smsgateway/service"
	"smsgateway/service/smsservice"

	"go.uber.org/zap"
)

// Express SMS worker: consumes the sms.express queue.
func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cnf := config.Load()

	logger.Logger.Info("starting express sms worker")

	// adapters
	mysqlAdapter := mysqlAdapter.New(cnf.Mysql)
	rabbitmqAdapter, err := rabbitmq.New(cnf.RabbitMQ)
	if err != nil {
		logger.Logger.Fatal("failed to connect to rabbitmq", zap.Error(err))
	}

	// repositories
	smsRepo := smsRepository.NewMysql(mysqlAdapter)
	operatorRepo := operatorRepository.NewMysql(mysqlAdapter)
	smsNumberRepo := smsNumberRepository.NewMysql(mysqlAdapter)

	// services
	operatorService := service.NewOperatorService(operatorRepo, smsNumberRepo)
	smsService := smsservice.NewSMSService(smsRepo, rabbitmqAdapter, operatorService)

	// operators
	operator := smsoperator.NewFirstOperatorOperator(cnf.FirstOperator)

	worker := smsworker.NewExpressWorker(rabbitmqAdapter, operator, operatorService, smsService, rabbitmq.QueueExpress)
	if err := worker.Run(ctx); err != nil {
		logger.Logger.Error("worker stopped with an error", zap.Error(err))
		os.Exit(1)
	}

	logger.Logger.Info("worker stopped")
}
