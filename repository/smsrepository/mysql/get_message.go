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

func (m MysqlSMSRepository) GetByIdempotencyKey(ctx context.Context, idempotencyKey string) (*entity.Message, error) {
	const op = "MysqlSMSRepository.GetByIdempotencyKey"
	var message entity.Message

	row := m.adapter.Client().QueryRow("SELECT user_id, source_number, receptor_number, content, type, status FROM messages WHERE idempotency_key = ?", idempotencyKey)

	if err := row.Scan(&message.UserID, &message.SourceNumber, &message.ReceptorNumber, &message.Content, &message.Type, &message.Status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = errmsg.WrapMySQLError(fmt.Sprintf("get message by idempotency key %s", idempotencyKey), err)
			return &message, richerror.New(op).WithKind(richerror.KindNotFound).
				WithMessage(errmsg.ErrorMsgNotFound)
		}

		err = errmsg.WrapMySQLError(fmt.Sprintf("get message by idempotency key %s", idempotencyKey), err)
		return &message, richerror.New(op).WithKind(richerror.KindUnexpected).
			WithErr(err).WithMessage(errmsg.ErrorMsgSomethingWentWrong)
	}

	return &message, nil
}
