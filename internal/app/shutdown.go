package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// GracefulShutdown waits for shutdown signal and executes cleanup
type GracefulShutdown struct {
	timeout  time.Duration
	signals  []os.Signal
	handlers []func(context.Context) error
}

// NewGracefulShutdown creates a new graceful shutdown handler
func NewGracefulShutdown(timeout time.Duration) *GracefulShutdown {
	return &GracefulShutdown{
		timeout:  timeout,
		signals:  []os.Signal{syscall.SIGINT, syscall.SIGTERM},
		handlers: make([]func(context.Context) error, 0),
	}
}

// AddHandler adds a cleanup handler
func (g *GracefulShutdown) AddHandler(handler func(context.Context) error) {
	g.handlers = append(g.handlers, handler)
}

// Wait waits for shutdown signal and executes handlers
func (g *GracefulShutdown) Wait() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, g.signals...)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()

	// Execute handlers in reverse order (LIFO)
	for i := len(g.handlers) - 1; i >= 0; i-- {
		if err := g.handlers[i](ctx); err != nil {
			return err
		}
	}

	return nil
}

// WaitWithChannel waits for shutdown signal and sends to channel
func (g *GracefulShutdown) WaitWithChannel() <-chan struct{} {
	done := make(chan struct{})

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, g.signals...)
		<-quit

		ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
		defer cancel()

		// Execute handlers in reverse order (LIFO)
		for i := len(g.handlers) - 1; i >= 0; i-- {
			_ = g.handlers[i](ctx)
		}

		close(done)
	}()

	return done
}
