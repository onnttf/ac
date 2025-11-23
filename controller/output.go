package controller

import (
	"errors"
	"net/http"

	"ac/util"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

// Response defines a standard JSON API response.
type Response struct {
	Code      int    `json:"code"`
	RequestId string `json:"request_id"`
	Msg       string `json:"msg"`
	Hint      string `json:"hint,omitempty"`
	Data      any    `json:"data" swaggertype:"object"`
	Err       string `json:"err,omitempty"`
}

// Success returns a standard success response with optional data.
func Success(ctx *gin.Context, data any) {
	if data == nil {
		data = struct{}{}
	}

	ctx.JSON(http.StatusOK, Response{
		Code:      0,
		RequestId: requestid.Get(ctx),
		Msg:       "success",
		Data:      data,
	})
}

// Failure maps an error to a standard response with code and message.
func Failure(ctx *gin.Context, err error) {
	response := Response{
		Code:      ErrSystemError.Code,
		Msg:       ErrSystemError.Msg,
		Hint:      ErrSystemError.Hint,
		RequestId: requestid.Get(ctx),
		Data:      struct{}{},
	}

	var customErr *util.Error
	if errors.As(err, &customErr) {
		if customErr.Code != 0 {
			response.Code = customErr.Code
		}
		if customErr.Msg != "" {
			response.Msg = customErr.Msg
		}
		if customErr.Hint != "" {
			response.Hint = customErr.Hint
		}
		if customErr.Cause != nil {
			response.Err = customErr.Cause.Error()
		}
	} else if err != nil {
		response.Err = err.Error()
	}

	ctx.JSON(http.StatusOK, response)
}
