package mlog

import (
	"context"
	"realtime-chat-system/pkg/logger"
)

// L returns logger from context or the singleton instance
// Usage: log := mlog.L(ctx)
func L(ctx context.Context) logger.ICustomLogger {
	if ctx == nil {
		return logger.NewCustomLogger("", logger.LoggerConfig{})
	}

	// Try to get logger from context
	l, ok := ctx.Value(logger.LoggerKey).(logger.ICustomLogger)
	if ok && l != nil {
		return l
	}

	// Fallback to singleton instance
	return logger.NewCustomLogger("", logger.LoggerConfig{})
}
