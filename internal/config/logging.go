package config

import (
	"fmt"
	"strings"

	"github.com/lfkeitel/verbose/v5"
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
	return verbose.New()
}

func NewLogger(c *Config, name string) *verbose.Logger {
	logger := verbose.New()
	setupStdOutLogging(logger, c.Logging.Level)

	if !c.Logging.Disabled {
		setupFileLogging(logger, c.Logging.Level, c.Logging.Path)
	}

	return logger
}

func setupStdOutLogging(logger *verbose.Logger, level string) {
	sh := verbose.NewTextTransport()
	if level, ok := logLevels[strings.ToLower(level)]; ok {
		sh.SetMinLevel(level)
	}
	logger.AddTransport(sh)
}

func setupFileLogging(logger *verbose.Logger, level, path string) {
	if path == "" {
		return
	}

	fh, err := verbose.NewFileTransportWith(path, verbose.NewJSONFormatter())
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	if level, ok := logLevels[strings.ToLower(level)]; ok {
		fh.SetMinLevel(level)
	}
	logger.AddTransport(fh)
}
