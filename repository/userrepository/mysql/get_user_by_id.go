package mysql

import (
	"database/sql"
	"errors"
	"fmt"
	"smsgateway/entity"
	"smsgateway/pkg/errmsg"
	"smsgateway/pkg/richerror"
)

func (m *MysqlUserRepository) GetUserByID(userID uint) (entity.User, error) {
	const op = "mysqluserrepo.GetUserByID"
	row := m.adapter.Client().QueryRow("SELECT * FROM users WHERE id = ?", userID)
	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = errmsg.WrapMySQLError(fmt.Sprintf("get user %s for update", userID), err)
			return user, richerror.New(op).WithKind(richerror.KindNotFound).
				WithMessage(err.Error())
		}

		return user, richerror.New(op).WithKind(richerror.KindUnexpected).
			WithErr(err).WithMessage(errmsg.ErrorMsgSomethingWentWrong).
			WithMeta(map[string]interface{}{"user_id": userID})
	}

	return user, nil
}

func scanUser(row *sql.Row) (entity.User, error) {
	user := entity.User{}
	err := row.Scan(&user.ID, &user.Name, &user.Username, &user.Balance)
	return user, err
}
