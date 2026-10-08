package config

import (
	mysqlAdapter "smsgateway/adapter/mysql"
	smsoperator "smsgateway/adapter/sms_operator"
)

type Config struct {
	HttpServer    HttpServer
	Mysql         mysqlAdapter.Config
	FirstOperator smsoperator.Config
}

type HttpServer struct {
	Host string
	Port string
}
