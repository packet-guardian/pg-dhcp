package verbose

// LogLevel is used to compare levels in a consistant manner
type LogLevel int

// These are the defined log levels
const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelNotice
	LogLevelWarning
	LogLevelError
	LogLevelCritical
	LogLevelAlert
	LogLevelEmergency
	LogLevelFatal
)

// LogLevel to stringified versions
var levelString = map[LogLevel]string{
	LogLevelDebug:     "Debug",
	LogLevelInfo:      "Info",
	LogLevelNotice:    "Notice",
	LogLevelWarning:   "Warning",
	LogLevelError:     "Error",
	LogLevelCritical:  "Critical",
	LogLevelAlert:     "Alert",
	LogLevelEmergency: "Emergency",
	LogLevelFatal:     "Fatal",
}

// String returns the stringified version of LogLevel.
// I.e., "Error" for LogLevelError, and "Debug" for LogLevelDebug
// It will return an empty string for any undefined level.
func (l LogLevel) String() string {
	if s, ok := levelString[l]; ok {
		return s
	}
	return ""
}
