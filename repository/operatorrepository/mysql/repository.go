package mysql

import (
	"smsgateway/adapter/mysql"
	"smsgateway/repository/operatorrepository"
)

var _ operatorrepository.OperatorRepository = &MysqlOperatorRepository{}

type MysqlOperatorRepository struct {
	adapter mysql.Adapter
}

func NewMysql(adapter mysql.Adapter) operatorrepository.OperatorRepository {
	return &MysqlOperatorRepository{adapter}
}
