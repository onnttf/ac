package controller

import "ac/util"

var (
	ErrSystemError  = util.NewError(1, "system error", "an unexpected system error occurred")
	ErrInvalidInput = util.NewError(2, "invalid input", "please check your input")
)
