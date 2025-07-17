package log_test

import (
	"os"

	"github.com/aca/qwer/log"
)

func Example() {
	// Basic usage - works just like slog
	log.Info("hello world", "count", 3)

	// With structured logging
	log.Info("user login", 
		"user", "john",
		"ip", "192.168.1.1",
	)

	// With context values
	logger := log.With("service", "api")
	logger.Info("request handled", "method", "GET", "path", "/users")

	// With groups
	log.WithGroup("http").Info("request",
		"method", "POST",
		"url", "/api/users",
		"status", 200,
	)
}

func Example_withAddSource() {
	// Enable source location in logs
	handler := log.NewTextHandler(os.Stdout, &log.HandlerOptions{
		AddSource: true,
		Level:     log.LevelDebug,
	})
	logger := log.NewLogger(handler)
	log.SetDefault(logger)

	// This will include the source file, line number, and function name
	log.Info("message with source location")
	log.Debug("debug message with source")
}

func ExampleJSONHandler() {
	// Use JSON format for structured logging
	handler := log.NewJSONHandler(os.Stdout, &log.HandlerOptions{
		AddSource: true,
	})
	logger := log.NewLogger(handler)
	log.SetDefault(logger)

	log.Info("structured log entry",
		"user_id", 123,
		"action", "purchase",
		"amount", 99.99,
	)
}
