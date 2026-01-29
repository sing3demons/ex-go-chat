package logger

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"os"
)

type contextKey string

const (
	TraceIDKey      contextKey = "trace_id"
	SpanIDKey       contextKey = "span_id"
	ParentSpanIDKey contextKey = "parent_id"
	OperationKey    contextKey = "operation"
	summary                    = "summary"
	detail                     = "detail"
	LoggerKey       contextKey = "logger"
)

func genID(bytes int) string {
	b := make([]byte, bytes)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func NewTraceID() string {
	return genID(16) // 128-bit
}

func NewSpanID() string {
	return genID(8) // 64-bit
}

func StartSpanFromContext(ctx context.Context, name string) context.Context {
	parent := ctx.Value(SpanIDKey)
	spanID := NewSpanID()

	ctx = context.WithValue(ctx, ParentSpanIDKey, parent)
	ctx = context.WithValue(ctx, SpanIDKey, name+spanID)
	return ctx
}

func StartAPI(r *http.Request, operation string) context.Context {
	traceID := r.Header.Get("X-Trace-Id")
	parentSpan := r.Header.Get("X-Parent-Span-Id")

	if traceID == "" {
		traceID = NewTraceID()
	}

	spanID := NewSpanID()

	ctx := r.Context()
	ctx = context.WithValue(ctx, TraceIDKey, traceID)
	ctx = context.WithValue(ctx, SpanIDKey, spanID)
	ctx = context.WithValue(ctx, ParentSpanIDKey, parentSpan)
	ctx = context.WithValue(ctx, OperationKey, operation)

	return ctx
}

// Logger provides logging functionality
type Logger struct {
	logger *slog.Logger
}

// New creates a new logger instance
func New() *Logger {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	return &Logger{logger: logger}
}

// Info logs an info message
func (l *Logger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, v ...any) {
	l.logger.Info(fmt.Sprintf(format, v...))
}

// Error logs an error message
func (l *Logger) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, v ...any) {
	l.logger.Error(fmt.Sprintf(format, v...))
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, v ...any) {
	l.logger.Debug(fmt.Sprintf(format, v...))
}

func (l *Logger) Warnf(format string, v ...any) {
	l.logger.Warn(fmt.Sprintf(format, v...))
}

func (l *Logger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}