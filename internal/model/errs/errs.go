package errs

import "errors"

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
)
