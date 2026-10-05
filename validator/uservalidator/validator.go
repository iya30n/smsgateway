package uservalidator

import (
	"smsgateway/repository/userrepository"
)

type Validator struct {
	userRepo userrepository.UserRepository
}

func New(userRepo userrepository.UserRepository) Validator {
	return Validator{userRepo: userRepo}
}