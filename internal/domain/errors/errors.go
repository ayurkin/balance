package errors

import (
	"errors"
)

var (
	UnknownUserIdError        = errors.New("user_id does not exist")
	NotEnoughUserBalanceError = errors.New("user_id has not enough balance")
	DatabaseError             = errors.New("database error")
)
