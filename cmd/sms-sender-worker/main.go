package main

import (
	"context"
	"os"
	"os/signal"
	"smsgateway/service"
	"smsgateway/service/smsservice"
	"syscall"

	mysqlAdapter "smsgateway/adapter/mysql"
	"smsgateway/adapter/rabbitmq"
	smsoperator "smsgateway/adapter/sms_operator"
	"smsgateway/config"
	"smsgateway/internal/worker/smsworker"
	operatorRepository "smsgateway/repository/operatorrepository/mysql"
	smsNumberRepository "smsgateway/repository/smsnumberrepository/mysql"
	smsRepository "smsgateway/repository/smsrepository/mysql"
)

func main() {
	queueName := getQueueNameFromCli()

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
	smsService := smsservice.NewSMSService(smsRepo, operatorService)

	// operators
	operator := smsoperator.NewFirstOperatorOperator(cnf.FirstOperator)

	if queueName == rabbitmq.QueueExpress {
		worker := smsworker.NewExpressWorker(rabbitmqAdapter, operator, operatorService, queueName)
		if err := worker.Run(ctx); err != nil {
			// TODO: handle the error
		}

		return
	}

	worker := smsworker.New(rabbitmqAdapter, operator, smsService, queueName)
	if err := worker.Run(ctx); err != nil {
		// TODO: handle the error
	}
}

func getQueueNameFromCli() string {
	queueName := rabbitmq.QueueNormal
	if len(os.Args) > 1 {
		qParam := os.Args[1]
		if qParam != rabbitmq.QueueNormal && qParam != rabbitmq.QueueExpress {
			panic("invalid queue name!")
		}

		queueName = qParam
	}

	return queueName
}
