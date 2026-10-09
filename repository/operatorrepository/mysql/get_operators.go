package mysql

import (
	"context"
	"smsgateway/entity"
	"smsgateway/pkg/errmsg"
	"smsgateway/pkg/richerror"
)

func (m MysqlOperatorRepository) GetActives(ctx context.Context) ([]entity.Operator, error) {
	const op = "mysqloperatorrepo.GetActives"

	rows, err := m.adapter.Client().QueryContext(ctx,
		"SELECT id, name, total_tps, express_reserved_tps, is_active FROM operators WHERE is_active = 1",
	)
	if err != nil {
		err = errmsg.WrapMySQLError("get active operators", err)
		return nil, richerror.New(op).WithKind(richerror.KindUnexpected).
			WithErr(err).WithMessage(errmsg.ErrorMsgSomethingWentWrong)
	}
	defer rows.Close()

	var operators []entity.Operator
	for rows.Next() {
		var operator entity.Operator
		if err := rows.Scan(&operator.ID, &operator.Name, &operator.TotalTPS, &operator.ExpressReservedTPS, &operator.IsActive); err != nil {
			err = errmsg.WrapMySQLError("scan active operator", err)
			return nil, richerror.New(op).WithKind(richerror.KindUnexpected).
				WithErr(err).WithMessage(errmsg.ErrorMsgSomethingWentWrong)
		}

		operators = append(operators, operator)
	}

	if err := rows.Err(); err != nil {
		err = errmsg.WrapMySQLError("iterate active operators", err)
		return nil, richerror.New(op).WithKind(richerror.KindUnexpected).
			WithErr(err).WithMessage(errmsg.ErrorMsgSomethingWentWrong)
	}

	return operators, nil
}
