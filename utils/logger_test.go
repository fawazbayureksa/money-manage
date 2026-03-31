package utils

import (
	"bytes"
	"log"
	"testing"
)

func TestInitLogger(t *testing.T) {
	InitLogger()
	if InfoLogger == nil {
		t.Error("InfoLogger should not be nil after initialization")
	}
	if WarningLogger == nil {
		t.Error("WarningLogger should not be nil after initialization")
	}
	if ErrorLogger == nil {
		t.Error("ErrorLogger should not be nil after initialization")
	}
}

func TestLogInfoNilSafe(t *testing.T) {
	saved := InfoLogger
	InfoLogger = nil
	LogInfo("should not panic")
	InfoLogger = saved
}

func TestLogWarningNilSafe(t *testing.T) {
	saved := WarningLogger
	WarningLogger = nil
	LogWarning("should not panic")
	WarningLogger = saved
}

func TestLogErrorNilSafe(t *testing.T) {
	saved := ErrorLogger
	ErrorLogger = nil
	LogError("should not panic")
	ErrorLogger = saved
}

func TestLogInfofNilSafe(t *testing.T) {
	saved := InfoLogger
	InfoLogger = nil
	LogInfof("should not panic %d", 1)
	InfoLogger = saved
}

func TestLogWarningfNilSafe(t *testing.T) {
	saved := WarningLogger
	WarningLogger = nil
	LogWarningf("should not panic %d", 1)
	WarningLogger = saved
}

func TestLogErrorfNilSafe(t *testing.T) {
	saved := ErrorLogger
	ErrorLogger = nil
	LogErrorf("should not panic %d", 1)
	ErrorLogger = saved
}

func TestLogInfoWritesMessage(t *testing.T) {
	var buf bytes.Buffer
	InfoLogger = log.New(&buf, "INFO: ", 0)

	LogInfo("test info message")

	if buf.Len() == 0 {
		t.Error("Expected log output, got nothing")
	}
	if !bytes.Contains(buf.Bytes(), []byte("test info message")) {
		t.Errorf("Expected log to contain 'test info message', got: %s", buf.String())
	}
}

func TestLogInfofFormatsMessage(t *testing.T) {
	var buf bytes.Buffer
	InfoLogger = log.New(&buf, "INFO: ", 0)

	LogInfof("test %s %d", "formatted", 42)

	if !bytes.Contains(buf.Bytes(), []byte("test formatted 42")) {
		t.Errorf("Expected formatted log, got: %s", buf.String())
	}
}

func TestLogWarningWritesMessage(t *testing.T) {
	var buf bytes.Buffer
	WarningLogger = log.New(&buf, "WARNING: ", 0)

	LogWarning("test warning message")

	if !bytes.Contains(buf.Bytes(), []byte("test warning message")) {
		t.Errorf("Expected log to contain 'test warning message', got: %s", buf.String())
	}
}

func TestLogWarningfFormatsMessage(t *testing.T) {
	var buf bytes.Buffer
	WarningLogger = log.New(&buf, "WARNING: ", 0)

	LogWarningf("warning %s %d", "code", 404)

	if !bytes.Contains(buf.Bytes(), []byte("warning code 404")) {
		t.Errorf("Expected formatted warning log, got: %s", buf.String())
	}
}

func TestLogErrorWritesMessage(t *testing.T) {
	var buf bytes.Buffer
	ErrorLogger = log.New(&buf, "ERROR: ", 0)

	LogError("test error message")

	if !bytes.Contains(buf.Bytes(), []byte("test error message")) {
		t.Errorf("Expected log to contain 'test error message', got: %s", buf.String())
	}
}

func TestLogErrorfFormatsMessage(t *testing.T) {
	var buf bytes.Buffer
	ErrorLogger = log.New(&buf, "ERROR: ", 0)

	LogErrorf("error %s %d", "code", 500)

	if !bytes.Contains(buf.Bytes(), []byte("error code 500")) {
		t.Errorf("Expected formatted error log, got: %s", buf.String())
	}
}
