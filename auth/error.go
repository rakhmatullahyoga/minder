package auth

import "errors"

var (
	// handler errors
	ErrParseInput   = errors.New("invalid input format")
	ErrInvalidInput = errors.New("invalid parameter")

	// service errors
	ErrUserNotFound          = errors.New("user not found")
	ErrUserAlreadyRegistered = errors.New("user already registered")
	ErrIncorrectCredentials  = errors.New("wrong email or password")

	// app error
	ErrEmptyDB = errors.New("cannot assign with empty db")

	// error for testing purpose
	ErrUnexpected = errors.New("unexpected error")
)
