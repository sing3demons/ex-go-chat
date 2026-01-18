package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"realtime-chat-system/pkg/logAction"
	"strings"
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
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Type      string    `json:"type,omitempty"`
	Service   string    `json:"service"`
	TraceID   string    `json:"trace_id,omitempty"` //request ทั้งเส้น
	SpanID    string    `json:"span_id,omitempty"`  // step ย่อย (เช่น db call)
	ParentID  string    `json:"parent_id,omitempty"`
	Message   string    `json:"message,omitempty"`
	UseCase   string    `json:"usecase,omitempty"` //login,register

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
		summaryPath := config.Summary.Path
		// Check if path is a directory, append default filename with date
		if filepath.Ext(summaryPath) == "" || summaryPath[len(summaryPath)-1] == '/' || summaryPath[len(summaryPath)-1] == filepath.Separator {
			dateStr := time.Now().Format("2006-01-02")
			summaryPath = filepath.Join(summaryPath, fmt.Sprintf("summary-%s.log", dateStr))
		}
		if err := os.MkdirAll(filepath.Dir(summaryPath), 0755); err != nil {
			return nil, fmt.Errorf("failed to create summary log directory: %w", err)
		}
		fl.summaryFile, err = os.OpenFile(summaryPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open summary log file: %w", err)
		}
		fl.config.Summary.Path = summaryPath
	}

	// Initialize detail log file
	if config.Detail.File && config.Detail.Path != "" {
		detailPath := config.Detail.Path
		// Check if path is a directory, append default filename with date
		if filepath.Ext(detailPath) == "" || detailPath[len(detailPath)-1] == '/' || detailPath[len(detailPath)-1] == filepath.Separator {
			dateStr := time.Now().Format("2006-01-02")
			detailPath = filepath.Join(detailPath, fmt.Sprintf("detail-%s.log", dateStr))
		}
		if err := os.MkdirAll(filepath.Dir(detailPath), 0755); err != nil {
			return nil, fmt.Errorf("failed to create detail log directory: %w", err)
		}
		fl.detailFile, err = os.OpenFile(detailPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open detail log file: %w", err)
		}
		fl.config.Detail.Path = detailPath
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

	// Convert to JSON and back to get a clean map[string]any structure
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return data
	}

	var dataMap map[string]any
	if err := json.Unmarshal(jsonBytes, &dataMap); err != nil {
		return data
	}

	// Apply masking rules to the map
	l.applyMaskingToMap(dataMap, rules)

	return dataMap
}

// applyMaskingToMap applies masking rules to a map structure (case-insensitive)
func (l *CustomLogger) applyMaskingToMap(dataMap map[string]any, rules []MaskRule) {
	for _, rule := range rules {
		if rule.IsArray {
			// For array rules, use LookupMany to find all matching values
			data, ok := LookupMany(rule.Field, dataMap)
			if ok {
				// Update each found value by traversing and modifying the original structure
				for _, val := range data {
					maskedVal := l.maskValue(val, rule)
					l.updateMapValue(dataMap, rule.Field, val, maskedVal)
				}
			}
		} else {
			// For single value rules, use LookupOne (case-insensitive)
			data, ok := LookupOne(rule.Field, dataMap)
			if ok {
				maskedVal := l.maskValue(data, rule)
				l.updateMapValue(dataMap, rule.Field, data, maskedVal)
			}
		}
	}
}

