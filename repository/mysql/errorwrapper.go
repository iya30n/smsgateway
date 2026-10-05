package mysql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
)

var (
	ErrDuplicatedPayment = errors.New("duplicated payment_consumer")
	ErrNoIDFound         = errors.New("no id found")
	ErrConflict          = errors.New("conflict")
	ErrUnavailable       = errors.New("service unavailable")
	ErrContextDeadline   = errors.New("context deadline exceeded")
	ErrNotFound          = errors.New("not found")
)

func WrapMySQLError(operation string, err error) error {
	if err == nil {
		return nil
	}

	var mysqlErr *mysql.MySQLError

	if errors.As(err, &mysqlErr) {
		switch mysqlErr.Number {
		case 1062:
			return fmt.Errorf(
				"%s: %w",
				operation,
				ErrConflict,
			)

		case 1451:
			return fmt.Errorf(
				"%s: %w",
				operation,
				ErrConflict,
			)

		case 1040, 1205, 1213:
			return fmt.Errorf(
				"%s: %w",
				operation,
				ErrUnavailable,
			)
		}
	}

	if errors.Is(err, driver.ErrBadConn) {
		return fmt.Errorf(
			"%s: %w",
			operation,
			ErrUnavailable,
		)
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf(
			"%s: %w",
			operation,
			ErrContextDeadline,
		)
	}

	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf(
			"%s: %w",
			operation,
			ErrNotFound,
		)
	}

	return fmt.Errorf("%s: %w", operation, err)
}
