# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go logging library (`github.com/qwer-go/log`) that extends Go's built-in `log/slog` package with enhanced features:

- **Colorful console output** with custom formatting
- **Dual output**: Console with colors + JSON file logging to `~/.cache/{appname}/` 
- **Extended log levels**: TRACE (-8) and EMERGENCY (12) in addition to standard slog levels
- **Drop-in replacement** for `log/slog` with same API
- **Custom handlers**: `consoleHandler` for colored output, `multiHandler` for multiple destinations

## Development Commands

### Testing
```bash
# Run all tests
go test .

# Run tests with verbose output
go test -v .

# Run specific test
go test -run TestBasicLogging

# Run tests with coverage
go test -cover .
```

### Building
```bash
# Build the module (verify compilation)
go build .

# Run examples
go run example/demo.go
go run example/levels.go  
go run example/emergency.go
```

### Module Management
```bash
# Tidy dependencies
go mod tidy

# Verify module
go mod verify
```

## Architecture

### Core Components

- **`log.go`**: Main implementation with custom handlers and logging functions
- **`consoleHandler`**: Provides ANSI colored console output with time, level, message, attributes, and source location
- **`multiHandler`**: Multiplexes logging to multiple handlers (console + file)
- **Custom levels**: `LevelTrace` (-8) and `LevelEmergency` (12)

### Key Features

1. **Automatic file logging**: Creates timestamped log files in `~/.cache/{appname}/`
2. **Source information**: Shows file:line when `AddSource: true`
3. **Color coding**: Different colors for log levels and attribute types
4. **slog compatibility**: All `slog` types and functions are re-exported

### File Structure

- `log.go`: Core logging implementation
- `log_test.go`: Unit tests covering all functionality
- `example_test.go`: Example usage patterns
- `example/`: Demonstration programs showing different use cases

## Testing Notes

- Tests use in-memory buffers to capture and verify log output
- JSON handler tests verify structured logging format
- Compatibility tests ensure drop-in replacement for `slog`
- One test currently fails due to filename expectation in `TestAddSource`

## Usage Patterns

The library is designed as a drop-in replacement for `log/slog`:
```go
import "github.com/qwer-go/log"

log.Info("message", "key", "value")  // Instead of slog.Info(...)
```