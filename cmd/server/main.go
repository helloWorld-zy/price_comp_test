package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"cruise-price-compare/internal/app"
	httpTransport "cruise-price-compare/internal/transport/http"
)

func main() {
	// Load configuration
	config := app.LoadConfigFromEnv()

	// Create container
	container, err := app.NewContainer(config)
	if err != nil {
		panic(fmt.Sprintf("failed to create container: %v", err))
	}

	// Setup router
	router := httpTransport.SetupRouter(
		httpTransport.RouterConfig{Mode: config.ServerMode},
		container.JWTService,
		container.Logger,
		container.Metrics,
	)

	// Register routes
	httpTransport.RegisterRoutes(router, container.Handlers)

	// Create server
	addr := fmt.Sprintf("%s:%d", config.ServerHost, config.ServerPort)
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Setup graceful shutdown
	shutdown := app.NewGracefulShutdown(30 * time.Second)
	shutdown.AddHandler(func(ctx context.Context) error {
		container.Logger.Info("shutting down HTTP server...")
		return server.Shutdown(ctx)
	})
	shutdown.AddHandler(func(ctx context.Context) error {
		container.Logger.Info("closing database connection...")
		return container.Close()
	})

	// Start server in goroutine
	go func() {
		container.Logger.Info("starting HTTP server", "addr", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			container.Logger.Error("HTTP server error", "error", err)
		}
	}()

	// Wait for shutdown
	<-shutdown.WaitWithChannel()
	container.Logger.Info("server shutdown complete")
}
