package verbose

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FileTransport writes log messages to a file to a directory
type FileTransport struct {
	Formatter   Formatter
	LockWrites  bool
	min         LogLevel
	max         LogLevel
	path        string
	m           sync.Mutex
	fd          *os.File
	reopenTimer time.Timer
}

const reopenTimerDur = 5 * time.Minute

// NewFileTransport takes the path and returns a FileTransport. If the path exists,
// file or directory mode will be Determined by what path is. If it doesn't exist,
// the mode will be file if path has an extension, otherwise it will be directory.
// In file mode, all log messages are written to a single file.
// In directory mode, each level is written to it's own file.
func NewFileTransport(path string) (*FileTransport, error) {
	return NewFileTransportWith(path, NewLineFormatter(false))
}

func NewFileTransportWith(path string, fmt Formatter) (*FileTransport, error) {
	path, _ = filepath.Abs(path)

	f := &FileTransport{
		Formatter:   fmt,
		min:         LogLevelDebug,
		max:         LogLevelFatal,
		path:        path,
		m:           sync.Mutex{},
		LockWrites:  true,
		reopenTimer: *time.NewTimer(reopenTimerDur),
	}

	if err := f.open(); err != nil {
		return nil, err
	}
	return f, nil
}

func (f *FileTransport) open() error {
	file, err := os.OpenFile(f.path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	f.fd = file
	return err
}

// SetLevel will set both the minimum and maximum log levels to l. This makes
// the handler only respond to the single level l.
func (f *FileTransport) SetLevel(l LogLevel) {
	f.min = l
	f.max = l
}

// SetMinLevel will set the minimum log level the handler will handle.
func (f *FileTransport) SetMinLevel(l LogLevel) {
	if l > f.max {
		return
	}
	f.min = l
}

// SetMaxLevel will set the maximum log level the handler will handle.
func (f *FileTransport) SetMaxLevel(l LogLevel) {
	if l < f.min {
		return
	}
	f.max = l
}

// Handles returns whether the handler handles log level l.
func (f *FileTransport) Handles(l LogLevel) bool {
	return (f.min <= l && l <= f.max)
}

// WriteLog will write the log message to a file.
func (f *FileTransport) WriteLog(e *Entry) {
	data := f.Formatter.FormatByte(e)

	if f.LockWrites {
		f.m.Lock()
		defer f.m.Unlock()
	}

	select {
	case <-f.reopenTimer.C:
		// Reopening the file in case the file handle has changed
		// This fixes issues with logrotate
		if err := f.reopen(); err != nil {
			fmt.Printf("Error writing to log file: %v\n", err)
			return
		}
		f.reopenTimer.Reset(reopenTimerDur)
	default:
	}

	// Attempt first write
	_, err := f.fd.Write(data)
	if err == nil {
		f.fd.Sync()
		return
	}

	// If write failed, try reopening the file
	if err := f.reopen(); err != nil {
		fmt.Printf("Error writing to log file: %v\n", err)
		return
	}

	// If reopen succeeded, try write again
	_, err = f.fd.Write(data)
	if err != nil {
		fmt.Printf("Error writing to log file: %v\n", err)
		return
	}

	f.fd.Sync()
}

// Close satisfies the interface, NOOP
func (f *FileTransport) Close() {
	f.fd.Close()
}

func (f *FileTransport) reopen() error {
	f.fd.Close() // Don't care about any errors
	tries := 0
	var err error

	// Attempt to reopen the file 5 times
	for tries < 5 {
		err = f.open()
		if err == nil {
			break
		}
		tries++
	}

	return err
}
