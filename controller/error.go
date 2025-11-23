package controller

import "ac/util"

var (
	// ErrSystemError represents an unexpected internal error.
	ErrSystemError = util.NewError(1, "system error", "an unexpected system error occurred")
	// ErrInvalidInput indicates the request contains invalid parameters.
	ErrInvalidInput = util.NewError(2, "invalid input", "please check your input")
)
