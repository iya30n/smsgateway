package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"smsgateway/entity"
	"smsgateway/pkg/errmsg"
	"smsgateway/pkg/richerror"
)

func (m MysqlOperatorRepository) GetByID(ctx context.Context, operatorID uint) (entity.Operator, error) {
	const op = "mysqloperatorrepo.GetByID"

	var operator entity.Operator
	err := m.adapter.Client().QueryRowContext(ctx,
		"SELECT id, name, total_tps, express_reserved_tps, is_active FROM operators WHERE id = ?",
		operatorID,
	).Scan(&operator.ID, &operator.Name, &operator.TotalTPS, &operator.ExpressReservedTPS, &operator.IsActive)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return operator, richerror.New(op).WithKind(richerror.KindNotFound).
				WithMessage(errmsg.ErrorMsgNotFound).
				WithMeta(map[string]interface{}{"operator_id": operatorID})
		}

		err = errmsg.WrapMySQLError(fmt.Sprintf("get operator %d", operatorID), err)
		return operator, richerror.New(op).WithKind(richerror.KindUnexpected).
			WithErr(err).WithMessage(errmsg.ErrorMsgSomethingWentWrong).
			WithMeta(map[string]interface{}{"operator_id": operatorID})
	}

	return operator, nil
}
