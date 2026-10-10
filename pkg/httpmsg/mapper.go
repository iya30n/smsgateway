package httpmsg

import (
	"errors"
	"net/http"
	"smsgateway/pkg/errmsg"
	"smsgateway/pkg/logger"
	"smsgateway/pkg/richerror"

	"go.uber.org/zap"
)

func MapRichErrKindsToHttpResponse(err error) (code int, message string) {
	// errors.As (not a type assertion) so a RichError wrapped by another layer is
	// still recognised.
	var richErr richerror.RichError
	if !errors.As(err, &richErr) {
		// Unhandled error type: this is a bug in the calling code, so log the raw
		// error rather than the generic message the client gets.
		logger.Logger.Error("unhandled error type",
			zap.String("op", "httpmsg.MapRichErrKindsToHttpResponse"),
			zap.Error(err),
		)

		return 500, errmsg.ErrorMsgSomethingWentWrong
	}

	switch richErr.Kind() {
	case richerror.KindInvalidInput:
		return http.StatusBadRequest, richErr.Error()
	case richerror.KindNotFound:
		return http.StatusNotFound, richErr.Error()
	case richerror.KindUnauthenticated:
		return http.StatusUnauthorized, richErr.Error()
	case richerror.KindForbidden:
		return http.StatusForbidden, richErr.Error()
	case richerror.KindUnexpected:
		logUnexpected(richErr)
		return http.StatusInternalServerError, richErr.Error()
	default:
		// An unset kind is a programming error, not a client error.
		logUnexpected(richErr)
		return http.StatusInternalServerError, richErr.Error()
	}
}

func logUnexpected(richErr richerror.RichError) {
	logger.Logger.Error("unexpected error",
		zap.String("op", string(richErr.Operation())),
		zap.String("kind", richErr.Kind().String()),
		zap.Error(richErr),
	)
}
