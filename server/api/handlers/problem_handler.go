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

func getBestUserStatus(userID uuid.UUID, problemID int) string {
	// Find all submissions for this user and problem
	var subs []models.Submission
	storage.DB.Where("user_id = ? AND problem_id = ?", userID, problemID).Find(&subs)

	if len(subs) == 0 {
		return "UNATTEMPTED"
	}

	bestStatus := "WA" // Default to WA if attempted
	for _, s := range subs {
		if s.Status == models.StatusAC {
			return "AC" // AC overrides everything
		}
		if s.Status == models.StatusPending || s.Status == models.StatusRunning {
			bestStatus = string(s.Status)
		}
	}
	return bestStatus
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
