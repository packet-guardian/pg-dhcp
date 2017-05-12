package config

import (
	"fmt"
	"strings"

	"github.com/lfkeitel/verbose"
)

var logLevels = map[string]verbose.LogLevel{
	"debug":     verbose.LogLevelDebug,
	"info":      verbose.LogLevelInfo,
	"notice":    verbose.LogLevelNotice,
	"warning":   verbose.LogLevelWarning,
	"error":     verbose.LogLevelError,
	"critical":  verbose.LogLevelCritical,
	"alert":     verbose.LogLevelAlert,
	"emergency": verbose.LogLevelEmergency,
	"fatal":     verbose.LogLevelFatal,
}

func NewEmptyLogger() *verbose.Logger {
	return verbose.New("null")
}

func NewLogger(c *Config, name string) *verbose.Logger {
	logger := verbose.New(name)
	if c.Logging.Disabled {
		return logger
	}

	setupStdOutLogging(logger, c.Logging.Level)
	setupFileLogging(logger, c.Logging.Level, c.Logging.Path)

	return logger
}

func setupStdOutLogging(logger *verbose.Logger, level string) {
	sh := verbose.NewStdoutHandler(true)
	setMinimumLoggingLevel(sh, level)
	logger.AddHandler("stdout", sh)
}

func setupFileLogging(logger *verbose.Logger, level, path string) {
	if path == "" {
		return
	}

	fh, err := verbose.NewFileHandler(path)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fh.SetFormatter(verbose.NewJSONFormatter())
	setMinimumLoggingLevel(fh, level)
	logger.AddHandler("file", fh)
}

func setMinimumLoggingLevel(logger verbose.Handler, level string) {
	if level, ok := logLevels[strings.ToLower(level)]; ok {
		logger.SetMinLevel(level)
	}
}
