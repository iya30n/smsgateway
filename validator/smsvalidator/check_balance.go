package smsvalidator

import (
	"errors"
	"smsgateway/entity"
	"smsgateway/pkg/errmsg"
)

func (v Validator) userHasEnoughBalance(smsType entity.SmsType, value interface{}) error {
	userID := value.(uint)
	hasEnough, err := v.userRepo.HasEnoughBalanceForSMS(userID, smsType)
	if err != nil {
		return errors.New(errmsg.ErrorMsgUserIDIsNotValid)
	}

	if !hasEnough {
		return errors.New(errmsg.ErrorMsgUserHasNotEnoughBalance)
	}

	return nil
}
