package logger

// NoOpLogger is a logger that does nothing, useful for testing
type NoOpLogger struct{}

func (n *NoOpLogger) Debug(message string, args ...interface{})      {}
func (n *NoOpLogger) Info(message string, args ...interface{})       {}
func (n *NoOpLogger) Warn(message string, args ...interface{})       {}
func (n *NoOpLogger) Error(message interface{}, args ...interface{}) {}
func (n *NoOpLogger) Errorf(format string, args ...interface{})      {}
func (n *NoOpLogger) Fatal(message interface{}, args ...interface{}) {}
func (n *NoOpLogger) WithFields(fields map[string]interface{}) Interface {
	return n
}
