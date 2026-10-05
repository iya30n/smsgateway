package userhandler

import (
	"net/http"
	"smsgateway/pkg/httpmsg"
	"smsgateway/service"

	"github.com/labstack/echo/v5"
)

func (h Handler) IncreaseBalance(c *echo.Context) error {
	var increaseBalanceReq service.IncreaseBalanceRequest
	if err := c.Bind(&increaseBalanceReq); err != nil {
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
