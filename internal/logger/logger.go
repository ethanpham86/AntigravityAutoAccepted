package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"

	"golang.org/x/sys/windows"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelError
)

var (
	currentLevel Level = LevelInfo
	logFile      *os.File
	mu           sync.Mutex
	sysLogger    *log.Logger
)

// Init initializes the application logger with a level and output file.
// Valid levels: "debug", "info", "error"
func Init(levelStr, filename string) error {
	mu.Lock()
	defer mu.Unlock()

	// Parse level
	switch strings.ToLower(levelStr) {
	case "debug":
		currentLevel = LevelDebug
	case "error":
		currentLevel = LevelError
	default:
		currentLevel = LevelInfo // default
	}

	// Open file
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file %s: %v", filename, err)
	}
	logFile = f

	// Build writer: logFile ALWAYS first (guaranteed to work),
	// then stdout ONLY if it's a valid console handle.
	// This prevents io.MultiWriter from aborting on invalid stdout
	// (which happens in -H windowsgui builds where stdout is NUL/invalid).
	var writer io.Writer
	if isStdoutValid() {
		writer = io.MultiWriter(logFile, os.Stdout) // file first, stdout second
	} else {
		writer = logFile // file only — stdout is broken/invalid
	}

	sysLogger = log.New(writer, "", log.LstdFlags)

	return nil
}

// isStdoutValid checks if os.Stdout is attached to a real console or pipe.
// Returns false for Windows GUI-subsystem apps where stdout handle is invalid.
func isStdoutValid() bool {
	handle := windows.Handle(os.Stdout.Fd())
	if handle == windows.InvalidHandle {
		return false
	}
	// Try GetConsoleMode — works for real console handles
	var mode uint32
	err := windows.GetConsoleMode(handle, &mode)
	if err == nil {
		return true // stdout is a console
	}
	// Not a console — could be a pipe or file redirect, still valid
	// Check by trying GetFileType
	ft, _ := windows.GetFileType(handle)
	return ft != windows.FILE_TYPE_UNKNOWN
}

// Close closes the underlying log file
func Close() {
	mu.Lock()
	defer mu.Unlock()
	if logFile != nil {
		logFile.Close()
	}
}

// Debug logs debug messages
func Debug(format string, v ...interface{}) {
	if currentLevel <= LevelDebug {
		sysLogger.Output(2, "[DEBUG] "+fmt.Sprintf(format, v...))
	}
}

// Info logs standard information
func Info(format string, v ...interface{}) {
	if currentLevel <= LevelInfo {
		sysLogger.Output(2, "[INFO]  "+fmt.Sprintf(format, v...))
	}
}

// Click logs click events with a dedicated tag for easy grep and readability.
// Format: [CLICK] ✓ "Accept" @ (450,320) | OCR conf=95% | BG
func Click(format string, v ...interface{}) {
	sysLogger.Output(2, "[CLICK] ✓ "+fmt.Sprintf(format, v...))
}

// Error logs error messages
func Error(format string, v ...interface{}) {
	if currentLevel <= LevelError {
		sysLogger.Output(2, "[ERROR] "+fmt.Sprintf(format, v...))
	}
}

// Fatal logs an error and exits
func Fatal(format string, v ...interface{}) {
	sysLogger.Output(2, "[FATAL] "+fmt.Sprintf(format, v...))
	Close()
	os.Exit(1)
}
