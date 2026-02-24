package auth

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"time"
)

// OTPEntry holds the OTP code and its expiration time
type OTPEntry struct {
	Code      string
	ExpiresAt time.Time
}

var (
	// otpStore is an in-memory thread-safe map to hold OTPs
	otpStore sync.Map
	// OTPTTL is the time-to-live for an OTP
	OTPTTL = 5 * time.Minute
)

// GenerateOTP creates a 6-digit random string and stores it in memory
func GenerateOTP(email string) (string, error) {
	// Generate a secure 6-digit random number
	maxVal := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, maxVal)
	if err != nil {
		return "", fmt.Errorf("failed to generate random OTP: %w", err)
	}
	code := fmt.Sprintf("%06d", n.Int64())

	// Store it with TTL
	entry := OTPEntry{
		Code:      code,
		ExpiresAt: time.Now().Add(OTPTTL),
	}
	otpStore.Store(email, entry)

	return code, nil
}

// VerifyOTP checks if the provided code matches the stored one for the email,
// ensuring it hasn't expired. It consumes the OTP upon successful verification.
func VerifyOTP(email, code string) bool {
	val, ok := otpStore.Load(email)
	if !ok {
		return false
	}

	entry := val.(OTPEntry)

	// Check expiration
	if time.Now().After(entry.ExpiresAt) {
		// Expired, clean it up
		otpStore.Delete(email)
		return false
	}

	// Check code match
	if entry.Code == code {
		// Consumed successfully, remove it to prevent replay attacks
		otpStore.Delete(email)
		return true
	}

	return false
}

// CleanupExpiredOTPs can be run as a background goroutine to periodically clean up RAM
func CleanupExpiredOTPs() {
	for {
		time.Sleep(1 * time.Minute)
		now := time.Now()
		otpStore.Range(func(key, value interface{}) bool {
			entry := value.(OTPEntry)
			if now.After(entry.ExpiresAt) {
				otpStore.Delete(key)
			}
			return true
		})
	}
}
