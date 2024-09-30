package logger

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"github.com/sirupsen/logrus"
)

type LogrusLogger struct {
	entry *logrus.Entry
}

// New creates a new logger with a given log level
func New(level string) Interface {
	var l logrus.Level

	switch strings.ToLower(level) {
	case "error":
		l = logrus.ErrorLevel
	case "warn":
		l = logrus.WarnLevel
	case "info":
		l = logrus.InfoLevel
	case "debug":
		l = logrus.DebugLevel
	default:
		l = logrus.InfoLevel
	}

	log := logrus.New()
	log.SetLevel(l)
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetOutput(os.Stdout)

	return &LogrusLogger{
		entry: logrus.NewEntry(log),
	}
}

// Debug logs a message at debug level
func (l *LogrusLogger) Debug(message string, args ...interface{}) {
	l.log("debug", message, args...)
}

// Info logs a message at info level
func (l *LogrusLogger) Info(message string, args ...interface{}) {
	l.log("info", message, args...)
}

// Warn logs a message at warn level
func (l *LogrusLogger) Warn(message string, args ...interface{}) {
	l.log("warn", message, args...)
}

// Error logs a message at error level, and logs debug details if debug is enabled
func (l *LogrusLogger) Error(message interface{}, args ...interface{}) {
	fields := map[string]interface{}{
		"stack_trace": string(debug.Stack()),
	}

	if l.entry.Logger.GetLevel() == logrus.DebugLevel {
		l.WithFields(fields).Debug(fmt.Sprintf("Error occurred: %v", message), args...)
		return
	}

	l.WithFields(fields).Error(message, args...)
}

// Errorf logs a formatted error message at error level (new method)
func (l *LogrusLogger) Errorf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	fields := map[string]interface{}{
		"stack_trace": string(debug.Stack()),
	}

	if l.entry.Logger.GetLevel() == logrus.DebugLevel {
		l.WithFields(fields).Debug(fmt.Sprintf("Error occurred: %s", message))
		return
	}

	l.WithFields(fields).Error(message)
}

// Fatal logs a message at fatal level and exits the application
func (l *LogrusLogger) Fatal(message interface{}, args ...interface{}) {
	fields := map[string]interface{}{
		"stack_trace": string(debug.Stack()),
	}

	l.WithFields(fields).Fatal(message, args...)
	os.Exit(1)
}

// WithFields allows structured logging by adding custom fields
func (l *LogrusLogger) WithFields(fields map[string]interface{}) Interface {
	return &LogrusLogger{
		entry: l.entry.WithFields(logrus.Fields(fields)),
	}
}

// log handles the generic logging logic
func (l *LogrusLogger) log(level, message string, args ...interface{}) {
	if len(args) > 0 {
		message = fmt.Sprintf(message, args...)
	}

	switch level {
	case "debug":
		l.entry.Debug(message)
	case "info":
		l.entry.Info(message)
	case "warn":
		l.entry.Warn(message)
	case "error":
		l.entry.Error(message)
	case "fatal":
		l.entry.Fatal(message)
	}
}
