package log

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

var (
	defaultLogger *slog.Logger
	logFile       *os.File
	appStartTime  = time.Now()
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
	colorWhite  = "\033[97m"

	colorBoldRed    = "\033[1;31m"
	colorBoldYellow = "\033[1;33m"
	colorBoldCyan   = "\033[1;36m"

	colorRedBg = "\033[41m"
)

// consoleHandler provides colorful console output
type consoleHandler struct {
	opts  slog.HandlerOptions
	mu    sync.Mutex
	out   io.Writer
	attrs []slog.Attr
	group string
}

func newConsoleHandler(w io.Writer, opts *slog.HandlerOptions) *consoleHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	return &consoleHandler{
		out:  w,
		opts: *opts,
	}
}

func (h *consoleHandler) Enabled(ctx context.Context, level slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}
	return level >= minLevel
}

func (h *consoleHandler) Handle(ctx context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	buf := &bytes.Buffer{}

	// Time in gray
	buf.WriteString(colorGray)
	buf.WriteString(r.Time.Format("15:04:05.000"))
	buf.WriteString(colorReset)
	buf.WriteString(" ")

	// Level with color
	levelColor := getLevelColor(r.Level)
	levelText := getLevelText(r.Level)
	buf.WriteString(levelColor)
	buf.WriteString(levelText)
	for i := len(levelText); i < 4; i++ {
		buf.WriteByte(' ')
	}
	buf.WriteString(colorReset)
	buf.WriteString(" ")

	// Message
	buf.WriteString(r.Message)

	// Attributes
	r.Attrs(func(a slog.Attr) bool {
		buf.WriteString(" ")
		appendAttr(buf, a, h.group)
		return true
	})

	// Prepended attributes
	for _, a := range h.attrs {
		buf.WriteString(" ")
		appendAttr(buf, a, h.group)
	}

	// Source in cyan at the end (if enabled)
	if h.opts.AddSource && r.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		f, _ := fs.Next()
		file := f.File
		// Shorten file path
		if idx := strings.LastIndex(file, "/"); idx >= 0 {
			if idx2 := strings.LastIndex(file[:idx], "/"); idx2 >= 0 {
				file = file[idx2+1:]
			}
		}
		buf.WriteString(" ")
		buf.WriteString(colorCyan)
		buf.WriteString(fmt.Sprintf("%s:%d", file, f.Line))
		buf.WriteString(colorReset)
	}

	buf.WriteByte('\n')

	_, err := h.out.Write(buf.Bytes())
	return err
}

func (h *consoleHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h2 := &consoleHandler{
		opts:  h.opts,
		out:   h.out,
		attrs: append(make([]slog.Attr, 0, len(h.attrs)+len(attrs)), append(h.attrs, attrs...)...),
		group: h.group,
	}
	return h2
}

func (h *consoleHandler) WithGroup(name string) slog.Handler {
	var newGroup string
	if h.group != "" {
		newGroup = h.group + "." + name
	} else {
		newGroup = name
	}
	h2 := &consoleHandler{
		opts:  h.opts,
		out:   h.out,
		attrs: append(make([]slog.Attr, 0, len(h.attrs)), h.attrs...),
		group: newGroup,
	}
	return h2
}

func getLevelColor(level slog.Level) string {
	switch {
	case level < LevelDebug:
		return colorGray // TRACE
	case level < LevelInfo:
		return colorBlue // DEBUG
	case level < LevelWarn:
		return colorGreen // INFO
	case level < LevelError:
		return colorYellow // WARN
	case level < LevelEmergency:
		return colorRed // ERROR
	default:
		return colorRedBg + colorWhite // EMERGENCY - white text on red background
	}
}

func getLevelText(level slog.Level) string {
	switch level {
	case LevelTrace:
		return "TRCE"
	case LevelDebug:
		return "DEBG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERRO"
	case LevelEmergency:
		return "EMER"
	default:
		return level.String()
	}
}

func appendAttr(buf *bytes.Buffer, a slog.Attr, group string) {
	if a.Equal(slog.Attr{}) {
		return
	}

	key := a.Key
	if group != "" {
		key = group + "." + key
	}

	buf.WriteString(colorPurple)
	buf.WriteString(key)
	buf.WriteString(colorReset)
	buf.WriteString("=")

	switch a.Value.Kind() {
	case slog.KindString:
		buf.WriteString(colorGreen)
		buf.WriteString(fmt.Sprintf("%q", a.Value.String()))
	case slog.KindInt64:
		buf.WriteString(colorCyan)
		buf.WriteString(fmt.Sprintf("%d", a.Value.Int64()))
	case slog.KindBool:
		buf.WriteString(colorYellow)
		buf.WriteString(fmt.Sprintf("%t", a.Value.Bool()))
	case slog.KindFloat64:
		buf.WriteString(colorCyan)
		buf.WriteString(fmt.Sprintf("%g", a.Value.Float64()))
	case slog.KindGroup:
		buf.WriteString("{")
		first := true
		for _, attr := range a.Value.Group() {
			if !first {
				buf.WriteString(" ")
			}
			appendAttr(buf, attr, key)
			first = false
		}
		buf.WriteString("}")
		return
	default:
		buf.WriteString(fmt.Sprintf("%v", a.Value.Any()))
	}
	buf.WriteString(colorReset)
}

