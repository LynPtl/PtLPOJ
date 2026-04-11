package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"pt_lpoj/middleware"
	"pt_lpoj/scheduler"
	"pt_lpoj/storage"
	"strings"

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

	if scheduler.GlobalQueue != nil {
		scheduler.GlobalQueue.Enqueue(sub.ID)
	}

	res := SubmitResponse{
		SubmissionID: sub.ID.String(),
		Message:      "Submission queued",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
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
func SSEStreamHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	pathPrefix := "/api/submissions/"
	pathSuffix := "/stream"

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
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if scheduler.GlobalQueue == nil {
		http.Error(w, "Judge system not initialized", http.StatusServiceUnavailable)
		return
	}
	resultCh := scheduler.GlobalQueue.Subscribe(subID)
	defer scheduler.GlobalQueue.Unsubscribe(subID)

	timeout := r.Context()

	for {
		select {
		case <-timeout.Done():
			fmt.Fprintf(w, "event: error\ndata: %s\n\n", "timeout waiting for result")
			flusher.Flush()
			return
		case result, ok := <-resultCh:
			if !ok {
				fmt.Fprintf(w, "event: error\ndata: %s\n\n", "subscription closed unexpectedly")
				flusher.Flush()
				return
			}

			data, _ := json.Marshal(result)
			fmt.Fprintf(w, "data: %s\n\n", string(data))
			flusher.Flush()

			fmt.Fprintf(w, "event: complete\n\n")
			flusher.Flush()
			return
		}
	}
}
