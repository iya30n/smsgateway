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

func (m *MysqlUserRepository) IncreaseBalance(ctx context.Context, userID uint, amount float64) (float64, error) {
	const op = "mysqluserrepo.IncreaseBalance"

	var newBalance float64

	tx, err := m.adapter.Client().BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		err = errmsg.WrapMySQLError("begin increase balance transaction", err)
		return 0, richerror.New(op).WithErr(err).WithKind(richerror.KindUnexpected).
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

	if err := tx.QueryRowContext(ctx, `SELECT id, balance FROM users WHERE id = ? FOR UPDATE`, userID).Scan(&userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return newBalance, richerror.New(op).WithKind(richerror.KindNotFound).
				WithMessage(errmsg.ErrorMsgNotFound)
		}

		err = errmsg.WrapMySQLError(fmt.Sprintf("get user %s for update", userID), err)
		return 0, richerror.New(op).WithErr(err).WithKind(richerror.KindUnexpected).
			WithMessage(err.Error())
	}

	if _, err := tx.Exec("UPDATE users SET balance = balance + ? WHERE id = ?", amount, userID); err != nil {
		err = errmsg.WrapMySQLError(fmt.Sprintf("increase user %s balance", userID), err)
		return newBalance, richerror.New(op).WithKind(richerror.KindUnexpected).
			WithErr(err).WithMessage(err.Error())
	}

	if _, err := tx.Exec("INSERT INTO transactions (user_id, credit, type) VALUES (?, ?, ?)", userID, amount, entity.TransactionTypeManualAdjustment); err != nil {
		return newBalance, richerror.New(op).WithKind(richerror.KindUnexpected).
			WithErr(err).WithMessage(errmsg.ErrorMsgSomethingWentWrong)
	}

	if err := tx.Commit(); err != nil {
		return 0, errmsg.WrapMySQLError("commit empty increase balance transaction", err)
	} else {
		committed = true
	}

	if err := m.adapter.Client().QueryRow("SELECT balance FROM users WHERE id = ?", userID).Scan(&newBalance); err != nil {
		return 0, richerror.New(op).WithKind(richerror.KindUnexpected).
			WithErr(err).WithMessage(errmsg.ErrorMsgSomethingWentWrong)
	}

	return newBalance, nil
}
