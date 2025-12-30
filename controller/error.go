package controller

import "ac/util"

var (
	// ErrSystemError represents an unexpected internal error.
	ErrSystemError = util.NewError(1, "system error", "an unexpected system error occurred")
	// ErrInvalidInput indicates the request contains invalid parameters.
	ErrInvalidInput = util.NewError(2, "invalid input", "please check your input")
	// ErrAlreadyExists indicates that a record already exists.
	ErrAlreadyExists = util.NewError(3, "already exists", "record already exists")
	// ErrNotFound indicates that a record does not exist.
	ErrNotFound = util.NewError(4, "not found", "record not found")
)
