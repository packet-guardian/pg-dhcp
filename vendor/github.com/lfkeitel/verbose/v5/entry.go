package verbose

import (
	"os"
	"time"
)

// Entry represents a log message going through the system
type Entry struct {
	Level     LogLevel
	Timestamp time.Time
	Logger    *Logger
	Message   string
	Data      Fields
}

// NewEntry creates a new, empty Entry
func NewEntry(l *Logger) *Entry {
	return &Entry{
		Logger: l,
		Data:   nil,
	}
}

// WithField adds a single field to the Entry.
func (e *Entry) WithField(key string, value interface{}) *Entry {
	return e.WithFields(Fields{key: value})
}

// WithFields adds a map of fields to the Entry.
func (e *Entry) WithFields(fields Fields) *Entry {
	data := make(Fields, len(e.Data)+len(fields))
	for k, v := range e.Data {
		data[k] = v
	}
	for k, v := range fields {
		data[k] = v
	}
	return &Entry{Logger: e.Logger, Data: data}
}

// Debug - Log Debug message
func (e *Entry) Debug(msg string) {
	e.Logger.log(e, LogLevelDebug, msg)
}

// Info - Log Info message
func (e *Entry) Info(msg string) {
	e.Logger.log(e, LogLevelInfo, msg)
}

// Notice - Log Notice message
func (e *Entry) Notice(msg string) {
	e.Logger.log(e, LogLevelNotice, msg)
}

// Warning - Log Warning message
func (e *Entry) Warning(msg string) {
	e.Logger.log(e, LogLevelWarning, msg)
}

// Error - Log Error message
func (e *Entry) Error(msg string) {
	e.Logger.log(e, LogLevelError, msg)
}

// Critical - Log Critical message
func (e *Entry) Critical(msg string) {
	e.Logger.log(e, LogLevelCritical, msg)
}

// Alert - Log Alert message
func (e *Entry) Alert(msg string) {
	e.Logger.log(e, LogLevelAlert, msg)
}

// Emergency - Log Emergency message
func (e *Entry) Emergency(msg string) {
	e.Logger.log(e, LogLevelEmergency, msg)
}

// Fatal - Log Fatal message
func (e *Entry) Fatal(msg string) {
	e.Logger.log(e, LogLevelFatal, msg)
	os.Exit(1)
}

// Panic - Log Panic message
func (e *Entry) Panic(msg string) {
	e.Logger.log(e, LogLevelEmergency, msg)
	panic(msg)
}
