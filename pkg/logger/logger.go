package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
)

type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

var (
	logLevel  = InfoLevel
	stdLogger = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
)

func SetLevel(level Level) {
	logLevel = level
}

// ParseLevel แปลงชื่อ level (debug, info, warn/warning, error ไม่สนตัวพิมพ์เล็กใหญ่) เป็น Level
func ParseLevel(name string) (Level, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "debug":
		return DebugLevel, nil
	case "info":
		return InfoLevel, nil
	case "warn", "warning":
		return WarnLevel, nil
	case "error":
		return ErrorLevel, nil
	}
	return InfoLevel, fmt.Errorf("unknown log level %q", name)
}

func Debug(v ...any) {
	if logLevel <= DebugLevel {
		stdLogger.SetPrefix("[DEBUG] ")
		_ = stdLogger.Output(2, sprint(v...))
	}
}

func Info(v ...any) {
	if logLevel <= InfoLevel {
		stdLogger.SetPrefix("[INFO] ")
		_ = stdLogger.Output(2, sprint(v...))
	}
}

func Warn(v ...any) {
	if logLevel <= WarnLevel {
		stdLogger.SetPrefix("[WARN] ")
		_ = stdLogger.Output(2, sprint(v...))
	}
}

func Error(v ...any) {
	if logLevel <= ErrorLevel {
		stdLogger.SetPrefix("[ERROR] ")
		_ = stdLogger.Output(2, sprint(v...))
	}
}

func sprint(v ...any) string {
	return fmt.Sprintln(v...)
}
