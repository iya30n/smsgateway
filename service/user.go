package service

import (
	"context"
	"smsgateway/param/userparam"
)

type UserService struct {
	userRepo UserRepository
}

func NewUserService(userRepo UserRepository) UserService {
	return UserService{
		userRepo: userRepo,
	}
}

type UserRepository interface {
	IncreaseBalance(ctx context.Context, userID uint, amount float64) (float64, error)
}

func (s UserService) IncreaseBalance(ctx context.Context, req userparam.IncreaseBalanceRequest) (userparam.IncreaseBalanceResponse, error) {
	// TODO: create a balance_request, and redirect user to ipg
	// TODO: create a transaction with type manual_adjustment, and link it to the balance_request after verifying ipg transaction
	// TODO: increase the user's balance after verifying ipg transaction, and link it to the balance_request and transaction

	newBalance, err := s.userRepo.IncreaseBalance(ctx, req.UserID, req.Amount)
	if err != nil {
		return userparam.IncreaseBalanceResponse{}, err
	}

	return userparam.IncreaseBalanceResponse{
		NewBalance: newBalance,
	}, nil
}
