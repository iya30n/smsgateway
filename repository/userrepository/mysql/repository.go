package mysql

import (
	"smsgateway/adapter/mysql"
	"smsgateway/repository/userrepository"
)

var _ userrepository.UserRepository = &MysqlUserRepository{}

type MysqlUserRepository struct {
	adapter mysql.Adapter
}

func NewMysql(adapter mysql.Adapter) userrepository.UserRepository {
	return &MysqlUserRepository{adapter}
}
