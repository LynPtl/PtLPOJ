package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"pt_lpoj/middleware"
	"pt_lpoj/models"
	"pt_lpoj/storage"
	"strings"
	"time"

	"github.com/google/uuid"
)

type SubmitRequest struct {
	ProblemID int    `json:"problem_id"`
	Code      string `json:"source_code"`
}

type SubmitResponse struct {
	SubmissionID string `json:"submission_id"`
	Message      string `json:"message"`
}

// CreateSubmissionHandler handles POST /api/submissions
func CreateSubmissionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ProblemID <= 0 || strings.TrimSpace(req.Code) == "" {
		http.Error(w, "Problem ID and Code are required", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middleware.UserContextKey).(uuid.UUID)

	sub, err := storage.CreateSubmission(userID, req.ProblemID, req.Code)
	if err != nil {
		http.Error(w, "Failed to create submission", http.StatusInternalServerError)
		return
	}

	res := SubmitResponse{
		SubmissionID: sub.ID.String(),
		Message:      "Submission queued",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted) // 202 Accepted
	json.NewEncoder(w).Encode(res)
}

// GetUserSubmissionsHandler handles GET /api/submissions
func GetUserSubmissionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value(middleware.UserContextKey).(uuid.UUID)

	subs, err := storage.GetUserSubmissions(userID)
	if err != nil {
		http.Error(w, "Failed to retrieve submissions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subs)
}

// SSEStreamHandler handles GET /api/submissions/{id}/stream
// Pushes real-time SSE updates to the client
func SSEStreamHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Basic route parsing
	pathPrefix := "/api/submissions/"
	pathSuffix := "/stream"

	// e.g. /api/submissions/{id}/stream
	// len("/api/submissions/") = 17, len("/stream") = 7
	// length checking
	if len(r.URL.Path) <= len(pathPrefix)+len(pathSuffix) {
		http.Error(w, "Invalid submission ID pattern", http.StatusBadRequest)
		return
	}

	idStr := r.URL.Path[len(pathPrefix) : len(r.URL.Path)-len(pathSuffix)]
	subID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid submission UUID", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middleware.UserContextKey).(uuid.UUID)

	// Validate Ownership
	sub, err := storage.GetSubmissionByID(subID)
	if err != nil || sub == nil {
		http.Error(w, "Submission not found", http.StatusNotFound)
		return
	}
	if sub.UserID != userID {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	// Make sure we allow CORS for local dev if web app was ever used, but VSCode doesn't strictly need it unless browser
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// We poll the database for updates to push down.
	// Normally we'd use Channels/Redis PubSub, but for SQLite WAL + minimal deployment
	// polling is very clean and fully persistent across app restarts.
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	// Timeout to prevent infinite lingering connection (e.g. 1 minute limit)
	timeout := time.After(60 * time.Second)

	for {
		select {
		case <-r.Context().Done():
			// Client disconnected
			return
		case <-timeout:
			// Safety cutoff
			fmt.Fprintf(w, "event: error\ndata: %s\n\n", "timeout waiting for result")
			flusher.Flush()
			return
		case <-ticker.C:
			currentSub, err := storage.GetSubmissionByID(subID)
			if err != nil {
				continue
			}

			// Encode and send Data
			data, _ := json.Marshal(currentSub)
			fmt.Fprintf(w, "data: %s\n\n", string(data))
			flusher.Flush()

			// Terminate stream if no longer pending/running
			if currentSub.Status != models.StatusPending && currentSub.Status != models.StatusRunning {
				// Send a closing event
				fmt.Fprintf(w, "event: complete\ndata: {\"finished\": true}\n\n")
				flusher.Flush()
				return
			}
		}
	}
}
