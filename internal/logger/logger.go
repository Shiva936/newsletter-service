package logger

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
	FATAL
)

func (l Level) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

type Logger struct {
	level  Level
	logger *log.Logger
}

var defaultLogger *Logger

func init() {
	defaultLogger = &Logger{
		level:  INFO,
		logger: log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile),
	}
}

// New creates a new logger instance
func New(level Level) *Logger {
	return &Logger{
		level:  level,
		logger: log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile),
	}
}

// SetLevel sets the logging level
func SetLevel(level Level) {
	defaultLogger.level = level
}

func (l *Logger) log(level Level, ctx context.Context, format string, v ...interface{}) {
	if level < l.level {
		return
	}

	prefix := fmt.Sprintf("[%s] ", level.String())
	if ctx != nil {
		if requestID := ctx.Value("request_id"); requestID != nil {
			prefix += fmt.Sprintf("[%s] ", requestID)
		}
		if userID := ctx.Value("user_id"); userID != nil {
			prefix += fmt.Sprintf("[user:%s] ", userID)
		}
	}

	message := fmt.Sprintf(format, v...)
	l.logger.Printf("%s%s", prefix, message)

	if level == FATAL {
		os.Exit(1)
	}
}

// Context-aware logging functions
func (l *Logger) Debug(ctx context.Context, format string, v ...interface{}) {
	l.log(DEBUG, ctx, format, v...)
}

func (l *Logger) Info(ctx context.Context, format string, v ...interface{}) {
	l.log(INFO, ctx, format, v...)
}

func (l *Logger) Warn(ctx context.Context, format string, v ...interface{}) {
	l.log(WARN, ctx, format, v...)
}

func (l *Logger) Error(ctx context.Context, format string, v ...interface{}) {
	l.log(ERROR, ctx, format, v...)
}

func (l *Logger) Fatal(ctx context.Context, format string, v ...interface{}) {
	l.log(FATAL, ctx, format, v...)
}

// Global logging functions using default logger
func Debug(ctx context.Context, format string, v ...interface{}) {
	defaultLogger.Debug(ctx, format, v...)
}

func Info(ctx context.Context, format string, v ...interface{}) {
	defaultLogger.Info(ctx, format, v...)
}

func Warn(ctx context.Context, format string, v ...interface{}) {
	defaultLogger.Warn(ctx, format, v...)
}

func Error(ctx context.Context, format string, v ...interface{}) {
	defaultLogger.Error(ctx, format, v...)
}

func Fatal(ctx context.Context, format string, v ...interface{}) {
	defaultLogger.Fatal(ctx, format, v...)
}

// Non-context logging functions for backward compatibility
func Printf(format string, v ...interface{}) {
	defaultLogger.Info(nil, format, v...)
}

func Println(v ...interface{}) {
	defaultLogger.Info(nil, fmt.Sprintln(v...))
}

// LoggerMiddleware adds request ID and structured logging to gin context
func LoggerMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		start := time.Now()
		requestID := fmt.Sprintf("%d-%d", time.Now().Unix(), time.Now().Nanosecond())

		// Add request ID to context
		ctx := context.WithValue(c.Request.Context(), "request_id", requestID)
		c.Request = c.Request.WithContext(ctx)

		// Log request
		Info(ctx, "Request started: %s %s", c.Request.Method, c.Request.URL.Path)

		c.Next()

		// Log response
		duration := time.Since(start)
		status := c.Writer.Status()

		if status >= 400 {
			Error(ctx, "Request completed: %s %s - %d (%v)",
				c.Request.Method, c.Request.URL.Path, status, duration)
		} else {
			Info(ctx, "Request completed: %s %s - %d (%v)",
				c.Request.Method, c.Request.URL.Path, status, duration)
		}
	})
}
