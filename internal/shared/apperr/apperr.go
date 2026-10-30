package apperr

import (
	"errors"
	"fmt"
	"net/http"
)

// Error is an application error carrying a stable business code and the HTTP
// status that should be returned to the client.
type Error struct {
	Code    int
	Status  int
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *Error) Unwrap() error { return e.Err }

func New(code, status int, message string) *Error {
	return &Error{Code: code, Status: status, Message: message}
}

func Wrap(err error, code, status int, message string) *Error {
	return &Error{Code: code, Status: status, Message: message, Err: err}
}

var (
	ErrBadRequest   = New(40000, http.StatusBadRequest, "invalid request")
	ErrUnauthorized = New(40100, http.StatusUnauthorized, "unauthorized")
	ErrForbidden    = New(40300, http.StatusForbidden, "forbidden")
	ErrNotFound     = New(40400, http.StatusNotFound, "not found")
	ErrConflict     = New(40900, http.StatusConflict, "conflict")
	ErrInternal     = New(50000, http.StatusInternalServerError, "internal server error")
)

func BadRequest(message string) *Error   { return New(40000, http.StatusBadRequest, message) }
func Unauthorized(message string) *Error { return New(40100, http.StatusUnauthorized, message) }
func Forbidden(message string) *Error    { return New(40300, http.StatusForbidden, message) }
func NotFound(message string) *Error     { return New(40400, http.StatusNotFound, message) }
func Internal(err error) *Error {
	return Wrap(err, 50000, http.StatusInternalServerError, "internal server error")
}

// As converts any error into an *Error. Unknown errors are treated as internal
// errors so their details are never exposed to clients.
func As(err error) *Error {
	var target *Error
	if errors.As(err, &target) {
		return target
	}
	return Internal(err)
}
