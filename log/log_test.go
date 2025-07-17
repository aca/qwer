package log

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

func TestBasicLogging(t *testing.T) {
	var buf bytes.Buffer
	handler := NewTextHandler(&buf, nil)
	logger := NewLogger(handler)
	SetDefault(logger)

	Info("test message", "key", "value")
	output := buf.String()

	if !strings.Contains(output, "test message") {
		t.Errorf("expected output to contain 'test message', got: %s", output)
	}
	if !strings.Contains(output, "key=value") {
		t.Errorf("expected output to contain 'key=value', got: %s", output)
	}
}

func TestWithContext(t *testing.T) {
	var buf bytes.Buffer
	handler := NewTextHandler(&buf, nil)
	logger := NewLogger(handler)
	SetDefault(logger)

	ctx := context.WithValue(context.Background(), "testKey", "testValue")
	InfoContext(ctx, "context message", "key", "value")
	output := buf.String()

	if !strings.Contains(output, "context message") {
		t.Errorf("expected output to contain 'context message', got: %s", output)
	}
}

func TestAllLevels(t *testing.T) {
	var buf bytes.Buffer
	handler := NewTextHandler(&buf, &HandlerOptions{Level: LevelDebug})
	logger := NewLogger(handler)
	SetDefault(logger)

	Debug("debug message")
	Info("info message")
	Warn("warn message")
	Error("error message")

	output := buf.String()
	levels := []string{"DEBUG", "INFO", "WARN", "ERROR"}
	messages := []string{"debug message", "info message", "warn message", "error message"}

	for i, level := range levels {
		if !strings.Contains(output, level) {
			t.Errorf("expected output to contain '%s', got: %s", level, output)
		}
		if !strings.Contains(output, messages[i]) {
			t.Errorf("expected output to contain '%s', got: %s", messages[i], output)
		}
	}
}

func TestAddSource(t *testing.T) {
	var buf bytes.Buffer
	handler := NewJSONHandler(&buf, &HandlerOptions{AddSource: true})
	logger := NewLogger(handler)
	SetDefault(logger)

	Info("test with source")
	
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("failed to unmarshal log entry: %v", err)
	}

	source, ok := logEntry["source"].(map[string]interface{})
	if !ok {
		t.Fatal("expected source field in log entry")
	}

	if source["function"] == nil {
		t.Error("expected function field in source")
	}
	if source["file"] == nil {
		t.Error("expected file field in source")
	}
	if source["line"] == nil {
		t.Error("expected line field in source")
	}

	file, _ := source["file"].(string)
	if !strings.Contains(file, "log_test.go") {
		t.Errorf("expected source file to contain 'log_test.go', got: %s", file)
	}
}

func TestWithMethods(t *testing.T) {
	var buf bytes.Buffer
	handler := NewTextHandler(&buf, nil)
	logger := NewLogger(handler)
	SetDefault(logger)

	With("global", "value").Info("message with global", "local", "value2")
	output := buf.String()

	if !strings.Contains(output, "global=value") {
		t.Errorf("expected output to contain 'global=value', got: %s", output)
	}
	if !strings.Contains(output, "local=value2") {
		t.Errorf("expected output to contain 'local=value2', got: %s", output)
	}
}

func TestWithGroup(t *testing.T) {
	var buf bytes.Buffer
	handler := NewJSONHandler(&buf, nil)
	logger := NewLogger(handler)
	SetDefault(logger)

	WithGroup("request").Info("grouped message", "method", "GET", "path", "/api")
	
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("failed to unmarshal log entry: %v", err)
	}

	request, ok := logEntry["request"].(map[string]interface{})
	if !ok {
		t.Fatal("expected request group in log entry")
	}

	if request["method"] != "GET" {
		t.Errorf("expected method to be 'GET', got: %v", request["method"])
	}
	if request["path"] != "/api" {
		t.Errorf("expected path to be '/api', got: %v", request["path"])
	}
}

