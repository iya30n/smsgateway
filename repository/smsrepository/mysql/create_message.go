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

// CreateMessage charges the user's balance and persists the message in one
// transaction. The caller gets the assigned ID back through the pointer so it
// can be carried into the queue payload.
func (m MysqlSMSRepository) CreateMessage(ctx context.Context, message *entity.Message) error {
	const op = "mysqlsmsrepo.CreateMessage"

	tx, err := m.adapter.Client().BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		err = errmsg.WrapMySQLError("begin create message transaction", err)
		return richerror.New(op).WithErr(err).WithKind(richerror.KindUnexpected).
			WithMessage(errmsg.ErrorMsgSomethingWentWrong)
	}

	committed := false
	defer func() {
		if committed {
			return
		}
		if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
			// TODO: put the error into logger service
		}
	}()

	var smsWageAmount float64
	if err := tx.QueryRow("SELECT amount FROM wages WHERE (user_id = ? AND type = ?) OR (type = ?)", message.UserID, message.Type, message.Type).Scan(&smsWageAmount); err != nil {
		return richerror.New(op).WithKind(richerror.KindUnexpected).
			WithErr(err).WithMessage(errmsg.ErrorMsgSomethingWentWrong)
	}

	if err := tx.QueryRowContext(ctx, `SELECT id, balance FROM users WHERE id = ? FOR UPDATE`, message.UserID).Scan(&message.UserID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return richerror.New(op).WithKind(richerror.KindNotFound).
				WithMessage(errmsg.ErrorMsgNotFound)
		}

		err = errmsg.WrapMySQLError(fmt.Sprintf("get user %d for update", message.UserID), err)
		return richerror.New(op).WithErr(err).WithKind(richerror.KindUnexpected).
			WithMessage(err.Error())
	}

	if _, err := tx.Exec("UPDATE users SET balance = balance - ? WHERE id = ?", smsWageAmount, message.UserID); err != nil {
		err = errmsg.WrapMySQLError(fmt.Sprintf("decrease user %d balance", message.UserID), err)
		return richerror.New(op).WithKind(richerror.KindUnexpected).
			WithErr(err).WithMessage(err.Error())
	}

	result, err := tx.ExecContext(ctx, `
		INSERT INTO messages (user_id, idempotency_key, source_number, receptor_number, content, type, status)
		VALUES (?,?,?,?,?,?,?)
	`, message.UserID, message.IdempotencyKey, message.SourceNumber, message.ReceptorNumber, message.Content, message.Type, message.Status)
	if err != nil {
		err = errmsg.WrapMySQLError(fmt.Sprintf("insert message for user %d", message.UserID), err)
		return richerror.New(op).WithKind(richerror.KindUnexpected).
			WithErr(err).WithMessage(err.Error())
	}

	messageID, err := result.LastInsertId()
	if err != nil {
		err = errmsg.WrapMySQLError(fmt.Sprintf("get inserted message id for user %d", message.UserID), err)
		return richerror.New(op).WithKind(richerror.KindUnexpected).
			WithErr(err).WithMessage(errmsg.ErrorMsgSomethingWentWrong)
	}

	message.ID = uint(messageID)

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO transactions (user_id, message_id, amount, type)
		VALUES (?,?,?,?)
	`, message.UserID, message.ID, smsWageAmount, entity.TransactionTypeSmsCharge); err != nil {
		err = errmsg.WrapMySQLError(fmt.Sprintf("insert transaction for user %d", message.UserID), err)
		return richerror.New(op).WithKind(richerror.KindUnexpected).
			WithErr(err).WithMessage(err.Error())
	}

	if err := tx.Commit(); err != nil {
		return errmsg.WrapMySQLError("commit empty creating message transaction", err)
	} else {
		committed = true
	}

	return nil
}
