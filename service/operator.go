package service

import (
	"context"
	"smsgateway/entity"
)

type OperatorService struct {
	operatorRepo OperatorRepository
}

func NewOperatorService(operatorRepo OperatorRepository) OperatorService {
	return OperatorService{operatorRepo: operatorRepo}
}

type OperatorRepository interface {
	GetActives(ctx context.Context) ([]entity.Operator, error)
}

func (s OperatorService) GetActives(ctx context.Context) ([]entity.Operator, error) {
	return s.operatorRepo.GetActives(ctx)
}