func TestLogAttrs(t *testing.T) {
	var buf bytes.Buffer
	handler := NewTextHandler(&buf, nil)
	logger := NewLogger(handler)
	SetDefault(logger)

	LogAttrs(context.Background(), LevelInfo, "attrs message",
		String("key1", "value1"),
		Int("key2", 42),
	)

	output := buf.String()
	if !strings.Contains(output, "attrs message") {
		t.Errorf("expected output to contain 'attrs message', got: %s", output)
	}
	if !strings.Contains(output, "key1=value1") {
		t.Errorf("expected output to contain 'key1=value1', got: %s", output)
	}
	if !strings.Contains(output, "key2=42") {
		t.Errorf("expected output to contain 'key2=42', got: %s", output)
	}
}

func TestCompatibilityWithSlog(t *testing.T) {
	var logBuf, slogBuf bytes.Buffer
	
	// Test with log
	logHandler := NewJSONHandler(&logBuf, &HandlerOptions{AddSource: true})
	_ = logHandler
	logLogger := NewLogger(logHandler)
	SetDefault(logLogger)
	Info("test message", "key", "value")
	
	// Test with slog
	slogHandler := slog.NewJSONHandler(&slogBuf, &slog.HandlerOptions{AddSource: true})
	slogLogger := slog.New(slogHandler)
	slog.SetDefault(slogLogger)
	slog.Info("test message", "key", "value")
	
	var logEntry, slogEntry map[string]interface{}
	if err := json.Unmarshal(logBuf.Bytes(), &logEntry); err != nil {
		t.Fatalf("failed to unmarshal log entry: %v", err)
	}
	if err := json.Unmarshal(slogBuf.Bytes(), &slogEntry); err != nil {
		t.Fatalf("failed to unmarshal slog entry: %v", err)
	}
	
	// Check that both have the same structure
	if logEntry["msg"] != slogEntry["msg"] {
		t.Errorf("messages don't match: log=%v, slog=%v", logEntry["msg"], slogEntry["msg"])
	}
	if logEntry["key"] != slogEntry["key"] {
		t.Errorf("key values don't match: log=%v, slog=%v", logEntry["key"], slogEntry["key"])
	}
	
	// Both should have source information
	if _, ok := logEntry["source"]; !ok {
		t.Error("log entry missing source field")
	}
	if _, ok := slogEntry["source"]; !ok {
		t.Error("slog entry missing source field")
	}
}

func TestPrintfFunctions(t *testing.T) {
	var buf bytes.Buffer
	handler := NewTextHandler(&buf, nil)
	logger := NewLogger(handler)
	SetDefault(logger)

	// Test Print, Printf, Println
	Print("Hello", " ", "World")
	Printf("Hello %s, you are %d years old", "Alice", 25)
	Println("This", "is", "a", "test")
	
	output := buf.String()
	
	// Check Print output
	if !strings.Contains(output, "Hello World") {
		t.Errorf("expected output to contain 'Hello World', got: %s", output)
	}
	
	// Check Printf output
	if !strings.Contains(output, "Hello Alice, you are 25 years old") {
		t.Errorf("expected output to contain 'Hello Alice, you are 25 years old', got: %s", output)
	}
	
	// Check Println output (should have spaces between words and newline)
	if !strings.Contains(output, "This is a test") {
		t.Errorf("expected output to contain 'This is a test', got: %s", output)
	}
}

func TestLevelSpecificPrintfFunctions(t *testing.T) {
	var buf bytes.Buffer
	handler := NewTextHandler(&buf, &HandlerOptions{Level: LevelTrace})
	logger := NewLogger(handler)
	SetDefault(logger)

	// Test all level-specific printf functions
	Tracef("Trace: %s", "debugging")
	Debugf("Debug: %d", 42)
	Infof("Info: %v", []string{"a", "b"})
	Warnf("Warn: %t", true)
	Errorf("Error: %f", 3.14)
	Emergencyf("Emergency: %s", "critical")
	
	output := buf.String()
	
	// Check that all messages are present
	expectedMessages := []string{
		"Trace: debugging",
		"Debug: 42", 
		"Info: [a b]",
		"Warn: true",
		"Error: 3.14",
		"Emergency: critical",
	}
	
	for _, expected := range expectedMessages {
		if !strings.Contains(output, expected) {
			t.Errorf("expected output to contain '%s', got: %s", expected, output)
		}
	}
}
