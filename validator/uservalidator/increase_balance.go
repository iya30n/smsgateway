package uservalidator

import (
	"errors"
	"smsgateway/param/userparam"
	"smsgateway/pkg/errmsg"
	"smsgateway/pkg/richerror"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/shopspring/decimal"
)

func (v Validator) ValidateIncreaseBalanceRequest(r userparam.IncreaseBalanceRequest) (map[string]string, error) {
	const op = "validator.ValidateIncreaseBalanceRequest"
	fieldErrors := make(map[string]string)

	err := validation.ValidateStruct(&r,
		validation.Field(&r.UserID, validation.Required, validation.Length(11, 11), validation.By(v.doesUserIDExist)),
		validation.Field(&r.Amount, validation.Required, validation.By(isAtLeastMinTopUp)),
	)

	if err == nil {
		return fieldErrors, nil
	}

	errV, ok := err.(validation.Errors)
	if !ok {
		return fieldErrors, richerror.New(op).WithKind(richerror.KindUnexpected).
			WithErr(err).WithMessage(errmsg.ErrorMsgSomethingWentWrong).
			WithMeta(map[string]interface{}{"rq": r})
	}

	for key, value := range errV {
		errR, ok := value.(richerror.RichError)
		if ok {
			return fieldErrors, errR
		}

		if value != nil {
			fieldErrors[key] = value.Error()
		}
	}

	return fieldErrors, richerror.New(op).WithKind(richerror.KindInvalidInput).
		WithErr(err).WithMessage(errmsg.ErrorMsgInvalidInput).
		WithMeta(map[string]interface{}{"rq": r})
}

const minTopUpAmount = 5_000_000

func isAtLeastMinTopUp(value interface{}) error {
	amount, ok := value.(decimal.Decimal)
	if !ok {
		return errors.New(errmsg.ErrorMsgInvalidInput)
	}

	if amount.LessThan(decimal.NewFromInt(minTopUpAmount)) {
		return errors.New(errmsg.ErrorMsgInvalidInput)
	}

	return nil
}

func (v Validator) doesUserIDExist(value interface{}) error {
	userID := value.(uint)
	_, err := v.userRepo.GetUserByID(userID)
	if err != nil {
		return errors.New(errmsg.ErrorMsgUserIDIsNotValid)
	}

	return nil
}
