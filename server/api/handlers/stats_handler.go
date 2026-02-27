package handlers

import (
	"encoding/json"
	"net/http"
	"pt_lpoj/middleware"
	"pt_lpoj/models"
	"pt_lpoj/storage"

	"github.com/google/uuid"
)

type UserStatsResponse struct {
	TotalSubmissions     int64               `json:"total_submissions"`
	ACCount              int64               `json:"ac_count"`
	UniqueProblemsSolved int64               `json:"unique_problems_solved"`
	RecentSubmissions    []models.Submission `json:"recent_submissions"`
}

// GetUserStatsHandler handles GET /api/user/stats
func GetUserStatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value(middleware.UserContextKey).(uuid.UUID)

	var stats UserStatsResponse

	// 1. Total Submissions
	storage.DB.Model(&models.Submission{}).Where("user_id = ?", userID).Count(&stats.TotalSubmissions)

	// 2. AC Count
	storage.DB.Model(&models.Submission{}).Where("user_id = ? AND status = ?", userID, models.StatusAC).Count(&stats.ACCount)

	// 3. Unique Problems Solved
	storage.DB.Model(&models.Submission{}).
		Where("user_id = ? AND status = ?", userID, models.StatusAC).
		Distinct("problem_id").
		Count(&stats.UniqueProblemsSolved)

	// 4. Recent Submissions (Last 5)
	storage.DB.Where("user_id = ?", userID).
		Order("created_at desc").
		Limit(5).
		Find(&stats.RecentSubmissions)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
