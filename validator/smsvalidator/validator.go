package smsvalidator

import (
	"smsgateway/repository/userrepository"
)

const (
	mobileNumberRegex = `^09\d{9}$`
)

type Validator struct {
	userRepo userrepository.UserRepository
}

func New(userRepo userrepository.UserRepository) Validator {
	return Validator{userRepo: userRepo}
}
