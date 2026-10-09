package smshandler

import (
	"net/http"
	"smsgateway/entity"
	"smsgateway/param/smsparam"
	"smsgateway/pkg/errmsg"
	"smsgateway/pkg/httpmsg"

	"github.com/labstack/echo/v5"
)

// i have defined send express as a separate function because it may have different logic in the real-world scenario.
func (h Handler) SendExpressSMS(c *echo.Context) error {
	var req smsparam.SendSMSRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errmsg.ErrorMsgInvalidInput)
	}

	req.IdempotencyKey = c.Request().Header.Get("Idempotency-Key")

	req.SmsType = entity.SmsTypeExpress

	errFields, err := h.smsValidator.ValidateSendSMSRequest(req)
	if err != nil {
		code, msg := httpmsg.MapRichErrKindsToHttpResponse(err)
		return c.JSON(code, map[string]any{
			"message": msg,
			"errors":  errFields,
		})
	}

	response, err := h.smsSvc.SendExpressSMS(c.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(httpmsg.MapRichErrKindsToHttpResponse(err))
	}

	return c.JSON(http.StatusOK, response)
}
