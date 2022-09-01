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
	debugFile, _ = os.Create(filepath.Join(os.TempDir(), "lazytrivy-logger.log"))
}

func Debug(format string, args ...interface{}) {
	log("DEBUG", format, args...)
}

func Error(format string, args ...interface{}) {
	log("ERROR", format, args...)
}

func log(level string, format string, args ...interface{}) {
	if debugEnabled {
		_, _ = fmt.Fprintf(debugFile, fmt.Sprintf("%s [%s] ", time.RFC3339, level)+fmt.Sprintf(format, args...))
		_, _ = fmt.Fprintln(debugFile)
	}
}
