package errors

import "errors"

var (
	ErrNotFound           = errors.New("not found")
	ErrAlreadyExists      = errors.New("already exists")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrInvalidInput       = errors.New("invalid input")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInternal           = errors.New("internal server error")
	ErrForbidden          = errors.New("forbidden")
	ErrInvalidTransition  = errors.New("invalid status transition")
	ErrOrderCompleted     = errors.New("order already completed")

	PGErrForeignKeyViolation = "23503"
	PGErrUniqueViolation     = "23505"
	PGErrInvalidTextRep      = "22P02"
	PGErrInvalidType         = "42804"
)
