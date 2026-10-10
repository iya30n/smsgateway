package smshandler

import (
	"net/http"
	"smsgateway/param/smsparam"
	"smsgateway/pkg/httpmsg"
	"smsgateway/pkg/logger"

	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
)

func (h Handler) ReportSMS(c *echo.Context) error {
	var req smsparam.ReportSMSRequest
	if err := c.Bind(&req); err != nil {
		logger.Logger.Warn("failed to bind report sms request", zap.Error(err))

		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	errFields, err := h.smsValidator.ValidateReportSMSRequest(req)
	if err != nil {
		code, msg := httpmsg.MapRichErrKindsToHttpResponse(err)
		return c.JSON(code, map[string]any{
			"message": msg,
			"errors":  errFields,
		})
	}

	response, err := h.smsSvc.ReportSMS(c.Request().Context(), req)
	if err != nil {
		code, msg := httpmsg.MapRichErrKindsToHttpResponse(err)
		return c.JSON(code, map[string]any{"message": msg})
	}

	return c.JSON(http.StatusOK, response)
}
