package smsvalidator

import (
	"errors"
	"smsgateway/pkg/errmsg"
)

func (v Validator) sourceNumberIsActive(value interface{}) error {
	number, ok := value.(string)
	if !ok {
		return errors.New(errmsg.ErrorMsgSourceNumberIsNotValid)
	}

	isActive, err := v.smsNumberRepo.IsActive(number)
	if err != nil {
		return errors.New(errmsg.ErrorMsgSourceNumberIsNotValid)
	}

	if !isActive {
		return errors.New(errmsg.ErrorMsgSourceNumberIsNotValid)
	}

	return nil
}
