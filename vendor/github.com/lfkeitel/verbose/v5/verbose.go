package verbose

var defaultLogger *Logger

func init() {
	defaultLogger = Simple()
}

func ClearTransports() {
	defaultLogger.ClearTransports()
}

func DefaultLogger() *Logger {
	return defaultLogger
}

func AddTransport(h Transport) {
	defaultLogger.AddTransport(h)
}

func Close() {
	defaultLogger.Close()
}

func WithField(key string, value interface{}) *Entry {
	return defaultLogger.WithField(key, value)
}

func WithFields(fields Fields) *Entry {
	return defaultLogger.WithFields(fields)
}

func Debug(msg string) {
	defaultLogger.Debug(msg)
}

func Info(msg string) {
	defaultLogger.Info(msg)
}

func Notice(msg string) {
	defaultLogger.Notice(msg)
}

func Warning(msg string) {
	defaultLogger.Warning(msg)
}

func Error(msg string) {
	defaultLogger.Error(msg)
}

func Critical(msg string) {
	defaultLogger.Critical(msg)
}

func Alert(msg string) {
	defaultLogger.Alert(msg)
}

func Emergency(msg string) {
	defaultLogger.Emergency(msg)
}

func Fatal(msg string) {
	defaultLogger.Fatal(msg)
}

func Panic(msg string) {
	defaultLogger.Panic(msg)
}
