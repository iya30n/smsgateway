package service

import (
	"context"
	"smsgateway/entity"
)

type OperatorService struct {
	operatorRepo  OperatorRepository
	smsNumberRepo SMSNumberRepository
}

func NewOperatorService(operatorRepo OperatorRepository, smsNumberRepo SMSNumberRepository) OperatorService {
	return OperatorService{
		operatorRepo:  operatorRepo,
		smsNumberRepo: smsNumberRepo,
	}
}

type OperatorRepository interface {
	GetActives(ctx context.Context) ([]entity.Operator, error)
}

type SMSNumberRepository interface {
	GetActiveNumberByOperator(operatorID uint) (string, bool, error)
}

type OperatorTarget struct {
	OperatorID uint
	Number     string
}

func (s OperatorService) GetActives(ctx context.Context) ([]entity.Operator, error) {
	return s.operatorRepo.GetActives(ctx)
}

func (s OperatorService) GetStableTargets(ctx context.Context) ([]OperatorTarget, error) {
	operators, err := s.operatorRepo.GetActives(ctx)
	if err != nil {
		return nil, err
	}

	targets := make([]OperatorTarget, 0, len(operators))
	for _, operator := range operators {
		number, ok, err := s.smsNumberRepo.GetActiveNumberByOperator(operator.ID)
		if err != nil {
			return nil, err
		}

		if !ok {
			continue
		}

		targets = append(targets, OperatorTarget{
			OperatorID: operator.ID,
			Number:     number,
		})
	}

	return targets, nil
}
