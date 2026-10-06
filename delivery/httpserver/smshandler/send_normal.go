package smshandler

import (
	"net/http"
	"smsgateway/entity"
	"smsgateway/pkg/errmsg"
	"smsgateway/pkg/httpmsg"
	"smsgateway/service"

	"github.com/labstack/echo/v5"
)

func (h Handler) SendNormalSMS(c *echo.Context) error {
	var req service.SendSMSRequest
	if err := c.Bind(&req); err != nil {
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

	response, err := h.smsSvc.SendNormalSMS(c.Request().Context(),req)
	if err != nil {
		return echo.NewHTTPError(httpmsg.MapRichErrKindsToHttpResponse(err))
	}

	return c.JSON(http.StatusOK, response)
}
