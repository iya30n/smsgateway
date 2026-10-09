package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"smsgateway/pkg/errmsg"
	"smsgateway/pkg/richerror"
)

// IsActive reports whether the number exists and is currently usable as a
// source number. A missing number is not an error: it is simply not active.
func (m MysqlSMSNumberRepository) IsActive(number string) (bool, error) {
	const op = "mysqlsmsnumberrepo.IsActive"

	var isActive bool
	err := m.adapter.Client().QueryRowContext(context.Background(),
		"SELECT is_active FROM sms_numbers WHERE number = ?",
		number,
	).Scan(&isActive)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		err = errmsg.WrapMySQLError(fmt.Sprintf("get sms number %s", number), err)
		return false, richerror.New(op).WithKind(richerror.KindUnexpected).
			WithErr(err).WithMessage(errmsg.ErrorMsgSomethingWentWrong).
			WithMeta(map[string]interface{}{"number": number})
	}

	return isActive, nil
}
