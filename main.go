package main

import (
	"ac/bootstrap"
	"ac/bootstrap/logger"
	"ac/controller/resource"
	"ac/controller/role"
	"ac/controller/system"
	"ac/controller/user"
	"ac/custom/define"
	"ac/custom/validator"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Initialize the system
	if err := bootstrap.Initialize(); err != nil {
		panic(fmt.Errorf("failed to initialize: %w", err))
	}

	e := echo.New()
	e.HideBanner = true
	e.Validator = validator.NewCustomValidator()
	e.Use(
		middleware.RequestID(),
		middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
			LogStatus:   true,
			LogLatency:  true,
			LogRemoteIP: true,
			LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
				kv := map[string]any{
					"type":       "access",
					"latency":    v.Latency,
					"remote_ip":  v.RemoteIP,
					"status":     v.Status,
					"method":     c.Request().Method,
					"uri":        c.Request().RequestURI,
					"request_id": c.Response().Header().Get(echo.HeaderXRequestID),
				}
				if v.Error == nil {
					logger.LogWith(logger.LevelInfo, "success", kv)
				} else {
					kv["error"] = v.Error.Error()
					logger.LogWith(logger.LevelError, "failure", kv)
				}
				return v.Error
			},
		}),
		middleware.Recover(),
	)

	// Middleware to add request ID to context
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if val := c.Response().Header().Get(echo.HeaderXRequestID); val != "" {
				req := c.Request()
				ctx := context.WithValue(req.Context(), define.RequestIDKey, val)
				c.SetRequest(req.WithContext(ctx))
			}
			return next(c)
		}
	})

	api := e.Group("/api")
	system.RegisterRoutes(api.Group("/system"))
	user.RegisterRoutes(api.Group("/user"))
	role.RegisterRoutes(api.Group("/role"))
	resource.RegisterRoutes(api.Group("/resource"))

	// Output all routes
	printRoutes(e)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Start server
	go func() {
		if err := e.Start(":1323"); err != nil && err != http.ErrServerClosed {
			panic(fmt.Errorf("failed to start server: %w", err))
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server with a timeout of 10 seconds.
	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		panic(fmt.Errorf("failed to shutdown server: %w", err))
	}
}

// printRoutes outputs all registered routes to the console, sorted by path.
func printRoutes(e *echo.Echo) {
	routes := e.Routes()

	sort.Slice(routes, func(i, j int) bool {
		return routes[i].Path < routes[j].Path
	})

	fmt.Println("Registered Routes:")
	for _, route := range routes {
		fmt.Printf("Method: %-6s | Path: %-30s\n", route.Method, route.Path)
	}
}
