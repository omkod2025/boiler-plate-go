package logger

import (
	"fmt"
	"log"
	"os"
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
