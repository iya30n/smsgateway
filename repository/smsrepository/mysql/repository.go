package mysql

import (
	"smsgateway/adapter/mysql"
	"smsgateway/repository/smsrepository"
)

var _ smsrepository.SMSRepository = &MysqlSMSRepository{}

type MysqlSMSRepository struct {
	adapter mysql.Adapter
}

func NewMysql(adapter mysql.Adapter) smsrepository.SMSRepository {
	return &MysqlSMSRepository{adapter}
}
