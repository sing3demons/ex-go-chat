package logger

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"os"
)

// {
//   "timestamp",
//   "level",
//   "service",
//   "trace_id",
//   "span_id",
//   "message"
// }

type contextKey string

const (
	TraceIDKey      contextKey = "trace_id"
	SpanIDKey       contextKey = "span_id"
	ParentSpanIDKey contextKey = "parent_id"
	OperationKey    contextKey = "operation"
)

type LoggerAction struct {
	Action            string `json:"action,omitempty"`
	ActionDescription string `json:"actionDescription,omitempty"`
	SubAction         string `json:"subAction,omitempty"`
}

type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Service   string `json:"service"`
	TraceID   string `json:"trace_id,omitempty"` //request ทั้งเส้น
	SpanID    string `json:"span_id,omitempty"`  // step ย่อย (เช่น db call)
	ParentID  string `json:"parent_id,omitempty"`
	Message   string `json:"message"`
	Operation string `json:"operation,omitempty"` //login,register

	Action            string `json:"action,omitempty"`
	ActionDescription string `json:"action_description,omitempty"`
	SubAction         string `json:"sub_action,omitempty"`

	Dependency string `json:"dependency,omitempty"`
	Duration   string `json:"duration,omitempty"`

	StartTime string `json:"-"`
}

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
	info  *log.Logger
	error *log.Logger
	debug *log.Logger
}

// New creates a new logger instance
func New() *Logger {
	return &Logger{
		info:  log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		error: log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
		debug: log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

// Info logs an info message
func (l *Logger) Info(v ...interface{}) {
	l.info.Println(v...)
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, v ...interface{}) {
	l.info.Printf(format, v...)
}

// Error logs an error message
func (l *Logger) Error(v ...interface{}) {
	l.error.Println(v...)
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.error.Printf(format, v...)
}

// Debug logs a debug message
func (l *Logger) Debug(v ...interface{}) {
	l.debug.Println(v...)
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, v ...interface{}) {
	l.debug.Printf(format, v...)
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	l.info.Printf("WARN: "+format, v...)
}
