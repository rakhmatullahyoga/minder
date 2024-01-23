package auth

import "errors"

var (
	// handler errors
	ErrParseInput = errors.New("invalid input format")

	// service errors
	ErrUserAlreadyRegistered = errors.New("user already registered")

	// app error
	ErrEmptyDB = errors.New("cannot assign with empty db")
)
