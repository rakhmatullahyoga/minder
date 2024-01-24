package candidate

import "errors"

var (
	// handler errors
	ErrInvalidInput = errors.New("invalid parameter")

	// service errors
	ErrNoCandidateAvailable = errors.New("no available candidate")
	ErrExceedQuota          = errors.New("already reach daily quota")
	ErrAlreadySwiped        = errors.New("candidate already swiped")
	ErrNoCachedCandidate    = errors.New("no cached candidate")
	ErrAlreadyVerified      = errors.New("already subscribe premium")

	// app error
	ErrEmptyDB    = errors.New("cannot assign with empty db")
	ErrEmptyCache = errors.New("cannot assign with empty cache")

	// error for testing purpose
	ErrUnexpected = errors.New("unexpected error")
)
