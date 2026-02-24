package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"pt_lpoj/auth"
	"pt_lpoj/storage"
)

type LoginRequest struct {
	Email string `json:"email"`
}

type VerifyRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type VerifyResponse struct {
	Token string `json:"token"`
}

// RequestOTPHandler handles POST /api/auth/login
// It validates if the email is in the DB (whitelist), generates an OTP, and sends it.
func RequestOTPHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	email := strings.TrimSpace(req.Email)
	if email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	// 1. Check Whitelist: Does this user exist in our DB?
	user, err := storage.GetUserByEmail(email)
	if err != nil || user == nil {
		// Generic message to prevent email enumeration, though internal system
		http.Error(w, "If you are registered, an OTP will be sent to your email.", http.StatusOK)
		return
	}

	// 2. Generate OTP
	code, err := auth.GenerateOTP(email)
	if err != nil {
		http.Error(w, "Failed to generate OTP", http.StatusInternalServerError)
		return
	}

	// 3. Send Email (Mocked)
	err = auth.SendOTPEmail(email, code)
	if err != nil {
		http.Error(w, "Failed to dispatch email", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "OTP sent successfully"}`))
}

// VerifyOTPHandler handles POST /api/auth/verify
// It exchanges a valid Email+OTP pair for a signed JWT
func VerifyOTPHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	email := strings.TrimSpace(req.Email)
	code := strings.TrimSpace(req.Code)

	if email == "" || code == "" {
		http.Error(w, "Email and Code are required", http.StatusBadRequest)
		return
	}

	// 1. Verify OTP
	isValid := auth.VerifyOTP(email, code)
	if !isValid {
		http.Error(w, "Invalid or Expired OTP", http.StatusUnauthorized)
		return
	}

	// 2. Fetch User to get UUID
	user, err := storage.GetUserByEmail(email)
	if err != nil || user == nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// 3. Generate JWT
	token, err := auth.GenerateJWT(user.ID)
	if err != nil {
		http.Error(w, "Failed to generate session token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(VerifyResponse{Token: token})
}
