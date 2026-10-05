package config

import (
	mysqlAdapter "smsgateway/adapter/mysql"
)

type Config struct {
	Mysql          mysqlAdapter.Config
}
