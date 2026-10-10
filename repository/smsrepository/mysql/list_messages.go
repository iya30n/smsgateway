package mysql

import (
	"context"
	"smsgateway/entity"
	"smsgateway/pkg/errmsg"
	"smsgateway/pkg/richerror"
	"strings"
)

func (m MysqlSMSRepository) ListMessages(ctx context.Context, filter entity.MessageFilter) ([]entity.Message, int, error) {
	const op = "MysqlSMSRepository.ListMessages"

	conditions := make([]string, 0, 6)
	args := make([]interface{}, 0, 8)

	if filter.UserID != 0 {
		conditions = append(conditions, "user_id = ?")
		args = append(args, filter.UserID)
	}

	if filter.ReceptorNumber != "" {
		conditions = append(conditions, "receptor_number = ?")
		args = append(args, filter.ReceptorNumber)
	}

	if filter.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, filter.Status)
	}

	if filter.Type != "" {
		conditions = append(conditions, "type = ?")
		args = append(args, filter.Type)
	}

	if filter.CreatedFrom != 0 {
		conditions = append(conditions, "created_at >= ?")
		args = append(args, filter.CreatedFrom)
	}

	if filter.CreatedTo != 0 {
		conditions = append(conditions, "created_at < ?")
		args = append(args, filter.CreatedTo)
	}

	where := ""
	if len(conditions) > 0 {
		where = " WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	if err := m.adapter.Client().QueryRowContext(ctx, "SELECT COUNT(*) FROM messages"+where, args...).Scan(&total); err != nil {
		err = errmsg.WrapMySQLError("count messages", err)
		return nil, 0, richerror.New(op).WithKind(richerror.KindUnexpected).
			WithErr(err).WithMessage(errmsg.ErrorMsgSomethingWentWrong)
	}

	query := "SELECT id, user_id, source_number, receptor_number, content, type, status, failed_reason, created_at FROM messages" +
		where + " ORDER BY id DESC LIMIT ? OFFSET ?"

	rows, err := m.adapter.Client().QueryContext(ctx, query, append(args, filter.Limit, filter.Offset)...)
	if err != nil {
		err = errmsg.WrapMySQLError("list messages", err)
		return nil, 0, richerror.New(op).WithKind(richerror.KindUnexpected).
			WithErr(err).WithMessage(errmsg.ErrorMsgSomethingWentWrong)
	}
	defer rows.Close()

	messages := make([]entity.Message, 0, filter.Limit)
	for rows.Next() {
		var message entity.Message
		if err := rows.Scan(
			&message.ID,
			&message.UserID,
			&message.SourceNumber,
			&message.ReceptorNumber,
			&message.Content,
			&message.Type,
			&message.Status,
			&message.FailedReason,
			&message.CreatedAt,
		); err != nil {
			err = errmsg.WrapMySQLError("scan message", err)
			return nil, 0, richerror.New(op).WithKind(richerror.KindUnexpected).
				WithErr(err).WithMessage(errmsg.ErrorMsgSomethingWentWrong)
		}

		messages = append(messages, message)
	}

	if err := rows.Err(); err != nil {
		err = errmsg.WrapMySQLError("iterate messages", err)
		return nil, 0, richerror.New(op).WithKind(richerror.KindUnexpected).
			WithErr(err).WithMessage(errmsg.ErrorMsgSomethingWentWrong)
	}

	return messages, total, nil
}
