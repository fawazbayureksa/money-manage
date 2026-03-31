package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSuccessResponseStruct(t *testing.T) {
	resp := SuccessResponse("test message", "test data")
	if !resp.Success {
		t.Error("Expected Success to be true")
	}
	if resp.Message != "test message" {
		t.Errorf("Expected message 'test message', got '%s'", resp.Message)
	}
	if resp.Data != "test data" {
		t.Errorf("Expected data 'test data', got '%v'", resp.Data)
	}
}

func TestSuccessResponseNilData(t *testing.T) {
	resp := SuccessResponse("ok", nil)
	if !resp.Success {
		t.Error("Expected Success to be true")
	}
	if resp.Data != nil {
		t.Errorf("Expected nil data, got '%v'", resp.Data)
	}
}

func TestErrorResponseStruct(t *testing.T) {
	resp := ErrorResponse("error message")
	if resp.Success {
		t.Error("Expected Success to be false")
	}
	if resp.Message != "error message" {
		t.Errorf("Expected message 'error message', got '%s'", resp.Message)
	}
	if resp.Data != nil {
		t.Errorf("Expected data to be nil, got '%v'", resp.Data)
	}
}

func TestJSONSuccessReturns200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	JSONSuccess(c, "success", map[string]string{"key": "value"})

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestJSONSuccessResponseBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	JSONSuccess(c, "created", gin.H{"id": 1})

	var response Response
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if !response.Success {
		t.Error("Expected Success to be true")
	}
	if response.Message != "created" {
		t.Errorf("Expected message 'created', got '%s'", response.Message)
	}
}

func TestJSONSuccessNilData(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	JSONSuccess(c, "ok", nil)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestJSONErrorStatusCode(t *testing.T) {
	statuses := []int{
		http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusNotFound,
		http.StatusConflict,
		http.StatusInternalServerError,
	}

	for _, status := range statuses {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		JSONError(c, status, "error")

		if w.Code != status {
			t.Errorf("Expected status %d, got %d", status, w.Code)
		}
	}
}

func TestJSONErrorResponseBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	JSONError(c, http.StatusBadRequest, "bad request")

	var response Response
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if response.Success {
		t.Error("Expected Success to be false")
	}
	if response.Message != "bad request" {
		t.Errorf("Expected message 'bad request', got '%s'", response.Message)
	}
	if response.Data != nil {
		t.Errorf("Expected data to be nil, got '%v'", response.Data)
	}
}
