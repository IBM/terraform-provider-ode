// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"time"
)

// Level represents the different log levels.
type Level int

const (
	Debug Level = iota
	Info
	Warn
	Error
	Trace
)

// StructuredLogger represents the core logging functionality.
type StructuredLogger struct {
	out      io.Writer
	fmt      Formatter
	colored  bool
	colors   map[Level]string
	mu       sync.Mutex
	minLevel Level
}

// String returns the string representation of the log level.
func (lv Level) String() string {
	switch lv {
	case Debug:
		return "DEBUG"
	case Info:
		return "INFO"
	case Warn:
		return "WARN"
	case Error:
		return "ERROR"
	case Trace:
		return "TRACE"
	default:
		return "UNKNOWN"
	}
}

// defaultColors maps log levels to their corresponding ANSI color codes.
var defaultColors = map[Level]string{
	Debug: "94", // Bright Blue
	Info:  "36", // Cyan
	Warn:  "93", // Bright Yellow
	Error: "91", // Bright Red
	Trace: "95", // Bright Magenta
}

// Formatter is an interface for formatting log messages.
type Formatter func(header string, fields map[string]any) []string

// Inline formats log messages in a single line.
func Inline() Formatter {
	return func(header string, fields map[string]any) []string {
		keys := make([]string, 0, len(fields))
		for k := range fields {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		var b strings.Builder
		b.WriteString(header)
		for _, k := range keys {
			_, err := fmt.Fprintf(&b, " %s=%v", k, fields[k])
			if err != nil {
				return nil
			}
		}
		return []string{b.String()}
	}
}

// JSON formats log messages in JSON format with optional indentation.
func JSON(indent string) Formatter {
	return func(header string, fields map[string]any) []string {
		j, err := json.MarshalIndent(fields, "", indent)
		if err != nil {
			return Inline()(header, fields)
		}
		return append([]string{header}, strings.Split(string(j), "\n")...)
	}
}

// Option is a function that customizes the logger configuration.
type Option func(*StructuredLogger)

// New initializes a new StructuredLogger instance with the provided writer and options.
func New(w io.Writer, opts ...Option) *StructuredLogger {
	l := &StructuredLogger{
		out:     w,
		fmt:     Inline(),
		colored: true,
		colors:  defaultColors,
	}
	for _, o := range opts {
		o(l)
	}
	return l
}

// WithFormatter sets the formatter for the logger.
func WithFormatter(f Formatter) Option {
	return func(l *StructuredLogger) {
		l.fmt = f
	}
}

// WithColor enables or disables color-coded output.
func WithColor(enabled bool) Option {
	return func(l *StructuredLogger) {
		l.colored = enabled
	}
}

func WithMinLevel(lv Level) Option {
	return func(l *StructuredLogger) {
		l.minLevel = lv
	}
}

// WithLevelColors sets custom color codes for each log level.
func WithLevelColors(m map[Level]string) Option {
	return func(l *StructuredLogger) {
		clone := make(map[Level]string, len(m))
		for k, v := range m {
			clone[k] = v
		}
		l.colors = clone
	}
}

// Log writes a log message with the given level, message, and key-value pairs.
func (l *StructuredLogger) Log(level Level, msg string, kvp ...any) {
	if level < l.minLevel {
		return
	}

	fields := map[string]any{"msg": msg}
	for i := 0; i+1 < len(kvp); i += 2 {
		if k, ok := kvp[i].(string); ok {
			fields[k] = kvp[i+1]
		}
	}

	prefix := fmt.Sprintf("[%s] %s:", time.Now().Format("15:04:05.000"), level)
	if l.colored {
		if c, ok := l.colors[level]; ok {
			prefix = fmt.Sprintf("\033[%sm%s\033[0m", c, prefix)
		}
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	for _, line := range l.fmt(prefix, fields) {
		fmt.Fprintln(l.out, line)
	}
}

//lint:ignore U1000 Ignore unused function
func getLogLevel(lvl string) Level {
	switch strings.ToLower(lvl) {
	case "debug":
		return Debug
	case "warn":
		return Warn
	case "error":
		return Error
	case "trace":
		return Trace
	default:
		return Info
	}
}

// Logger interface defines methods for logging messages at different levels.
type Logger interface {
	Debug(msg string, kvs ...any)
	Info(msg string, kvs ...any)
	Warn(msg string, kvs ...any)
	Error(msg string, kvs ...any)
	Trace(msg string, kvs ...any)
}

func (l *StructuredLogger) Debug(msg string, kvs ...any) { l.Log(Debug, msg, kvs...) }
func (l *StructuredLogger) Info(msg string, kvs ...any)  { l.Log(Info, msg, kvs...) }
func (l *StructuredLogger) Warn(msg string, kvs ...any)  { l.Log(Warn, msg, kvs...) }
func (l *StructuredLogger) Error(msg string, kvs ...any) { l.Log(Error, msg, kvs...) }
func (l *StructuredLogger) Trace(msg string, kvs ...any) { l.Log(Trace, msg, kvs...) }
