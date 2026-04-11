package handlers

import (
	"encoding/json"
	"net/http"
	"pt_lpoj/middleware"
	"pt_lpoj/models"
	"pt_lpoj/storage"
	"strconv"

	"github.com/google/uuid"
)

// ProblemResponse extends the base problem with user-specific status
type ProblemResponse struct {
	models.Problem
	UserStatus string `json:"user_status"` // "AC", "WA", "PENDING", or "UNATTEMPTED"
}

// ProblemDetailResponse includes the scaffolding and markdown
type ProblemDetailResponse struct {
	ProblemResponse
	Markdown string `json:"markdown"`
	Scaffold string `json:"scaffold"`
}

// getBestUserStatus returns the best submission status for a user on a specific problem.
// It uses a single optimized query instead of loading all submissions.
func getBestUserStatus(userID uuid.UUID, problemID int) string {
	var acCount int64
	storage.DB.Model(&models.Submission{}).
		Where("user_id = ? AND problem_id = ? AND status = ?", userID, problemID, models.StatusAC).
		Count(&acCount)
	if acCount > 0 {
		return "AC"
	}

	var pendingCount int64
	storage.DB.Model(&models.Submission{}).
		Where("user_id = ? AND problem_id = ? AND (status = ? OR status = ?)", userID, problemID, models.StatusPending, models.StatusRunning).
		Count(&pendingCount)
	if pendingCount > 0 {
		return "PENDING"
	}

	var attemptCount int64
	storage.DB.Model(&models.Submission{}).
		Where("user_id = ? AND problem_id = ?", userID, problemID).
		Count(&attemptCount)
	if attemptCount > 0 {
		return "WA"
	}

	return "UNATTEMPTED"
}

// GetProblemsHandler handles GET /api/problems
func GetProblemsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value(middleware.UserContextKey).(uuid.UUID)
	allProblems := storage.GetAllProblems()

	var res []ProblemResponse
	for _, p := range allProblems {
		status := getBestUserStatus(userID, p.ID)
		res = append(res, ProblemResponse{
			Problem:    p,
			UserStatus: status,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// GetProblemDetailHandler handles GET /api/problems/{id}
func GetProblemDetailHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Basic route parsing without external router library
	// Path should be /api/problems/1001
	pathPrefix := "/api/problems/"
	idStr := r.URL.Path[len(pathPrefix):]

	problemID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid problem ID", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middleware.UserContextKey).(uuid.UUID)

	problem, err := storage.GetProblemByID(problemID)
	if err != nil {
		http.Error(w, "Problem not found", http.StatusNotFound)
		return
	}

	md, err := storage.GetProblemFile(problemID, "problem.md")
	if err != nil {
		http.Error(w, "Markdown missing", http.StatusInternalServerError)
		return
	}

	scaffold, err := storage.GetProblemFile(problemID, "scaffold.py")
	if err != nil {
		http.Error(w, "Scaffold missing", http.StatusInternalServerError)
		return
	}

	status := getBestUserStatus(userID, problemID)

	res := ProblemDetailResponse{
		ProblemResponse: ProblemResponse{
			Problem:    *problem,
			UserStatus: status,
		},
		Markdown: md,
		Scaffold: scaffold,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
