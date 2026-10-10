package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"smsgateway/entity"
	"smsgateway/pkg/errmsg"
	"smsgateway/pkg/logger"
	"smsgateway/pkg/richerror"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

func (m *MysqlUserRepository) IncreaseBalance(ctx context.Context, userID uint, amount decimal.Decimal) (decimal.Decimal, error) {
	const op = "mysqluserrepo.IncreaseBalance"

	var newBalance decimal.Decimal

	tx, err := m.adapter.Client().BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		err = errmsg.WrapMySQLError("begin increase balance transaction", err)
		return decimal.Zero, richerror.New(op).WithErr(err).WithKind(richerror.KindUnexpected).
			WithMessage(errmsg.ErrorMsgSomethingWentWrong)
	}

	committed := false
	defer func() {
		if committed {
			return
		}
		if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
			logger.Logger.Error("failed to roll back increase balance transaction",
				zap.String("op", op),
				zap.Error(rollbackErr),
			)
		}
	}()

	if err := tx.QueryRowContext(ctx, `SELECT id, balance FROM users WHERE id = ? FOR UPDATE`, userID).Scan(&userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return newBalance, richerror.New(op).WithKind(richerror.KindNotFound).
				WithMessage(errmsg.ErrorMsgNotFound)
		}

		err = errmsg.WrapMySQLError(fmt.Sprintf("get user %d for update", userID), err)
		return decimal.Zero, richerror.New(op).WithErr(err).WithKind(richerror.KindUnexpected).
			WithMessage(err.Error())
	}

	if _, err := tx.Exec("UPDATE users SET balance = balance + ? WHERE id = ?", amount, userID); err != nil {
		err = errmsg.WrapMySQLError(fmt.Sprintf("increase user %d balance", userID), err)
		return newBalance, richerror.New(op).WithKind(richerror.KindUnexpected).
			WithErr(err).WithMessage(err.Error())
	}

	if _, err := tx.ExecContext(ctx, "INSERT INTO transactions (user_id, amount, type) VALUES (?, ?, ?)", userID, amount, entity.TransactionTypeManualAdjustment); err != nil {
		err = errmsg.WrapMySQLError(fmt.Sprintf("insert manual adjustment transaction for user %d", userID), err)
		return newBalance, richerror.New(op).WithKind(richerror.KindUnexpected).
			WithErr(err).WithMessage(errmsg.ErrorMsgSomethingWentWrong)
	}

	if err := tx.Commit(); err != nil {
		return decimal.Zero, errmsg.WrapMySQLError("commit empty increase balance transaction", err)
	} else {
		committed = true
	}

	if err := m.adapter.Client().QueryRow("SELECT balance FROM users WHERE id = ?", userID).Scan(&newBalance); err != nil {
		return decimal.Zero, richerror.New(op).WithKind(richerror.KindUnexpected).
			WithErr(err).WithMessage(errmsg.ErrorMsgSomethingWentWrong)
	}

	return newBalance, nil
}

func (m *MysqlUserRepository) HasEnoughBalanceForSMS(userID uint, smsType entity.SmsType) (bool, error) {
	const op = "mysqluserrepo.HasEnoughBalanceForSMS"

	var balance decimal.Decimal
	if err := m.adapter.Client().QueryRow("SELECT balance FROM users WHERE id = ?", userID).Scan(&balance); err != nil {
		err = errmsg.WrapMySQLError(fmt.Sprintf("get user %d for update", userID), err)
		return false, richerror.New(op).WithKind(richerror.KindUnexpected).
			WithErr(err).WithMessage(errmsg.ErrorMsgSomethingWentWrong)
	}

	var smsWageAmount decimal.Decimal
	if err := m.adapter.Client().QueryRow("SELECT amount FROM wages WHERE (user_id = ? AND type = ?) OR (type = ?)", userID, smsType, entity.SmsTypeExpress).Scan(&smsWageAmount); err != nil {
		err = errmsg.WrapMySQLError(fmt.Sprintf("get user %d for update", userID), err)
		return false, richerror.New(op).WithKind(richerror.KindUnexpected).
			WithErr(err).WithMessage(errmsg.ErrorMsgSomethingWentWrong)
	}

	return balance.GreaterThanOrEqual(smsWageAmount), nil
}
