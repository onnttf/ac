package middleware

import (
	"time"

	"ac/bootstrap/logger"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap/zapcore"
)

func Logger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		startTime := time.Now()
		ctx.Next()

		status := ctx.Writer.Status()
		latency := time.Since(startTime).Milliseconds()

		level := zapcore.InfoLevel
		msg := "http request finished"
		fields := map[string]any{
			"client_ip":    ctx.ClientIP(),
			"latency_ac":   latency,
			"status":       status,
			"method":       ctx.Request.Method,
			"uri":          ctx.Request.RequestURI,
			"request_id":   requestid.Get(ctx),
			"user_agent":   ctx.Request.UserAgent(),
			"content_type": ctx.Request.Header.Get("Content-Type"),
		}

		logger.LogWith(ctx, level, msg, fields)
	}
}
