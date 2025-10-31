package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Logger wraps zerolog logger
type Logger struct {
	logger zerolog.Logger
}

// Config contains logger configuration
type Config struct {
	Level      string
	Format     string // json or console
	OutputPath string
}

// New creates a new logger instance
func New(cfg Config) *Logger {
	var output io.Writer = os.Stdout

	// Set log level
	level := parseLevel(cfg.Level)
	zerolog.SetGlobalLevel(level)

	// Configure output format
	if cfg.Format == "console" {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	// Create logger
	logger := zerolog.New(output).
		With().
		Timestamp().
		Caller().
		Logger()

	return &Logger{logger: logger}
}

// parseLevel converts string to zerolog level
func parseLevel(level string) zerolog.Level {
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	default:
		return zerolog.InfoLevel
	}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string) {
	l.logger.Debug().Msg(msg)
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, v ...interface{}) {
	l.logger.Debug().Msgf(format, v...)
}

// Info logs an info message
func (l *Logger) Info(msg string) {
	l.logger.Info().Msg(msg)
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, v ...interface{}) {
	l.logger.Info().Msgf(format, v...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string) {
	l.logger.Warn().Msg(msg)
}

// Warnf logs a formatted warning message
func (l *Logger) Warnf(format string, v ...interface{}) {
	l.logger.Warn().Msgf(format, v...)
}

// Error logs an error message
func (l *Logger) Error(msg string) {
	l.logger.Error().Msg(msg)
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.logger.Error().Msgf(format, v...)
}

// ErrorWithErr logs an error with error object
func (l *Logger) ErrorWithErr(err error, msg string) {
	l.logger.Error().Err(err).Msg(msg)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string) {
	l.logger.Fatal().Msg(msg)
}

// Fatalf logs a formatted fatal message and exits
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.logger.Fatal().Msgf(format, v...)
}

// With returns a logger with additional fields
func (l *Logger) With(key string, value interface{}) *Logger {
	newLogger := l.logger.With().Interface(key, value).Logger()
	return &Logger{logger: newLogger}
}

// WithFields returns a logger with multiple fields
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	ctx := l.logger.With()
	for k, v := range fields {
		ctx = ctx.Interface(k, v)
	}
	return &Logger{logger: ctx.Logger()}
}

// WithError returns a logger with error field
func (l *Logger) WithError(err error) *Logger {
	newLogger := l.logger.With().Err(err).Logger()
	return &Logger{logger: newLogger}
}

// GetZerolog returns the underlying zerolog logger
func (l *Logger) GetZerolog() zerolog.Logger {
	return l.logger
}

// Global logger functions for convenience

var globalLogger *Logger

// Init initializes the global logger
func Init(cfg Config) {
	globalLogger = New(cfg)
	log.Logger = globalLogger.logger
}

// Debug logs a debug message using global logger
func Debug(msg string) {
	if globalLogger != nil {
		globalLogger.Debug(msg)
	}
}

// Info logs an info message using global logger
func Info(msg string) {
	if globalLogger != nil {
		globalLogger.Info(msg)
	}
}

// Warn logs a warning message using global logger
func Warn(msg string) {
	if globalLogger != nil {
		globalLogger.Warn(msg)
	}
}

// Error logs an error message using global logger
func Error(msg string) {
	if globalLogger != nil {
		globalLogger.Error(msg)
	}
}

// ErrorWithErr logs an error with error object using global logger
func ErrorWithErr(err error, msg string) {
	if globalLogger != nil {
		globalLogger.ErrorWithErr(err, msg)
	}
}

// Fatal logs a fatal message and exits using global logger
func Fatal(msg string) {
	if globalLogger != nil {
		globalLogger.Fatal(msg)
	}
}
