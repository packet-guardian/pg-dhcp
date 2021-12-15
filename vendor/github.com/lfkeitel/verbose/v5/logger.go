package verbose

import (
	"time"
)

// A Transport is an object that can be used by the Logger to log a message
type Transport interface {
	// Handles returns if it wants to handle a particular log level
	// This can be used to suppress the higher log levels in production.
	Handles(LogLevel) bool

	// WriteLog actually logs the message using any system the Handler wishes.
	// The Handler only needs to accept an Event.
	WriteLog(*Entry)

	// Close is used to give a handler a chance to close any open resources.
	Close()
}

// Fields type, used to pass to `WithFields`.
type Fields map[string]interface{}

// Simple creates a logger with both a TextTransport
// The transport uses its default min and max levels.
func Simple() *Logger {
	l := New()
	l.AddTransport(NewTextTransport())
	return l
}

// A Logger takes a message and writes it to as many handlers as possible
type Logger struct {
	Name       string
	transports []Transport
}

// New will create a new Logger with name n. If with the same name
// already exists, it will be replaced with the new logger.
func New() *Logger {
	return &Logger{
		transports: make([]Transport, 0, 2),
	}
}

// AddTransport will add Handler h to the logger named n. If a handler with
// the same name already exists, it will be overwritten.
func (l *Logger) AddTransport(h Transport) {
	if h != nil {
		l.transports = append(l.transports, h)
	}
}

func (l *Logger) ClearTransports() {
	l.transports = make([]Transport, 0, 2)
}

// Close calls Close() on all the handlers then removes itself from the logger registry
func (l *Logger) Close() {
	for _, h := range l.transports {
		h.Close()
	}
}

// Log is the generic function to log a message with the handlers.
// All other logging functions are simply wrappers around this.
func (l *Logger) log(e *Entry, level LogLevel, msg string) {
	e.Level = level
	e.Message = msg
	e.Timestamp = time.Now()
	for _, h := range l.transports {
		if h.Handles(level) {
			h.WriteLog(e)
		}
	}
}

// WithField creates an Entry with a single field
func (l *Logger) WithField(key string, value interface{}) *Entry {
	return NewEntry(l).WithFields(Fields{key: value})
}

// WithFields creates an Entry with multiple fields
func (l *Logger) WithFields(fields Fields) *Entry {
	return NewEntry(l).WithFields(fields)
}

// Debug - Log Debug message
func (l *Logger) Debug(msg string) {
	NewEntry(l).Debug(msg)
}

// Info - Log Info message
func (l *Logger) Info(msg string) {
	NewEntry(l).Info(msg)
}

// Notice - Log Notice message
func (l *Logger) Notice(msg string) {
	NewEntry(l).Notice(msg)
}

// Warning - Log Warning message
func (l *Logger) Warning(msg string) {
	NewEntry(l).Warning(msg)
}

// Error - Log Error message
func (l *Logger) Error(msg string) {
	NewEntry(l).Error(msg)
}

// Critical - Log Critical message
func (l *Logger) Critical(msg string) {
	NewEntry(l).Critical(msg)
}

// Alert - Log Alert message
func (l *Logger) Alert(msg string) {
	NewEntry(l).Alert(msg)
}

// Emergency - Log Emergency message
func (l *Logger) Emergency(msg string) {
	NewEntry(l).Emergency(msg)
}

// Fatal - Log Fatal message
func (l *Logger) Fatal(msg string) {
	NewEntry(l).Fatal(msg)
}

// Panic - Log Panic message
func (l *Logger) Panic(msg string) {
	NewEntry(l).Panic(msg)
}
