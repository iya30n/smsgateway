package userrepository

import (
	"context"
	"smsgateway/entity"
)

type UserRepository interface {
	GetUserByID(userID uint) (entity.User, error)
	IncreaseBalance(ctx context.Context, userID uint, amount float64) (float64, error)
	HasEnoughBalanceForSMS(userID uint, smsType entity.SmsType) (bool, error)
}
