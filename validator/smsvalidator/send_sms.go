package smsvalidator

import (
	"errors"
	"regexp"
	"smsgateway/pkg/errmsg"
	"smsgateway/pkg/richerror"
	"smsgateway/param/smsparam"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func (v Validator) ValidateSendSMSRequest(r smsparam.SendSMSRequest) (map[string]string, error) {
	const op = "validator.ValidateSendSMSRequest"
	fieldErrors := make(map[string]string)

	err := validation.ValidateStruct(&r,
		validation.Field(&r.IdempotencyKey, validation.Required, validation.Length(16, 191)),
		validation.Field(&r.UserID, validation.Required, validation.Length(11, 11), validation.By(func(value interface{}) error {
			return v.userHasEnoughBalance(r, value)
		})),
		validation.Field(&r.SourceNumber, validation.Required, validation.Length(7, 14), validation.Match(regexp.MustCompile(mobileNumberRegex))),
		validation.Field(&r.ReceptorNumber, validation.Required, validation.Length(11, 11), validation.Match(regexp.MustCompile(mobileNumberRegex))),
		validation.Field(&r.Content, validation.Required, validation.Length(1, 160)),
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

func (v Validator) userHasEnoughBalance(r smsparam.SendSMSRequest, value interface{}) error {
	userID := value.(uint)
	hasEnough, err := v.userRepo.HasEnoughBalanceForSMS(userID, r.SmsType)
	if err != nil {
		return errors.New(errmsg.ErrorMsgUserIDIsNotValid)
	}

	if !hasEnough {
		return errors.New(errmsg.ErrorMsgUserHasNotEnoughBalance)
	}

	return nil
}
