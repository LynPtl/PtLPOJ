package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestOTPGenerationAndVerification(t *testing.T) {
	email := "ptlantern@gmail.com"

	// 1. Generate OTP
	code, err := GenerateOTP(email)
	if err != nil {
		t.Fatalf("Failed to generate OTP: %v", err)
	}
	if len(code) != 6 {
		t.Errorf("Expected 6-digit OTP, got %s", code)
	}

	// 2. Fail invalid code
	if VerifyOTP(email, "000000") {
		t.Errorf("VerifyOTP should fail for incorrect code")
	}

	// 3. Success for correct code
	if !VerifyOTP(email, code) {
		t.Errorf("VerifyOTP should pass for correct code")
	}

	// 4. Fail replay attack (should be consumed)
	if VerifyOTP(email, code) {
		t.Errorf("VerifyOTP should fail on second attempt (already consumed)")
	}
}

func TestOTPExpiration(t *testing.T) {
	email := "expire@test.com"
	// Temporarily override TTL for fast test
	oldTTL := OTPTTL
	OTPTTL = 1 * time.Second
	defer func() { OTPTTL = oldTTL }()

	code, _ := GenerateOTP(email)

	// Wait for expiration
	time.Sleep(1500 * time.Millisecond)

	if VerifyOTP(email, code) {
		t.Errorf("VerifyOTP should fail after expiration")
	}
}

func TestJWTGenerationAndValidation(t *testing.T) {
	userID := uuid.New()

	// 1. Generate Token
	token, err := GenerateJWT(userID)
	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}

	// 2. Validate Token
	parsedID, err := ValidateJWT(token)
	if err != nil {
		t.Fatalf("Failed to validate JWT: %v", err)
	}

	if parsedID != userID {
		t.Errorf("Expected UUID %s, got %s", userID, parsedID)
	}

	// 3. Reject tampering
	tamperedToken := token + "junk"
	_, err = ValidateJWT(tamperedToken)
	if err == nil {
		t.Errorf("Expected tampered token to fail validation")
	}
}
