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
	operatorRepository "smsgateway/repository/operatorrepository/mysql"
	smsNumberRepository "smsgateway/repository/smsnumberrepository/mysql"
	smsRepository "smsgateway/repository/smsrepository/mysql"
	"smsgateway/service"
	"smsgateway/service/smsservice"
)

// Normal SMS worker: consumes the sms.normal queue.
func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cnf := config.Load()

	// adapters
	mysqlAdapter := mysqlAdapter.New(cnf.Mysql)
	rabbitmqAdapter, err := rabbitmq.New(cnf.RabbitMQ)
	if err != nil {
		// TODO: alert
		panic(err)
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

	worker := smsworker.New(rabbitmqAdapter, operator, smsService, rabbitmq.QueueNormal)
	if err := worker.Run(ctx); err != nil {
		// TODO: handle the error
	}
}
