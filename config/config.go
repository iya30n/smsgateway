package config

import (
	mysqlAdapter "smsgateway/adapter/mysql"
	rabbitmqAdapter "smsgateway/adapter/rabbitmq"
	smsoperator "smsgateway/adapter/sms_operator"
)

type Config struct {
	HttpServer    HttpServer
	Mysql         mysqlAdapter.Config
	RabbitMQ      rabbitmqAdapter.Config
	FirstOperator smsoperator.Config
}

type HttpServer struct {
	Host string
	Port string
}
