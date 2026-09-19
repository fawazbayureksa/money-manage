package dto

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.POST("/debts", func(c *gin.Context) {
		var req CreateDebtRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, req)
	})

	r.PUT("/debts/balance", func(c *gin.Context) {
		var req UpdateDebtBalanceRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, req)
	})

	return r
}

func executeRequest(r *gin.Engine, method, url, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, url, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestCreateDebtRequest_Validation(t *testing.T) {
	r := setupTestRouter()

	tests := []struct {
		name       string
		payload    string
		expectCode int
	}{
		{
			name: "Valid with zero interest rate",
			payload: `{
				"name": "Personal Loan 0%",
				"debt_type": "personal_loan",
				"original_amount": 5000000,
				"current_balance": 5000000,
				"interest_rate": 0,
				"minimum_payment": 500000,
				"payment_due_day": 10,
				"start_date": "2026-01-01"
			}`,
			expectCode: http.StatusOK,
		},
		{
			name: "Valid with positive interest rate",
			payload: `{
				"name": "BCA Credit Card",
				"debt_type": "credit_card",
				"original_amount": 10000000,
				"current_balance": 8000000,
				"interest_rate": 24.5,
				"minimum_payment": 800000,
				"payment_due_day": 15,
				"start_date": "2026-01-01"
			}`,
			expectCode: http.StatusOK,
		},
		{
			name: "Valid with omitted interest rate (defaults to 0)",
			payload: `{
				"name": "Borrow from friend",
				"debt_type": "other",
				"original_amount": 2000000,
				"current_balance": 2000000,
				"minimum_payment": 200000,
				"payment_due_day": 5,
				"start_date": "2026-01-01"
			}`,
			expectCode: http.StatusOK,
		},
		{
			name: "Valid with zero current balance",
			payload: `{
				"name": "Past Loan",
				"debt_type": "personal_loan",
				"original_amount": 1000000,
				"current_balance": 0,
				"interest_rate": 0,
				"minimum_payment": 100000,
				"payment_due_day": 1,
				"start_date": "2026-01-01"
			}`,
			expectCode: http.StatusOK,
		},
		{
			name: "Invalid negative interest rate",
			payload: `{
				"name": "Invalid Debt",
				"debt_type": "other",
				"original_amount": 1000000,
				"current_balance": 1000000,
				"interest_rate": -1,
				"minimum_payment": 100000,
				"payment_due_day": 1,
				"start_date": "2026-01-01"
			}`,
			expectCode: http.StatusBadRequest,
		},
		{
			name: "Invalid interest rate greater than 100",
			payload: `{
				"name": "Invalid Debt",
				"debt_type": "other",
				"original_amount": 1000000,
				"current_balance": 1000000,
				"interest_rate": 150,
				"minimum_payment": 100000,
				"payment_due_day": 1,
				"start_date": "2026-01-01"
			}`,
			expectCode: http.StatusBadRequest,
		},
		{
			name: "Invalid missing name",
			payload: `{
				"debt_type": "other",
				"original_amount": 1000000,
				"current_balance": 1000000,
				"interest_rate": 5,
				"minimum_payment": 100000,
				"payment_due_day": 1,
				"start_date": "2026-01-01"
			}`,
			expectCode: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := executeRequest(r, "POST", "/debts", tc.payload)
			if w.Code != tc.expectCode {
				t.Errorf("expected status %d, got %d, body: %s", tc.expectCode, w.Code, w.Body.String())
			}
		})
	}
}

func TestUpdateDebtBalanceRequest_Validation(t *testing.T) {
	r := setupTestRouter()

	tests := []struct {
		name       string
		payload    string
		expectCode int
	}{
		{
			name: "Valid with zero balance",
			payload: `{
				"new_balance": 0,
				"as_of_date": "2026-02-11"
			}`,
			expectCode: http.StatusOK,
		},
		{
			name: "Valid with positive balance",
			payload: `{
				"new_balance": 500000,
				"as_of_date": "2026-02-11"
			}`,
			expectCode: http.StatusOK,
		},
		{
			name: "Invalid with negative balance",
			payload: `{
				"new_balance": -100,
				"as_of_date": "2026-02-11"
			}`,
			expectCode: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := executeRequest(r, "PUT", "/debts/balance", tc.payload)
			if w.Code != tc.expectCode {
				t.Errorf("expected status %d, got %d, body: %s", tc.expectCode, w.Code, w.Body.String())
			}
		})
	}
}
