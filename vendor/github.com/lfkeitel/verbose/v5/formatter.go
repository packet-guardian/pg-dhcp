package verbose

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Formatter interface {
	Format(*Entry) string
	FormatByte(*Entry) []byte
}

type JSONFormatter struct {
	timeFormat string
}

func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{
		timeFormat: time.RFC3339,
	}
}

func (j *JSONFormatter) Format(e *Entry) string {
	return string(j.FormatByte(e))
}

func (j *JSONFormatter) FormatByte(e *Entry) []byte {
	data := map[string]interface{}{
		"timestamp": e.Timestamp.Format(j.timeFormat),
		"level":     strings.ToUpper(e.Level.String()),
		"logger":    e.Logger.Name,
		"message":   e.Message,
		"data":      e.Data,
	}

	bytes, _ := json.Marshal(data)
	return append(bytes, '\n')
}

func (j *JSONFormatter) SetTimeFormat(f string) {
	j.timeFormat = f
}

type LineFormatter struct {
	timeFormat string
	UseColor   bool
}

func NewLineFormatter(color bool) *LineFormatter {
	return &LineFormatter{
		timeFormat: time.RFC3339,
		UseColor:   color,
	}
}

func (l *LineFormatter) Format(e *Entry) string {
	return string(l.FormatByte(e))
}

func (l *LineFormatter) FormatByte(e *Entry) []byte {
	if l.UseColor {
		return l.formatColor(e)
	}
	return l.formatNoColor(e)
}

func (l *LineFormatter) formatNoColor(e *Entry) []byte {
	buf := &bytes.Buffer{}
	fmt.Fprintf(
		buf,
		"%s: %s:",
		e.Timestamp.Format(l.timeFormat),
		strings.ToUpper(e.Level.String()),
	)

	if e.Logger.Name != "" {
		fmt.Fprintf(
			buf,
			" %s:",
			e.Logger.Name,
		)
	}

	fmt.Fprintf(
		buf,
		" %s",
		e.Message,
	)

	dataLen := len(e.Data)
	if dataLen > 0 {
		buf.WriteString(" |")
		for k, v := range e.Data {
			fmt.Fprintf(buf, ` "%s": %#v`, k, v)
			if dataLen > 1 {
				buf.WriteByte(',')
			}
			dataLen--
		}
	}
	buf.WriteByte('\n')
	return buf.Bytes()
}

func (l *LineFormatter) formatColor(e *Entry) []byte {
	buf := &bytes.Buffer{}

	fmt.Fprintf(
		buf,
		"%s%s: %s%s:",
		ColorGrey,
		e.Timestamp.Format(l.timeFormat),
		colors[e.Level],
		strings.ToUpper(e.Level.String()),
	)

	if e.Logger.Name != "" {
		fmt.Fprintf(
			buf,
			" %s%s:",
			ColorGreen,
			e.Logger.Name,
		)
	}

	fmt.Fprintf(
		buf,
		" %s%s",
		ColorReset,
		e.Message,
	)

	dataLen := len(e.Data)
	if dataLen > 0 {
		buf.WriteString(" |")
		for k, v := range e.Data {
			fmt.Fprintf(buf, ` "%s": %#v`, k, v)
			if dataLen > 1 {
				buf.WriteByte(',')
			}
			dataLen--
		}
	}
	buf.WriteByte('\n')
	return buf.Bytes()
}

func (l *LineFormatter) SetTimeFormat(f string) {
	l.timeFormat = f
}

type LogFmtFormatter struct {
	timeFormat string
}

func NewLogFmtFormatter() *LogFmtFormatter {
	return &LogFmtFormatter{
		timeFormat: time.RFC3339,
	}
}

func (l *LogFmtFormatter) Format(e *Entry) string {
	lines := make([]string, 0, len(e.Data)+4)

	lines = append(lines, fmt.Sprintf(`timestamp="%s"`, e.Timestamp.Format(l.timeFormat)))
	lines = append(lines, fmt.Sprintf(`level=%s`, strings.ToUpper(e.Level.String())))
	lines = append(lines, fmt.Sprintf(`logger="%s"`, e.Logger.Name))
	lines = append(lines, fmt.Sprintf(`msg="%s"`, e.Message))

	for k, v := range e.Data {
		lines = append(lines, fmt.Sprintf(`%s="%v"`, strings.ReplaceAll(k, " ", "_"), v))
	}

	return strings.Join(lines, " ") + "\n"
}

func (l *LogFmtFormatter) FormatByte(e *Entry) []byte {
	return []byte(l.Format(e))
}

func (l *LogFmtFormatter) SetTimeFormat(f string) {
	l.timeFormat = f
}
