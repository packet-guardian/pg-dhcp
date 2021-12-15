package verbose

import (
	"fmt"
	"io"
	"os"

	"github.com/lfkeitel/verbose/v5/tty"
)

// Color is an escaped color code for the terminal
type Color string

// Pre-defined colors
const (
	ColorReset   Color = "\033[0m"
	ColorRed     Color = "\033[31m"
	ColorGreen   Color = "\033[32m"
	ColorYellow  Color = "\033[33m"
	ColorBlue    Color = "\033[34m"
	ColorMagenta Color = "\033[35m"
	ColorCyan    Color = "\033[36m"
	ColorWhite   Color = "\033[37m"
	ColorGrey    Color = "\033[90m"
)

var colors = map[LogLevel]Color{
	LogLevelDebug:     ColorBlue,
	LogLevelInfo:      ColorCyan,
	LogLevelNotice:    ColorCyan,
	LogLevelWarning:   ColorMagenta,
	LogLevelError:     ColorRed,
	LogLevelCritical:  ColorRed,
	LogLevelAlert:     ColorRed,
	LogLevelEmergency: ColorRed,
	LogLevelFatal:     ColorRed,
}

// TextTransport writes log message to standard out
// It even uses color!
type TextTransport struct {
	Formatter Formatter
	min       LogLevel
	max       LogLevel
	out       io.Writer // Usually os.Stdout, mainly used for testing
}

// NewTextTransport creates a new TextTransport, surprise!
// Color specifies if the log messages will be printed to a colored terminal.
func NewTextTransport() *TextTransport {
	return NewTextTransportWith(NewLineFormatter(tty.CheckIfTerminal(os.Stderr)))
}

func NewTextTransportWith(fmt Formatter) *TextTransport {
	return &TextTransport{
		min:       LogLevelDebug,
		max:       LogLevelFatal,
		out:       os.Stderr,
		Formatter: fmt,
	}
}

func (s *TextTransport) SetOutput(out io.Writer) {
	s.out = out
}

// SetLevel will set both the minimum and maximum log levels to l. This makes
// the handler only respond to the single level l.
func (s *TextTransport) SetLevel(l LogLevel) {
	s.min = l
	s.max = l
}

// SetMinLevel will set the minimum log level the handler will handle.
func (s *TextTransport) SetMinLevel(l LogLevel) {
	if l > s.max {
		return
	}
	s.min = l
}

// SetMaxLevel will set the maximum log level the handler will handle.
func (s *TextTransport) SetMaxLevel(l LogLevel) {
	if l < s.min {
		return
	}
	s.max = l
}

// Handles returns whether the handler handles log level l.
func (s *TextTransport) Handles(l LogLevel) bool {
	return (s.min <= l && l <= s.max)
}

// WriteLog writes the log message to standard output
func (s *TextTransport) WriteLog(e *Entry) {
	fmt.Fprint(s.out, s.Formatter.Format(e))
}

// Close satisfies the interface, NOOP
func (s *TextTransport) Close() {}
