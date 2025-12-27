package utils

import (
	"fmt"
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "secret123"
	
	// Test hashing
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}
	
	if hash == "" {
		t.Error("Hash should not be empty")
	}
	
	if hash == password {
		t.Error("Hash should not equal plain password")
	}
	
	fmt.Printf("Original password: %s\n", password)
	fmt.Printf("Hashed password: %s\n", hash)
}

func TestCheckPasswordHash(t *testing.T) {
	password := "secret123"
	wrongPassword := "wrong123"
	
	// Hash the password
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}
	
	// Test correct password
	if !CheckPasswordHash(password, hash) {
		t.Error("CheckPasswordHash should return true for correct password")
	}
	
	// Test wrong password
	if CheckPasswordHash(wrongPassword, hash) {
		t.Error("CheckPasswordHash should return false for wrong password")
	}
	
	fmt.Printf("\nPassword verification test:\n")
	fmt.Printf("Correct password '%s': %v\n", password, CheckPasswordHash(password, hash))
	fmt.Printf("Wrong password '%s': %v\n", wrongPassword, CheckPasswordHash(wrongPassword, hash))
}

func TestHashPasswordDifferentHashes(t *testing.T) {
	password := "secret123"
	
	// Hash the same password twice
	hash1, _ := HashPassword(password)
	hash2, _ := HashPassword(password)
	
	// Hashes should be different due to salt
	if hash1 == hash2 {
		t.Error("Two hashes of the same password should be different (due to salt)")
	}
	
	// But both should verify correctly
	if !CheckPasswordHash(password, hash1) {
		t.Error("First hash should verify correctly")
	}
	
	if !CheckPasswordHash(password, hash2) {
		t.Error("Second hash should verify correctly")
	}
	
	fmt.Printf("\nSalt test - same password hashed twice:\n")
	fmt.Printf("Hash 1: %s\n", hash1)
	fmt.Printf("Hash 2: %s\n", hash2)
	fmt.Printf("Both verify correctly: %v\n", CheckPasswordHash(password, hash1) && CheckPasswordHash(password, hash2))
}

// Example function demonstrating usage
func ExampleHashPassword() {
	hash, _ := HashPassword("secret123")
	fmt.Println(CheckPasswordHash("secret123", hash)) // true
	// Output: true
}
