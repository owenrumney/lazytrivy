package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var (
	debugEnabled bool
	debugFile    *os.File
)

func EnableDebugging() {
	debugEnabled = true

	home, err := os.UserHomeDir()
	if err != nil {
		home = "/"
	}
	logDir := filepath.Join(home, ".lazytrivy", "logs")
	_ = os.MkdirAll(logDir, os.ModePerm)

	logFile := filepath.Join(logDir, "lazytrivy.log")
	debugFile, _ = os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600) // nolint: nosnakecase
}

func EnableTracing() {
	EnableDebugging()
}

func Tracef(format string, args ...interface{}) {
	logf("TRACE", format, args...)
}

func Debugf(format string, args ...interface{}) {
	logf("DEBUG", format, args...)
}

func Errorf(format string, args ...interface{}) {
	logf("ERROR", format, args...)
}

func logf(level string, format string, args ...interface{}) {
	if debugEnabled {
		_, _ = fmt.Fprintf(debugFile, fmt.Sprintf("%s [%s] ", time.Now().Format(time.RFC3339), level)+fmt.Sprintf(format, args...))
		_, _ = fmt.Fprintln(debugFile)
	}
}
