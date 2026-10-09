package mysql

import (
	"smsgateway/adapter/mysql"
	"smsgateway/repository/smsnumberrepository"
)

var _ smsnumberrepository.SMSNumberRepository = &MysqlSMSNumberRepository{}

type MysqlSMSNumberRepository struct {
	adapter mysql.Adapter
}

func NewMysql(adapter mysql.Adapter) smsnumberrepository.SMSNumberRepository {
	return &MysqlSMSNumberRepository{adapter}
}
