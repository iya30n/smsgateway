package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	mysqlAdapter "smsgateway/adapter/mysql"
	"smsgateway/adapter/rabbitmq"
	smsoperator "smsgateway/adapter/sms_operator"
	"smsgateway/config"
	"smsgateway/internal/worker/smsworker"
	smsRepository "smsgateway/repository/smsrepository/mysql"
	"smsgateway/service"
)

func main() {
	queueName := os.Args[1]
	if len(queueName) > 0 && (queueName != rabbitmq.QueueNormal || queueName != rabbitmq.QueueExpress) {
		panic("invalid queue name!")
	} else {
		queueName = rabbitmq.QueueNormal
		fmt.Println("normal queue selected by default!")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cnf := config.Load()

	// adapters
	mysqlAdapter := mysqlAdapter.New(cnf.Mysql)
	rabbitmqAdapter, err := rabbitmq.New(cnf.RabbitMQ)
	if err != nil {
		panic(err)
		// TODO: alert
	}

	// repositories
	smsRepo := smsRepository.NewMysql(mysqlAdapter)

	// services
	smsService := service.NewSMSService(smsRepo)

	// operators
	operator := smsoperator.NewFirstOperatorOperator(cnf.FirstOperator)

	// worker
	worker := smsworker.New(rabbitmqAdapter, operator, smsService, queueName)

	if err := worker.Run(ctx); err != nil {
		// TODO: handle the error
	}
}