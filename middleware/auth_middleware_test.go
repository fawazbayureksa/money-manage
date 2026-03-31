package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"my-api/utils"
)

func init() {
	gin.SetMode(gin.TestMode)
	utils.InitLogger()
}

func makeValidToken(userID uint) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": float64(userID),
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})
	str, _ := token.SignedString([]byte("secret"))
	return str
}

func makeExpiredToken(userID uint) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": float64(userID),
		"exp":     time.Now().Add(-time.Hour).Unix(),
	})
	str, _ := token.SignedString([]byte("secret"))
	return str
}

func setupRouter() *gin.Engine {
	r := gin.New()
	r.Use(AuthMiddleware())
	r.GET("/protected", func(c *gin.Context) {
		uid, _ := c.Get("user_id")
		c.JSON(http.StatusOK, gin.H{"user_id": uid})
	})
	return r
}

func TestAuthMiddlewareMissingAuthHeader(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401, got %d", w.Code)
	}
}

func TestAuthMiddlewareInvalidBearerFormat(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "JustOneToken")

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for bad format, got %d", w.Code)
	}
}

func TestAuthMiddlewareInvalidToken(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer this.is.not.valid")

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for invalid token, got %d", w.Code)
	}
}

func TestAuthMiddlewareExpiredToken(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+makeExpiredToken(5))

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for expired token, got %d", w.Code)
	}
}

func TestAuthMiddlewareValidToken(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+makeValidToken(7))

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 for valid token, got %d", w.Code)
	}
}

func TestAuthMiddlewareSetsUserID(t *testing.T) {
	var captured interface{}
	r := gin.New()
	r.Use(AuthMiddleware())
	r.GET("/check", func(c *gin.Context) {
		captured, _ = c.Get("user_id")
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/check", nil)
	req.Header.Set("Authorization", "Bearer "+makeValidToken(42))

	r.ServeHTTP(w, req)

	if captured != uint(42) {
		t.Errorf("Expected user_id uint(42), got %v (type %T)", captured, captured)
	}
}

func TestAuthMiddlewareWrongSecret(t *testing.T) {
	// Token signed with a different secret should be rejected.
	badToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": float64(1),
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	badStr, _ := badToken.SignedString([]byte("wrongsecret"))

	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+badStr)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for wrong-secret token, got %d", w.Code)
	}
}

func TestAuthMiddlewareAbortsPipeline(t *testing.T) {
	// Verify that the next handler is NOT called when auth fails.
	nextCalled := false
	r := gin.New()
	r.Use(AuthMiddleware())
	r.GET("/guarded", func(c *gin.Context) {
		nextCalled = true
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/guarded", nil)
	// No Authorization header.
	r.ServeHTTP(w, req)

	if nextCalled {
		t.Error("Next handler should not be called when auth fails")
	}
}
