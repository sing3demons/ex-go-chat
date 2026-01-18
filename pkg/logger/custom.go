package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"realtime-chat-system/pkg/logAction"
	"sync"
	"time"
)

// {
//   "timestamp",
//   "level",
//   "service",
//   "trace_id",
//   "span_id",
//   "message"
// }

type LoggerAction struct {
	Action            string `json:"action,omitempty"`
	ActionDescription string `json:"actionDescription,omitempty"`
	SubAction         string `json:"subAction,omitempty"`
}

type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Type      string `json:"type,omitempty"`
	Service   string `json:"service"`
	TraceID   string `json:"trace_id,omitempty"` //request ทั้งเส้น
	SpanID    string `json:"span_id,omitempty"`  // step ย่อย (เช่น db call)
	ParentID  string `json:"parent_id,omitempty"`
	Message   string `json:"message"`
	UseCase   string `json:"usecase,omitempty"` //login,register

	Action            string `json:"action,omitempty"`
	ActionDescription string `json:"action_description,omitempty"`
	SubAction         string `json:"sub_action,omitempty"`

	Dependency string `json:"dependency,omitempty"`

	StartTime time.Time `json:"-"`

	ResponseTime   int64          `json:"responseTime,omitempty"`
	ResultCode     string         `json:"resultCode,omitempty"`
	ResultFlag     string         `json:"resultFlag,omitempty"`
	AdditionalInfo map[string]any `json:"additionalInfo,omitempty"`
}

type ICustomLogger interface {
	Init(useCase string, span_id string)
	Info(action logAction.LoggerAction, data any, maskingData ...MaskRule)
	Debug(action logAction.LoggerAction, data any, maskingData ...MaskRule)
	Error(action logAction.LoggerAction, data any, maskingData ...MaskRule)
	Flush(code int, msg string)
	FlushError(code int, msg string)
	SetDependencyMetadata(metadata DependencyMetadata) ICustomLogger
	AddMetadata(key string, value any) ICustomLogger

	TraceID() string
	SetTraceID(trace_id string)
	SetSpanID(span_id string)
	SpanID() string
	SetUseCase(useCase string)
}

// FileLogger handles writing logs to files with rotation support
type FileLogger struct {
	mu              sync.Mutex
	summaryFile     io.WriteCloser
	detailFile      io.WriteCloser
	config          LoggerConfig
	fileSizeTracker map[string]int64
}

func NewFileLogger(config LoggerConfig) (*FileLogger, error) {
	fl := &FileLogger{
		config:          config,
		fileSizeTracker: make(map[string]int64),
	}

	var err error

	// Initialize summary log file
	if config.Summary.File && config.Summary.Path != "" {
		if err := os.MkdirAll(filepath.Dir(config.Summary.Path), 0755); err != nil {
			return nil, fmt.Errorf("failed to create summary log directory: %w", err)
		}
		fl.summaryFile, err = os.OpenFile(config.Summary.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open summary log file: %w", err)
		}
	}

	// Initialize detail log file
	if config.Detail.File && config.Detail.Path != "" {
		if err := os.MkdirAll(filepath.Dir(config.Detail.Path), 0755); err != nil {
			return nil, fmt.Errorf("failed to create detail log directory: %w", err)
		}
		fl.detailFile, err = os.OpenFile(config.Detail.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open detail log file: %w", err)
		}
	}

	return fl, nil
}

func (fl *FileLogger) Log(logType string, msg any) {
	fl.mu.Lock()
	defer fl.mu.Unlock()

	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		jsonBytes = []byte(fmt.Sprintf("%+v", msg))
	}
	jsonBytes = append(jsonBytes, '\n')

	switch logType {
	case "summary":
		if fl.config.Summary.Console {
			// log.Printf("%s\n", string(jsonBytes))
			os.Stdout.Write(jsonBytes)
		}
		if fl.summaryFile != nil {
			if _, err := fl.summaryFile.Write(jsonBytes); err != nil {
				log.Printf("failed to write summary log: %v", err)
			}
			fl.checkRotation(fl.config.Summary.Path)
		}

	case "detail":
		if fl.config.Detail.Console {
			// log.Printf("%s\n", string(jsonBytes))
			os.Stdout.Write(jsonBytes)
		}
		if fl.detailFile != nil {
			if _, err := fl.detailFile.Write(jsonBytes); err != nil {
				log.Printf("failed to write detail log: %v", err)
			}
			fl.checkRotation(fl.config.Detail.Path)
		}
	}
}

