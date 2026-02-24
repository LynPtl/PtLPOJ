package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"pt_lpoj/models"
	"pt_lpoj/storage"
	"testing"
)

func initTestDB(t *testing.T) {
	err := storage.InitDB("file::memory:")
	if err != nil {
		t.Fatalf("Failed to init db: %v", err)
	}

	// Insert test user into whitelist
	_, err = storage.CreateUser("ptlantern@gmail.com", models.RoleAdmin)
	if err != nil {
		t.Fatalf("Failed to seed user: %v", err)
	}
}

func TestRequestOTPHandler_Whitelist(t *testing.T) {
	initTestDB(t)

	// Subcase 1: Reject unregistered user (silent rejection though HTTP 200 to prevent enumeration)
	reqBody := `{"email": "hacker@evil.com"}`
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBufferString(reqBody))
	rr := httptest.NewRecorder()

	RequestOTPHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200 OK (silent rejection), got %d", rr.Code)
	}
	if !bytes.Contains(rr.Body.Bytes(), []byte("If you are registered")) {
		t.Errorf("Expected silent rejection message, got %s", rr.Body.String())
	}

	// Subcase 2: Allow registered user
	validReq := `{"email": "ptlantern@gmail.com"}`
	req2 := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBufferString(validReq))
	rr2 := httptest.NewRecorder()

	RequestOTPHandler(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", rr2.Code)
	}
	if !bytes.Contains(rr2.Body.Bytes(), []byte("OTP sent successfully")) {
		t.Errorf("Expected success message, got %s", rr2.Body.String())
	}
}
