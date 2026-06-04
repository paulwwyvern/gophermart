package passwordhash

import "golang.org/x/crypto/bcrypt"

type Error struct {
	error error
}

func NewError(error error) *Error {
	return &Error{error: error}
}

func (e *Error) Error() string {
	return e.error.Error()
}

func (e *Error) Unwrap() error {
	return e.error
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return "", NewError(err)
	}
	return string(bytes), nil
}

func VerifyPassword(hash, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return NewError(err)
	}
	return nil
}
