package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"smsgateway/entity"
	"smsgateway/pkg/errmsg"
	"smsgateway/pkg/richerror"

	"github.com/shopspring/decimal"
)

func (m MysqlSMSRepository) UpdateState(ctx context.Context, message entity.Message) error {
	const op = "mysqlsmsrepo.UpdateState"

	result, err := m.adapter.Client().ExecContext(ctx, `
		UPDATE messages
		SET status = ?, failed_reason = ?, updated_at = ?
		WHERE id = ?
	`, message.Status, message.FailedReason, time.Now().Unix(), message.ID)
	if err != nil {
		err = errmsg.WrapMySQLError(fmt.Sprintf("update message %d state", message.ID), err)
		return richerror.New(op).WithErr(err).WithKind(richerror.KindUnexpected).
			WithMessage(errmsg.ErrorMsgSomethingWentWrong)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		err = errmsg.WrapMySQLError(fmt.Sprintf("get affected rows for message %d", message.ID), err)
		return richerror.New(op).WithErr(err).WithKind(richerror.KindUnexpected).
			WithMessage(errmsg.ErrorMsgSomethingWentWrong)
	}

	if affected == 0 {
		return richerror.New(op).WithKind(richerror.KindNotFound).
			WithMessage(errmsg.ErrorMsgNotFound)
	}

	return nil
}

func (m MysqlSMSRepository) UpdateStateToFailed(ctx context.Context, message entity.Message) error {
	const op = "mysqlsmsrepo.UpdateStateToFailed"

	message.Status = entity.MessageStatusFailed

	tx, err := m.adapter.Client().BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		err = errmsg.WrapMySQLError("begin update message transaction", err)
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

	var wageTrxAmount decimal.Decimal
	if err := tx.QueryRowContext(ctx, "SELECT amount FROM transactions WHERE message_id = ?", message.ID).Scan(&wageTrxAmount); err != nil {
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

	if _, err := tx.Exec("UPDATE users SET balance = balance + ? WHERE id = ?", wageTrxAmount, message.UserID); err != nil {
		err = errmsg.WrapMySQLError(fmt.Sprintf("increase user %d balance", message.UserID), err)
		return richerror.New(op).WithKind(richerror.KindUnexpected).
			WithErr(err).WithMessage(err.Error())
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE messages
		SET status = ?, failed_reason = ?, updated_at = ?
		WHERE id = ?
	`, message.Status, message.FailedReason, time.Now().Unix(), message.ID); err != nil {
		err = errmsg.WrapMySQLError(fmt.Sprintf("update message %d state", message.ID), err)
		return richerror.New(op).WithErr(err).WithKind(richerror.KindUnexpected).
			WithMessage(errmsg.ErrorMsgSomethingWentWrong)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO transactions (user_id, message_id, amount, type)
		VALUES (?,?,?,?)
	`, message.UserID, message.ID, wageTrxAmount, entity.TransactionTypeReversal); err != nil {
		err = errmsg.WrapMySQLError(fmt.Sprintf("insert reversal transaction for user %d", message.UserID), err)
		return richerror.New(op).WithKind(richerror.KindUnexpected).
			WithErr(err).WithMessage(err.Error())
	}

	if err := tx.Commit(); err != nil {
		return errmsg.WrapMySQLError("commit empty updating message transaction", err)
	} else {
		committed = true
	}

	return nil
}
