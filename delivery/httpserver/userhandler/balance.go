package userhandler

import (
	"net/http"
	"smsgateway/param/userparam"
	"smsgateway/pkg/httpmsg"
	"smsgateway/pkg/logger"

	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
)

func (h Handler) IncreaseBalance(c *echo.Context) error {
	var increaseBalanceReq userparam.IncreaseBalanceRequest
	if err := c.Bind(&increaseBalanceReq); err != nil {
		logger.Logger.Warn("failed to bind increase balance request", zap.Error(err))

		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	errFields, err := h.userValidator.ValidateIncreaseBalanceRequest(increaseBalanceReq)
	if err != nil {
		code, msg := httpmsg.MapRichErrKindsToHttpResponse(err)
		return c.JSON(code, map[string]any{
			"message": msg,
			"errors":  errFields,
		})
	}

	response, err := h.userSvc.IncreaseBalance(c.Request().Context(), increaseBalanceReq)
	if err != nil {
		return echo.NewHTTPError(httpmsg.MapRichErrKindsToHttpResponse(err))
	}

	return c.JSON(http.StatusOK, response)
}
