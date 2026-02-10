package utils

import (
	"log"
	"os"
)

var (
	InfoLogger    *log.Logger
	WarningLogger *log.Logger
	ErrorLogger   *log.Logger
)

func InitLogger() {
	InfoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarningLogger = log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func LogInfo(message string) {
	if InfoLogger != nil {
		InfoLogger.Println(message)
	}
}

func LogWarning(message string) {
	if WarningLogger != nil {
		WarningLogger.Println(message)
	}
}

func LogError(message string) {
	if ErrorLogger != nil {
		ErrorLogger.Println(message)
	}
}

func LogInfof(format string, v ...interface{}) {
	if InfoLogger != nil {
		InfoLogger.Printf(format, v...)
	}
}

func LogWarningf(format string, v ...interface{}) {
	if WarningLogger != nil {
		WarningLogger.Printf(format, v...)
	}
}

func LogErrorf(format string, v ...interface{}) {
	if ErrorLogger != nil {
		ErrorLogger.Printf(format, v...)
	}
}
