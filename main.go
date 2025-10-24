package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ac/bootstrap"
	"ac/controller"
	"ac/controller/user"
	"ac/middleware"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"

	_ "ac/docs"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @Title           Access Control
// @Version         1.0
// @Description     This is the API documentation for Access Control.

// @Contact.Name   Zhang Peng
// @Contact.Url    https://github.com/onnttf
// @Contact.Email  onnttf@gmail.com

// @License.Name  MIT
// @License.Url   https://opensource.org/licenses/MIT

// @Host      localhost:8082
// @BasePath  /
// @Schemes   http https
// @Accept    json
// @Produce   json
func main() {
	if err := bootstrap.Initialize(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: main: failed to initialize application, error=%v\n", err)
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	router := gin.New()
	router.Use(middleware.Logger(), gin.Recovery(), requestid.New())
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	router.GET("/ping", func(ctx *gin.Context) {
		controller.Success(ctx, map[string]any{
			"message":   "pong",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	innerApi := router.Group("/internal-api")
	user.RegisterInternalRoutes(innerApi)

	srv := &http.Server{
		Addr:         ":8082",
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		fmt.Fprintf(os.Stdout, "INFO: main: http server starting on address=%s\n", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Fprintf(os.Stderr, "ERROR: main: http server failed to start, error=%v\n", err)
		}
	}()

	<-ctx.Done()

	fmt.Fprintf(os.Stdout, "INFO: main: graceful shutdown initiated\n")
	fmt.Fprintf(os.Stdout, "WARN: main: press ctrl+c again for immediate shutdown\n")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: main: graceful shutdown failed, error=%v\n", err)
	} else {
		fmt.Fprintf(os.Stdout, "INFO: main: server gracefully stopped\n")
	}

	fmt.Fprintf(os.Stdout, "INFO: main: application shutdown complete\n")
}
