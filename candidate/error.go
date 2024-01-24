package candidate

import "errors"

var (
	// service errors
	ErrNoCandidateAvailable = errors.New("no available candidate")

	// app error
	ErrEmptyDB    = errors.New("cannot assign with empty db")
	ErrEmptyCache = errors.New("cannot assign with empty cache")

	// error for testing purpose
	ErrUnexpected = errors.New("unexpected error")
)
