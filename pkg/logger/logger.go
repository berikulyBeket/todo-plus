package logger

// Interface defines the logger interface for logging at different levels
type Interface interface {
	Debug(message string, args ...interface{})
	Info(message string, args ...interface{})
	Warn(message string, args ...interface{})
	Error(message interface{}, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(message interface{}, args ...interface{})
	WithFields(fields map[string]interface{}) Interface
}
