package userrepository

import (
	"context"
	"smsgateway/adapter/mysql"
)

type UserRepository interface {
	IncreaseBalance(ctx context.Context, userID uint, amount float64) (float64, error)
}

var _ UserRepository = &MysqlUserRepository{}

type MysqlUserRepository struct {
	adapter mysql.Adapter
}

func NewMysql(adapter mysql.Adapter) UserRepository {
	return &MysqlUserRepository{adapter}
}
