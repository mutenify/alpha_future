package domain

import "errors"

var (
	ErrNotFound            = errors.New("not found")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")
	ErrConflict            = errors.New("conflict")
	ErrInvalidTransition   = errors.New("invalid status transition")
	ErrIdempotencyMismatch = errors.New("idempotency key reused with different body")
	ErrSourceUnavailable   = errors.New("external source unavailable")
)

// ValidationError — ошибка входных данных, текст показывается пользователю.
type ValidationError struct{ Msg string }

func (e ValidationError) Error() string { return e.Msg }

func Invalid(msg string) error { return ValidationError{Msg: msg} }
