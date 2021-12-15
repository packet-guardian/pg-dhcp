# Verbose - Logging for Go

[![GoDoc](https://godoc.org/github.com/lfkeitel/verbose?status.svg)](https://godoc.org/github.com/lfkeitel/verbose)

Verbose is a library for simple, organized structured logging. Verbose is a very
flexible library allowing for multiple log outputs and formatters for each
output ensuring data is always formatted correctly for each destination.

## Usage

The easiest way to start with Verbose is to use the package level logger:

```go
import (
    log "github.com/lfkeitel/verbose/v5"
)

func main() {
    log.WithFields(log.Fields{
        "path": "/",
        "user": "Alice",
    }).Info("Received request")

    // Or without fields
    log.Info("Hello")
}
```

You can also create loggers to pass around:

```go
import (
    log "github.com/lfkeitel/verbose/v5"
)

func main() {
    logger := log.New()
    ft, _ := log.NewFileTransport("applogs.txt") // Ignoring errors
    logger.AddTransport(ft)

    logger.WithFields(log.Fields{
        "path": "/",
        "user": "Alice",
    }).Info("Received request")

    // Or without fields
    logger.Info("Hello")
}
```

## Supported Log Levels

- Debug
- Info
- Notice
- Warning
- Error
- Critical
- Alert
- Emergency
- Fatal (calls os.Exit(1))

You can also use the following functions:

- Print
- Panic (calls panic() after writing log)

Verbose does not facilitate formatted log message. Instead, structured logging
is preferred and highly encouraged.

## Structured Logging

```go
logger.WithField("field 1", data)

logger.WithFields(verbose.Fields{
    "field 1": "value 1",
    "field 2": 42,
}).Debug("This is a debug message")
```

The fields will be formatted appropriately by the handler.

## Transports

A Logger initially is nothing more than a shell. Without transports it won't do
anything. This library comes with two transports. You can use your own
transports so long as they satisfy the verbose.Handler interface. You can add a
handler by calling `logger.AddHandler()`. A Logger will loop through all the
transports and send the message to any that report they can handle the log
level.

### TextTransport

The TextTransport will print colored log messages to stderr. The output can be
changed by calling `.SetOutput()` with an io.Writer.

```go
sh := verbose.NewTextTransport()
```

The text transport will use color if the output is a valid terminal. Otherwise
color is disabled. To force color to be enabled or disabled, create a custom
LineFormatter with color set to true or false.

### FileTransport

The FileTransport will write log messages to a file. If the file exists, it will
be appended to. Otherwise the file will be created.

```go
fh := verbose.NewFileTransport(path)
```

The file transport uses a non-color line formatter by default. You can change
the formatter by setting `.Formatter` on the transport object.

## Formatters

A formatter is used to actually construct a log line that a transport will then
store or display. This library comes with 3 formatters but anything satisfying
the interface can be used.

### Time Format

The time format used by formatters can be set using the
Formatter.SetTimeFormat() method. The default time format for included
formatters is RFC3339: "2006-01-02T15:04:05Z07:00". The time format can be any
valid Go time format.

### JSONFormatter

The JSON formatter is great when the logs are being processed by a centralized
logging solution or some other computerized system. It will generate a JSON
object with the following structure:

```json
{
    "timestamp": "1970-01-01T12:00:00Z",
    "level": "INFO",
    "logger": "app",
    "message": "Hello, world",
    "data": {
        "field 1": "data 1",
        "field 2": "data 2"
    }
}
```

Any structured fields will go in the data object.

### LineFormatter

The line formatter is designed to be human readable either for a file that will
mainly be viewed by humans, or for standard output. A sample output line would
be:

```
1970-01-01T12:00:00Z: INFO: app: message: | "field 1": "value 1", "field 2": "value 2"
```

The formatter can format with or without color. To change the color setting
after creation, set the `.UseColor` field.

### LogfmtFormatter

The logfmt formatter formats logs into the [logfmt](https://www.brandur.org/logfmt) format.

```
timestamp="2019-09-17T15:54:51-05:00" level=INFO logger="" msg="This happened" foo="bar" result="3"
```

## Release Notes

v5.0.0

- Renamed handlers to transports
- Simplified API by removing all format and line print functions
- Simplified Transport and Formatter interfaces
- Added a package level logger
- Added [logfmt](https://www.brandur.org/logfmt) formatter
- Made logger name optional
- Removed logger store, the library no longer maintains a repository of created
  loggers

v4.0.0

- Expanded Formatter interface
    - SetTimeFormat(string)
- Added generator functions for Formatters
    - NewJSONFormatter()
    - NewLineFormatter()
    - NewColoredLineFormatter()
- Use RFC3339 as the default time format

v3.0.0

- Expanded Handler interface
    - SetFormatter(Formatter)
    - SetLevel(LogLevel)
    - SetMinLevel(LogLevel)
    - SetMaxLevel(LogLevel)
- Added support for formatters
    - Included formatters:
        - JSON
        - Line
        - Line with Color
- Use Fatal as the default Handler max for StdOut and FileHandlers

v2.0.0

- Added support for structured logging
- Removed LogLevelCustom
- Added [x]ln() functions to be compatible with the std lib logger

v1.0.0

- Initial Release

## Versioning

For transparency into the release cycle and in striving to maintain backward
compatibility, this application is maintained under the Semantic Versioning
guidelines. Sometimes I screw up, but I'll adhere to these rules whenever
possible.

Releases will be numbered with the following format:

`<major>.<minor>.<patch>`

And constructed with the following guidelines:

- Breaking backward compatibility **bumps the major** while resetting minor and
  patch
- New additions without breaking backward compatibility **bumps the minor**
  while resetting the patch
- Bug fixes and misc changes **bumps only the patch**

For more information on SemVer, please visit <http://semver.org/>.

## License

This package is released under the terms of the MIT license. Please see LICENSE
for more information.
