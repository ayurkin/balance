package errors

import "fmt"

type UserError struct {
	UserId int64
}

type UnknownUserIdError struct {
	UserError
}

func (e UnknownUserIdError) Error() string {
	return fmt.Sprintf("user_id %d does not exist", e.UserId)
}

type NotEnoughUserBalance struct {
	UserError
}

func (e NotEnoughUserBalance) Error() string {
	return fmt.Sprintf("user_id %d has not enough balance", e.UserId)
}