func (fl *FileLogger) checkRotation(filePath string) {
	if filePath == "" || fl.config.Rotation.MaxSize <= 0 {
		return
	}

	info, err := os.Stat(filePath)
	if err != nil {
		return
	}

	if info.Size() > fl.config.Rotation.MaxSize {
		fl.rotate(filePath)
	}
}

func (fl *FileLogger) rotate(filePath string) {
	timestamp := time.Now().Format("20060102_150405")
	backupPath := fmt.Sprintf("%s.%s", filePath, timestamp)

	if err := os.Rename(filePath, backupPath); err != nil {
		log.Printf("failed to rotate log file: %v", err)
		return
	}

	// Reopen the file
	if file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
		if filePath == fl.config.Summary.Path && fl.summaryFile != nil {
			fl.summaryFile.Close()
			fl.summaryFile = file
		} else if filePath == fl.config.Detail.Path && fl.detailFile != nil {
			fl.detailFile.Close()
			fl.detailFile = file
		}
	}
}

func (fl *FileLogger) Close() error {
	fl.mu.Lock()
	defer fl.mu.Unlock()

	if fl.summaryFile != nil {
		fl.summaryFile.Close()
	}
	if fl.detailFile != nil {
		fl.detailFile.Close()
	}
	return nil
}

// LoggerDefault maintains backward compatibility and logs to console
type LoggerDefault struct {
}

func (l *LoggerDefault) Log(msg any) {
	log.Printf("%+v\n", msg)
}

type CustomLogger struct {
	mu         sync.RWMutex // Protects concurrent access
	fileLogger *FileLogger
	logEntry   LogEntry
	masker     Masker
}

// Singleton instance and initialization
var (
	customLoggerInstance *CustomLogger
	once                 sync.Once
)

// RotationConfig defines log rotation settings
type RotationConfig struct {
	MaxSize    int64 // Maximum size in bytes before rotation (default: 100MB)
	MaxAge     int   // Maximum number of days to retain old logs (default: 30)
	MaxBackups int   // Maximum number of backup files to keep (default: 10)
	Compress   bool  // Whether to compress rotated files (default: true)
}
type LogOutputConfig struct {
	Path    string
	Console bool
	File    bool
}
type LoggerConfig struct {
	Summary  LogOutputConfig
	Detail   LogOutputConfig
	Rotation RotationConfig
}

// NewCustomLogger creates or returns the singleton CustomLogger instance
func NewCustomLogger(serviceName string, config LoggerConfig) ICustomLogger {
	once.Do(func() {
		fileLogger, err := NewFileLogger(config)
		if err != nil {
			log.Printf("failed to initialize file logger: %v", err)
		}
		customLoggerInstance = &CustomLogger{
			fileLogger: fileLogger,
			logEntry:   LogEntry{Service: serviceName},
			masker:     NewMasker("X"),
		}
	})

	return customLoggerInstance
}

// SetLoggerConfig initializes the file logger with configuration
func SetLoggerConfig(config LoggerConfig) error {
	if customLoggerInstance == nil {
		return fmt.Errorf("logger not initialized, call NewCustomLogger first")
	}

	customLoggerInstance.mu.Lock()
	defer customLoggerInstance.mu.Unlock()

	fileLogger, err := NewFileLogger(config)
	if err != nil {
		return err
	}

	if customLoggerInstance.fileLogger != nil {
		customLoggerInstance.fileLogger.Close()
	}
	customLoggerInstance.fileLogger = fileLogger
	return nil
}

