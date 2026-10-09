package operatorrepository

import (
	"context"
	"smsgateway/entity"
)

type OperatorRepository interface {
	GetByID(ctx context.Context, operatorID uint) (entity.Operator, error)
}
