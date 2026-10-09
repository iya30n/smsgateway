package smsvalidator

import (
	"regexp"
	"smsgateway/entity"
	"smsgateway/param/smsparam"
	"smsgateway/pkg/errmsg"
	"smsgateway/pkg/richerror"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func (v Validator) ValidateSendExpressSMSRequest(r smsparam.SendExpressSMSRequest) (map[string]string, error) {
	const op = "validator.ValidateSendExpressSMSRequest"
	fieldErrors := make(map[string]string)

	err := validation.ValidateStruct(&r,
		validation.Field(&r.IdempotencyKey, validation.Required, validation.Length(16, 191)),
		validation.Field(&r.UserID, validation.Required, validation.Length(11, 11), validation.By(func(value interface{}) error {
			return v.userHasEnoughBalance(entity.SmsTypeExpress, value)
		})),
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
