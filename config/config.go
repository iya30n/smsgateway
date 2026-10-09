package config

import (
	mysqlAdapter "smsgateway/adapter/mysql"
	rabbitmqAdapter "smsgateway/adapter/rabbitmq"
	redisAdapter "smsgateway/adapter/redis"
	smsoperator "smsgateway/adapter/sms_operator"
	capacityRedis "smsgateway/capacity/redis"
)

type Config struct {
	HttpServer    HttpServer
	Mysql         mysqlAdapter.Config
	RabbitMQ      rabbitmqAdapter.Config
	Redis         redisAdapter.Config
	Capacity      capacityRedis.Config
	FirstOperator smsoperator.Config
}

type HttpServer struct {
	Host string
	Port string
}
