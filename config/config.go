package config

import (
	mysqlAdapter "smsgateway/adapter/mysql"
)

type Config struct {
	HttpServer     HttpServer
	Mysql          mysqlAdapter.Config
}

type HttpServer struct {
	Port string
}