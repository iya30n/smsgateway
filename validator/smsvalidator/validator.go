package smsvalidator

import (
	"smsgateway/repository/smsnumberrepository"
	"smsgateway/repository/userrepository"
)

const (
	mobileNumberRegex = `^09\d{9}$`
)

type Validator struct {
	userRepo      userrepository.UserRepository
	smsNumberRepo smsnumberrepository.SMSNumberRepository
}

func New(userRepo userrepository.UserRepository, smsNumberRepo smsnumberrepository.SMSNumberRepository) Validator {
	return Validator{
		userRepo:      userRepo,
		smsNumberRepo: smsNumberRepo,
	}
}
