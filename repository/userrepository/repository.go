package userrepository

import (
	"context"
	"smsgateway/entity"

	"github.com/shopspring/decimal"
)

type UserRepository interface {
	GetUserByID(userID uint) (entity.User, error)
	IncreaseBalance(ctx context.Context, userID uint, amount decimal.Decimal) (decimal.Decimal, error)
	HasEnoughBalanceForSMS(userID uint, smsType entity.SmsType) (bool, error)
}
