package smshandler

import (
	"net/http"
	"smsgateway/entity"
	"smsgateway/param/smsparam"
	"smsgateway/pkg/errmsg"
	"smsgateway/pkg/httpmsg"
	"smsgateway/pkg/logger"

	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
)

func (h Handler) SendNormalSMS(c *echo.Context) error {
	var req smsparam.SendSMSRequest
	if err := c.Bind(&req); err != nil {
		logger.Logger.Warn("failed to bind send sms request", zap.Error(err))

		return echo.NewHTTPError(http.StatusBadRequest, errmsg.ErrorMsgInvalidInput)
	}

	req.IdempotencyKey = c.Request().Header.Get("Idempotency-Key")

	req.SmsType = entity.SmsTypeNormal

	errFields, err := h.smsValidator.ValidateSendSMSRequest(req)
	if err != nil {
		code, msg := httpmsg.MapRichErrKindsToHttpResponse(err)
		return c.JSON(code, map[string]any{
			"message": msg,
			"errors":  errFields,
		})
	}

	response, err := h.smsSvc.SendNormalSMS(c.Request().Context(), req)
	if err != nil {
		code, msg := httpmsg.MapRichErrKindsToHttpResponse(err)
		return c.JSON(code, map[string]any{"message": msg})
	}

	return c.JSON(http.StatusOK, response)
}
