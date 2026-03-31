package utils

import (
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
)

func TestGenerateToken(t *testing.T) {
	token, err := GenerateToken(1)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	if token == "" {
		t.Error("Token should not be empty")
	}
}

func TestGenerateTokenIsParseable(t *testing.T) {
	userID := uint(42)
	tokenStr, err := GenerateToken(userID)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte("secret"), nil
	})
	if err != nil {
		t.Fatalf("Failed to parse token: %v", err)
	}
	if !token.Valid {
		t.Error("Token should be valid")
	}
}

func TestGenerateTokenContainsUserID(t *testing.T) {
	userID := uint(99)
	tokenStr, _ := GenerateToken(userID)

	token, _ := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte("secret"), nil
	})
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("Failed to get claims")
	}
	uid, ok := claims["user_id"].(float64)
	if !ok {
		t.Fatal("user_id claim missing or wrong type")
	}
	if uint(uid) != userID {
		t.Errorf("Expected user_id %d, got %d", userID, uint(uid))
	}
}

func TestGenerateTokenContainsExpiry(t *testing.T) {
	tokenStr, _ := GenerateToken(1)

	token, _ := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte("secret"), nil
	})
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("Failed to get claims")
	}
	if _, exists := claims["exp"]; !exists {
		t.Error("Token should contain 'exp' claim")
	}
}

func TestGenerateTokenExpiresInFuture(t *testing.T) {
	tokenStr, _ := GenerateToken(1)

	token, _ := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte("secret"), nil
	})
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("Failed to get claims")
	}
	exp, ok := claims["exp"].(float64)
	if !ok {
		t.Fatal("exp claim is not a float64")
	}
	expTime := time.Unix(int64(exp), 0)
	if !expTime.After(time.Now()) {
		t.Error("Token expiry should be in the future")
	}
}

func TestGenerateTokenDifferentForDifferentUsers(t *testing.T) {
	token1, _ := GenerateToken(1)
	token2, _ := GenerateToken(2)

	if token1 == token2 {
		t.Error("Tokens for different users should be different")
	}
}

func TestGenerateTokenUsesHS256(t *testing.T) {
	tokenStr, _ := GenerateToken(1)

	token, _ := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte("secret"), nil
	})
	if token.Method != jwt.SigningMethodHS256 {
		t.Errorf("Expected HS256 signing method, got %v", token.Method)
	}
}
