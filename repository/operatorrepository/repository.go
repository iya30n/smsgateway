package operatorrepository

import (
	"context"
	"smsgateway/entity"
)

type OperatorRepository interface {
	GetByID(ctx context.Context, operatorID uint) (entity.Operator, error)
	// GetActives returns the operators that are currently in service. It backs
	// operator selection, so callers can spread traffic over a healthy set
	// instead of pinning every message to one operator.
	GetActives(ctx context.Context) ([]entity.Operator, error)
}
