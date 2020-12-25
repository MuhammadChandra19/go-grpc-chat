package errors

import "errors"

type Error struct {
	Code string
	Err  error
}

const (
	CodeSystemError     = "system_error"
	CodeValidationError = "validation_error"
	CodeNotFoundError   = "not-found"
	CodeNotAuthorized   = "not-authorized"
)

func (e *Error) Error() string {
	return e.Err.Error()
}

// N to build an error object, that contains code and its message
func N(code, message string) error {
	return &Error{
		Code: code,
		Err:  errors.New(message),
	}
}

// Is to check if an error is an *Error from given code
func Is(code string, err error) bool {
	e, ok := err.(*Error)
	if !ok {
		return false
	}

	if e.Code != code {
		return false
	}

	return true
}

// GetCode func to get error code
func GetCode(err error) string {
	e, ok := err.(*Error)
	if !ok {
		return "undefined-error"
	}

	return e.Code
}