// CloseLogger closes all open file handles
func CloseLogger() error {
	if customLoggerInstance != nil && customLoggerInstance.fileLogger != nil {
		customLoggerInstance.mu.Lock()
		defer customLoggerInstance.mu.Unlock()
		return customLoggerInstance.fileLogger.Close()
	}
	return nil
}

func (l *CustomLogger) AddMetadata(key string, value any) ICustomLogger {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.logEntry.AdditionalInfo == nil {
		l.logEntry.AdditionalInfo = make(map[string]any)
	}
	l.logEntry.AdditionalInfo[key] = value

	return l
}

type DependencyMetadata struct {
	Dependency   string `json:"dependency,omitempty"`
	ResponseTime int64  `json:"responseTime,omitempty"`
	ResultCode   string `json:"resultCode,omitempty"`
	ResultFlag   string `json:"resultFlag,omitempty"`
}

func (l *CustomLogger) SetDependencyMetadata(metadata DependencyMetadata) ICustomLogger {
	l.mu.Lock()
	defer l.mu.Unlock()

	if metadata.Dependency != "" {
		l.logEntry.Dependency = metadata.Dependency
	}
	if metadata.ResponseTime != 0 {
		l.logEntry.ResponseTime = metadata.ResponseTime
	}
	if metadata.ResultCode != "" {
		l.logEntry.ResultCode = metadata.ResultCode
	}
	if metadata.ResultFlag != "" {
		l.logEntry.ResultFlag = metadata.ResultFlag
	}
	return l
}

func (l *CustomLogger) maskData(data any, rules []MaskRule) any {
	if len(rules) == 0 {
		return data
	}

	// Convert to map for easier manipulation
	var dataMap map[string]any

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return data
	}

	if err := json.Unmarshal(jsonBytes, &dataMap); err != nil {
		return data
	}

	for _, rule := range rules {
		if val, exists := dataMap[rule.Field]; exists {
			if rule.IsArray {
				// apply masking to each element in the array
				if arr, ok := val.([]any); ok {
					maskedArr := make([]any, len(arr))
					for i, item := range arr {
						if str, ok := item.(string); ok {
							maskedArr[i] = l.masker.Mask(str, rule)
						} else {
							maskedArr[i] = item
						}
					}
					dataMap[rule.Field] = maskedArr
				}
			} else {
				if str, ok := val.(string); ok {
					dataMap[rule.Field] = l.masker.Mask(str, rule)
				}
			}
		}
	}

	return dataMap
}

func (l *CustomLogger) log(level string, action logAction.LoggerAction, data any, maskingData ...MaskRule) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var maskedData any
	if len(maskingData) > 0 {
		maskedData = l.maskData(data, maskingData)
	} else {
		maskedData = data
	}
	logMsg := LogEntry{
		Timestamp:         time.Now().String(),
		Level:             level,
		Type:              detail,
		Service:           l.logEntry.Service,
		TraceID:           l.logEntry.TraceID,
		SpanID:            l.logEntry.SpanID,
		Message:           dataToString(maskedData),
		Action:            action.Action,
		ActionDescription: action.ActionDescription,
		SubAction:         action.SubAction,
	}
	if l.logEntry.ResponseTime != 0 {
		logMsg.ResponseTime = l.logEntry.ResponseTime
		l.logEntry.ResponseTime = 0
	}
	if l.logEntry.ResultCode != "" {
		logMsg.ResultCode = l.logEntry.ResultCode
		l.logEntry.ResultCode = ""
	}
	if l.logEntry.ResultFlag != "" {
		logMsg.ResultFlag = l.logEntry.ResultFlag
		l.logEntry.ResultFlag = ""
	}
	if l.logEntry.Dependency != "" {
		logMsg.Dependency = l.logEntry.Dependency
		l.logEntry.Dependency = ""
	}
	if l.fileLogger != nil {
		l.fileLogger.Log(detail, logMsg)
	} else {
		os.Stdout.Write([]byte(fmt.Sprintf("%+v\n", logMsg)))
	}
}