// updateMapValue updates a value in the map at a given path (case-insensitive)
func (l *CustomLogger) updateMapValue(dataMap map[string]any, path string, oldVal any, newVal any) {
	parts := strings.Split(path, ".")
	if len(parts) == 0 {
		return
	}

	// Navigate to the parent map
	currentMap := dataMap
	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]
		// Try exact match first, then case-insensitive
		if nextMap, exists := currentMap[part].(map[string]any); exists {
			currentMap = nextMap
		} else if nextMap, exists := l.findMapCaseInsensitive(currentMap, part); exists {
			currentMap = nextMap
		} else if arr, exists := currentMap[part].([]any); exists {
			// Handle array elements
			for _, item := range arr {
				if itemMap, ok := item.(map[string]any); ok {
					l.updateMapValue(itemMap, strings.Join(parts[i+1:], "."), oldVal, newVal)
				}
			}
			return
		} else if arr, exists := l.findArrayCaseInsensitive(currentMap, part); exists {
			// Try case-insensitive array lookup
			for _, item := range arr {
				if itemMap, ok := item.(map[string]any); ok {
					l.updateMapValue(itemMap, strings.Join(parts[i+1:], "."), oldVal, newVal)
				}
			}
			return
		} else {
			return // Path doesn't exist
		}
	}

	// Update the final value - with case-insensitive matching
	lastPart := parts[len(parts)-1]
	actualKey := l.findKeyInMap(currentMap, lastPart)
	if actualKey == "" {
		return // Key not found
	}

	if val, exists := currentMap[actualKey]; exists {
		if arr, ok := val.([]any); ok {
			// If it's an array, update matching elements
			for i, item := range arr {
				if l.valuesEqual(item, oldVal) {
					arr[i] = newVal
				}
			}
			currentMap[actualKey] = arr
		} else if l.valuesEqual(val, oldVal) {
			// If it's a single value, update if it matches
			currentMap[actualKey] = newVal
		}
	}
}

// findKeyInMap finds a key in a map (case-insensitive)
func (l *CustomLogger) findKeyInMap(m map[string]any, key string) string {
	// Try exact match first
	if _, exists := m[key]; exists {
		return key
	}

	// Try case-insensitive match
	lowerKey := strings.ToLower(key)
	for k := range m {
		if strings.ToLower(k) == lowerKey {
			return k
		}
	}

	return ""
}

// findMapCaseInsensitive finds a nested map with case-insensitive key
func (l *CustomLogger) findMapCaseInsensitive(m map[string]any, key string) (map[string]any, bool) {
	lowerKey := strings.ToLower(key)
	for k, v := range m {
		if strings.ToLower(k) == lowerKey {
			if nextMap, ok := v.(map[string]any); ok {
				return nextMap, true
			}
			break
		}
	}
	return nil, false
}

// findArrayCaseInsensitive finds an array with case-insensitive key
func (l *CustomLogger) findArrayCaseInsensitive(m map[string]any, key string) ([]any, bool) {
	lowerKey := strings.ToLower(key)
	for k, v := range m {
		if strings.ToLower(k) == lowerKey {
			if arr, ok := v.([]any); ok {
				return arr, true
			}
			break
		}
	}
	return nil, false
}

// valuesEqual compares two values for equality
func (l *CustomLogger) valuesEqual(a, b any) bool {
	switch av := a.(type) {
	case string:
		if bv, ok := b.(string); ok {
			return av == bv
		}
	case float64:
		if bv, ok := b.(float64); ok {
			return av == bv
		}
	case bool:
		if bv, ok := b.(bool); ok {
			return av == bv
		}
	}
	return false
}

// maskValue applies masking to a single value (string, number, etc)
func (l *CustomLogger) maskValue(val any, rule MaskRule) any {
	switch v := val.(type) {
	case string:
		return l.masker.Mask(v, rule)
	case float64:
		// Try to mask as string if it looks like a phone number or similar
		return l.masker.Mask(fmt.Sprintf("%v", v), rule)

	default:
		return val
	}
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
		Timestamp:         time.Now(),
		Level:             level,
		Type:              detail,
		UseCase:          l.logEntry.UseCase,
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
		Timestamp:    time.Now(),
		Level:        "INFO",
		UseCase:      l.logEntry.UseCase,
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
		Timestamp:    time.Now(),
		Level:        "ERROR",
		Type:         summary,
		UseCase:      l.logEntry.UseCase,
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
