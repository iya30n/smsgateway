package smsvalidator

import (
	"errors"
	"regexp"
	"smsgateway/entity"
	"smsgateway/param/smsparam"
	"smsgateway/pkg/errmsg"
	"smsgateway/pkg/richerror"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func (v Validator) ValidateReportSMSRequest(r smsparam.ReportSMSRequest) (map[string]string, error) {
	const op = "validator.ValidateReportSMSRequest"
	fieldErrors := make(map[string]string)

	structErr := validation.ValidateStruct(&r,
		validation.Field(&r.Receptor, validation.Match(regexp.MustCompile(mobileNumberRegex))),
		validation.Field(&r.Status, validation.In(
			entity.MessageStatusInitiated,
			entity.MessageStatusQueued,
			entity.MessageStatusSent,
			entity.MessageStatusFailed,
		)),
		validation.Field(&r.SmsType, validation.In(
			entity.SmsTypeNormal,
			entity.SmsTypeExpress,
		)),
		validation.Field(&r.From, validation.Date(smsparam.ReportDateLayout)),
		validation.Field(&r.To, validation.Date(smsparam.ReportDateLayout)),
		validation.Field(&r.Page, validation.When(r.Page != 0, validation.Min(1))),
		validation.Field(&r.PerPage, validation.When(r.PerPage != 0,
			validation.Min(1), validation.Max(smsparam.MaxReportPerPage))),
	)

	if structErr != nil {
		errV, ok := structErr.(validation.Errors)
		if !ok {
			return fieldErrors, richerror.New(op).WithKind(richerror.KindUnexpected).
				WithErr(structErr).WithMessage(errmsg.ErrorMsgSomethingWentWrong).
				WithMeta(map[string]interface{}{"rq": r})
		}

		for key, value := range errV {
			if errR, ok := value.(richerror.RichError); ok {
				return fieldErrors, errR
			}

			if value != nil {
				fieldErrors[key] = value.Error()
			}
		}
	}

	if rangeErr := v.validateReportDateRange(r); rangeErr != nil {
		fieldErrors["to"] = rangeErr.Error()
	}

	if len(fieldErrors) == 0 {
		return fieldErrors, nil
	}

	return fieldErrors, richerror.New(op).WithKind(richerror.KindInvalidInput).
		WithErr(structErr).WithMessage(errmsg.ErrorMsgInvalidInput).
		WithMeta(map[string]interface{}{"rq": r})
}

func (v Validator) validateReportDateRange(r smsparam.ReportSMSRequest) error {
	if r.From == "" || r.To == "" {
		return nil
	}

	from, err := time.Parse(smsparam.ReportDateLayout, r.From)
	if err != nil {
		return nil
	}

	to, err := time.Parse(smsparam.ReportDateLayout, r.To)
	if err != nil {
		return nil
	}

	if to.Before(from) {
		return errors.New(errmsg.ErrorMsgInvalidDateRange)
	}

	return nil
}
