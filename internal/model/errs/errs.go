package errs

import (
	"errors"
	"fmt"
)

var (
	ErrUserAlreadyExists        = errors.New("user already exists")
	ErrUserNotFound             = errors.New("user not found")
	ErrAuthFailed               = errors.New("auth failed")
	ErrOrderNumberInvalid       = errors.New("order number invalid")
	ErrOrderNumberUnprocessable = errors.New("order number unprocessable")
	ErrOrderAlreadyUploaded     = errors.New("order already uploaded")
	ErrOrderAlreadyExists       = errors.New("order already exists")
	ErrOrderNotFound            = errors.New("order not found")
	ErrOrderConflict            = errors.New("order conflict")
	ErrBalanceNotEnough         = errors.New("balance not enough")
	ErrAccrualNotRegistered     = errors.New("accrual not registered")
)

type ErrAccrualTooManyRequests struct {
	RetryAfter int
	Body       string
}

func (e ErrAccrualTooManyRequests) Error() string {
	return fmt.Sprintf("Retry after %d seconds, %s", e.RetryAfter, e.Body)
}
