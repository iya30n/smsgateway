package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"smsgateway/pkg/errmsg"
	"smsgateway/pkg/richerror"
)

func (m MysqlSMSNumberRepository) GetActiveNumberByOperator(operatorID uint) (string, bool, error) {
	const op = "mysqlsmsnumberrepo.GetActiveNumberByOperator"

	var number string
	err := m.adapter.Client().QueryRowContext(context.Background(),
		"SELECT number FROM sms_numbers WHERE operator_id = ? AND is_active = 1 LIMIT 1",
		operatorID,
	).Scan(&number)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		}

		err = errmsg.WrapMySQLError(fmt.Sprintf("get active number for operator %d", operatorID), err)
		return "", false, richerror.New(op).WithKind(richerror.KindUnexpected).
			WithErr(err).WithMessage(errmsg.ErrorMsgSomethingWentWrong).
			WithMeta(map[string]interface{}{"operator_id": operatorID})
	}

	return number, true, nil
}