func (l *CustomLogger) Debug(action logAction.LoggerAction, data any, maskingData ...MaskRule) {
	l.log("DEBUG", action, data, maskingData...)
}
func (l *CustomLogger) Error(action logAction.LoggerAction, data any, maskingData ...MaskRule) {
	l.log("ERROR", action, data, maskingData...)
}
func (l *CustomLogger) Info(action logAction.LoggerAction, data any, maskingData ...MaskRule) {
	l.log("INFO", action, data, maskingData...)
}
func (l *CustomLogger) Init(useCase, span_id string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	l.logEntry.StartTime = now
	l.logEntry.UseCase = useCase
	if span_id == "" {
		span_id = NewSpanID()
	}
	l.logEntry.SpanID = span_id
	l.logEntry.TraceID = NewTraceID()
}

func (l *CustomLogger) SetUseCase(useCase string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logEntry.UseCase = useCase
}

func (l *CustomLogger) SetTraceID(trace_id string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logEntry.TraceID = trace_id
}

func (l *CustomLogger) TraceID() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.logEntry.TraceID
}

func (l *CustomLogger) SetSpanID(span_id string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logEntry.SpanID = span_id
}
func (l *CustomLogger) SpanID() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.logEntry.SpanID
}
func (l *CustomLogger) Flush(code int, msg string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.logEntry.StartTime.IsZero() {
		panic("Flush called before Init")
	}
	logMsg := LogEntry{
		Timestamp:    time.Now().String(),
		Level:        "INFO",
		Type:         summary,
		Service:      l.logEntry.Service,
		TraceID:      l.logEntry.TraceID,
		SpanID:       l.logEntry.SpanID,
		ResultCode:   fmt.Sprintf("%d", code),
		ResponseTime: time.Since(l.logEntry.StartTime).Milliseconds(),
	}

	if l.logEntry.AdditionalInfo != nil {
		logMsg.AdditionalInfo = l.logEntry.AdditionalInfo

		l.logEntry.AdditionalInfo = nil
	}
	if l.fileLogger != nil {
		l.fileLogger.Log(summary, logMsg)
	}
	// reset log entry
	l.logEntry = LogEntry{Service: l.logEntry.Service}
}
func (l *CustomLogger) FlushError(code int, msg string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.logEntry.StartTime.IsZero() {
		panic("FlushError called before Init")
	}
	logMsg := LogEntry{
		Timestamp:    time.Now().String(),
		Level:        "ERROR",
		Type:         summary,
		Service:      l.logEntry.Service,
		TraceID:      l.logEntry.TraceID,
		SpanID:       l.logEntry.SpanID,
		ResultCode:   fmt.Sprintf("%d", code),
		ResponseTime: time.Since(l.logEntry.StartTime).Milliseconds(),
		Message:      msg,
	}
	if l.logEntry.AdditionalInfo != nil {
		logMsg.AdditionalInfo = l.logEntry.AdditionalInfo

		l.logEntry.AdditionalInfo = nil
	}
	if l.fileLogger != nil {
		l.fileLogger.Log(summary, logMsg)
	}
	// reset log entry
	l.logEntry = LogEntry{Service: l.logEntry.Service}
}

func ApplyMasking(data any, masker Masker, rules []MaskRule) any {
	switch v := data.(type) {
	case map[string]any:
		for key, val := range v {
			// check if there is a masking rule for this field
			for _, rule := range rules {
				if rule.Field == key {
					if rule.IsArray {
						// apply masking to each element in the array
						if arr, ok := val.([]any); ok {
							maskedArr := make([]any, len(arr))
							for i, item := range arr {
								if str, ok := item.(string); ok {
									maskedArr[i] = masker.Mask(str, rule)
								} else {
									maskedArr[i] = item
								}
							}
							v[key] = maskedArr
						}
					} else {
						if str, ok := val.(string); ok {
							v[key] = masker.Mask(str, rule)
						}
					}
				}
			}
		}
		return v
	default:
		return data
	}
}

func dataToString(data any) string {
	if data == nil {
		return ""
	}

	if str, ok := data.(string); ok {
		return str
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return ""
	}

	return string(jsonBytes)
}