// multiHandler writes to multiple handlers
type multiHandler struct {
	handlers []slog.Handler
}

func (h *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	var errs []error
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, r.Level) {
			if err := handler.Handle(ctx, r); err != nil {
				errs = append(errs, err)
			}
		}
	}
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

func (h *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithAttrs(attrs)
	}
	return &multiHandler{handlers: handlers}
}

func (h *multiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithGroup(name)
	}
	return &multiHandler{handlers: handlers}
}

func init() {
	// Create colorful console handler
	consoleHandler := newConsoleHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     LevelDebug,
	})

	handlers := []slog.Handler{consoleHandler}

	// Create file handler
	// cacheDir := os.TempDir() // Default to temp dir
	// if envCacheDir := os.Getenv("XDG_CACHE_HOME"); envCacheDir != "" {
	// 	cacheDir = envCacheDir
	// } else if homeDir, err := os.UserHomeDir(); err == nil {
	// 	cacheDir = filepath.Join(homeDir, ".cache")
	// }

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		cacheDir = os.TempDir() // Fallback to temp dir if UserCacheDir fails
	}

	// Get application name from args
	appName := filepath.Base(os.Args[0])
	if appName == "" {
		appName = "log"
	}

	// Create log directory
	logDir := filepath.Join(cacheDir, appName)
	if err := os.MkdirAll(logDir, 0755); err == nil {
		// Create log file with timestamp
		logFileName := fmt.Sprintf("%s_%s.log", appStartTime.Format("20060102-150405"), appName)
		logPath := filepath.Join(logDir, logFileName)

		if file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
			logFile = file
			// ReplaceAttr for JSON to handle custom level names
			replaceAttr := func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.LevelKey {
					level := a.Value.Any().(slog.Level)
					switch level {
					case LevelTrace:
						return slog.String(slog.LevelKey, "TRACE")
					case LevelEmergency:
						return slog.String(slog.LevelKey, "EMER")
					}
				}
				return a
			}
			fileHandler := slog.NewJSONHandler(logFile, &slog.HandlerOptions{
				AddSource:   true,
				Level:       LevelTrace,
				ReplaceAttr: replaceAttr,
			})
			handlers = append(handlers, fileHandler)
		}
	}

	// Create multi-handler
	handler := &multiHandler{handlers: handlers}
	defaultLogger = slog.New(handler)

	// Set log as the default slog logger
	slog.SetDefault(defaultLogger)
}

func SetDefault(l *slog.Logger) {
	defaultLogger = l
	slog.SetDefault(l)
}

func Default() *slog.Logger {
	return defaultLogger
}

func With(args ...any) *slog.Logger {
	return defaultLogger.With(args...)
}

func WithGroup(name string) *slog.Logger {
	return defaultLogger.WithGroup(name)
}

func Trace(msg string, args ...any) {
	log(context.Background(), LevelTrace, msg, args...)
}

func TraceContext(ctx context.Context, msg string, args ...any) {
	log(ctx, LevelTrace, msg, args...)
}

func Debug(msg string, args ...any) {
	log(context.Background(), slog.LevelDebug, msg, args...)
}

func DebugContext(ctx context.Context, msg string, args ...any) {
	log(ctx, slog.LevelDebug, msg, args...)
}

func Info(msg string, args ...any) {
	log(context.Background(), slog.LevelInfo, msg, args...)
}

func InfoContext(ctx context.Context, msg string, args ...any) {
	log(ctx, slog.LevelInfo, msg, args...)
}

func Warn(msg string, args ...any) {
	log(context.Background(), slog.LevelWarn, msg, args...)
}

func WarnContext(ctx context.Context, msg string, args ...any) {
	log(ctx, slog.LevelWarn, msg, args...)
}

func Error(msg string, args ...any) {
	log(context.Background(), slog.LevelError, msg, args...)
}

func ErrorContext(ctx context.Context, msg string, args ...any) {
	log(ctx, slog.LevelError, msg, args...)
}

func Emergency(msg string, args ...any) {
	log(context.Background(), LevelEmergency, msg, args...)
}

func EmergencyContext(ctx context.Context, msg string, args ...any) {
	log(ctx, LevelEmergency, msg, args...)
}

func Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	log(ctx, level, msg, args...)
}

func LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	defaultLogger.LogAttrs(ctx, level, msg, attrs...)
}

func log(ctx context.Context, level slog.Level, msg string, args ...any) {
	if !defaultLogger.Enabled(ctx, level) {
		return
	}

	var pc uintptr
	var pcs [1]uintptr
	runtime.Callers(3, pcs[:])
	pc = pcs[0]

	r := slog.NewRecord(time.Now(), level, msg, pc)
	r.Add(args...)
	if ctx == nil {
		ctx = context.Background()
	}
	_ = defaultLogger.Handler().Handle(ctx, r)
}

func NewLogger(h slog.Handler) *slog.Logger {
	return slog.New(h)
}

func NewTextHandler(w io.Writer, opts *slog.HandlerOptions) *slog.TextHandler {
	return slog.NewTextHandler(w, opts)
}

func NewJSONHandler(w io.Writer, opts *slog.HandlerOptions) *slog.JSONHandler {
	return slog.NewJSONHandler(w, opts)
}

// Printf-style logging functions

// Print logs a message at Info level using fmt.Sprint-style formatting
func Print(v ...any) {
	if !defaultLogger.Enabled(context.Background(), LevelInfo) {
		return
	}
	Info(fmt.Sprint(v...))
}

// Printf logs a message at Info level using fmt.Sprintf-style formatting
func Printf(format string, v ...any) {
	if !defaultLogger.Enabled(context.Background(), LevelInfo) {
		return
	}
	Info(fmt.Sprintf(format, v...))
}

// Println logs a message at Info level using fmt.Sprintln-style formatting
func Println(v ...any) {
	if !defaultLogger.Enabled(context.Background(), LevelInfo) {
		return
	}
	Info(fmt.Sprintln(v...))
}

// Level-specific printf functions

// Tracef logs a message at Trace level using fmt.Sprintf-style formatting
func Tracef(format string, v ...any) {
	if !defaultLogger.Enabled(context.Background(), LevelTrace) {
		return
	}
	Trace(fmt.Sprintf(format, v...))
}

// Debugf logs a message at Debug level using fmt.Sprintf-style formatting
func Debugf(format string, v ...any) {
	if !defaultLogger.Enabled(context.Background(), LevelDebug) {
		return
	}
	Debug(fmt.Sprintf(format, v...))
}

// Infof logs a message at Info level using fmt.Sprintf-style formatting
func Infof(format string, v ...any) {
	if !defaultLogger.Enabled(context.Background(), LevelInfo) {
		return
	}
	Info(fmt.Sprintf(format, v...))
}

// Warnf logs a message at Warn level using fmt.Sprintf-style formatting
func Warnf(format string, v ...any) {
	if !defaultLogger.Enabled(context.Background(), LevelWarn) {
		return
	}
	Warn(fmt.Sprintf(format, v...))
}

// Errorf logs a message at Error level using fmt.Sprintf-style formatting
func Errorf(format string, v ...any) {
	if !defaultLogger.Enabled(context.Background(), LevelError) {
		return
	}
	Error(fmt.Sprintf(format, v...))
}

// Emergencyf logs a message at Emergency level using fmt.Sprintf-style formatting
func Emergencyf(format string, v ...any) {
	if !defaultLogger.Enabled(context.Background(), LevelEmergency) {
		return
	}
	Emergency(fmt.Sprintf(format, v...))
}

type Handler = slog.Handler
type HandlerOptions = slog.HandlerOptions
type Level = slog.Level
type Leveler = slog.Leveler
type LogValuer = slog.LogValuer
type Record = slog.Record
type Source = slog.Source
type TextHandler = slog.TextHandler
type JSONHandler = slog.JSONHandler
type Attr = slog.Attr
type Value = slog.Value
type Kind = slog.Kind

const (
	LevelTrace     = slog.Level(-8)
	LevelDebug     = slog.LevelDebug
	LevelInfo      = slog.LevelInfo
	LevelWarn      = slog.LevelWarn
	LevelError     = slog.LevelError
	LevelEmergency = slog.Level(12)
)

var (
	Any           = slog.Any
	AnyValue      = slog.AnyValue
	Bool          = slog.Bool
	BoolValue     = slog.BoolValue
	Duration      = slog.Duration
	DurationValue = slog.DurationValue
	Float64       = slog.Float64
	Float64Value  = slog.Float64Value
	Group         = slog.Group
	GroupValue    = slog.GroupValue
	Int           = slog.Int
	Int64         = slog.Int64
	Int64Value    = slog.Int64Value
	IntValue      = slog.IntValue
	String        = slog.String
	StringValue   = slog.StringValue
	Time          = slog.Time
	TimeValue     = slog.TimeValue
	Uint64        = slog.Uint64
	Uint64Value   = slog.Uint64Value
)
