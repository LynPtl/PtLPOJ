package storage

import (
	"errors"
	"pt_lpoj/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CreateUser inserts a newly authenticated user into the database.
func CreateUser(email string, role models.Role) (*models.User, error) {
	u := &models.User{
		Email: email,
		Role:  role,
	}
	result := DB.Create(u)
	if result.Error != nil {
		return nil, result.Error
	}
	return u, nil
}

// GetUserByEmail finds a user by their email address.
func GetUserByEmail(email string) (*models.User, error) {
	var u models.User
	result := DB.Where("email = ?", email).First(&u)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil, nil if user doesn't exist yet
		}
		return nil, result.Error
	}
	return &u, nil
}

// CreateSubmission records a new user code submission.
func CreateSubmission(userID uuid.UUID, problemID int, code string) (*models.Submission, error) {
	s := &models.Submission{
		UserID:    userID,
		ProblemID: problemID,
		Code:      code,
		Status:    models.StatusPending,
	}
	result := DB.Create(s)
	if result.Error != nil {
		return nil, result.Error
	}
	return s, nil
}

// UpdateSubmissionStatus updates the outcome and performance metrics of a code evaluation.
func UpdateSubmissionStatus(id uuid.UUID, status models.SubmissionStatus, msg string, timeMs int, memKb int, failedCase int) error {
	result := DB.Model(&models.Submission{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":            status,
		"message":           msg,
		"execution_time_ms": timeMs,
		"memory_peak_kb":    memKb,
		"failed_at_case":    failedCase,
	})
	return result.Error
}

// GetUserSubmissions retrieves a user's chronological submission history.
func GetUserSubmissions(userID uuid.UUID) ([]models.Submission, error) {
	var submissions []models.Submission
	result := DB.Where("user_id = ?", userID).Order("created_at desc").Find(&submissions)
	return submissions, result.Error
}

// GetSubmissionByID retrieves a single submission.
func GetSubmissionByID(id uuid.UUID) (*models.Submission, error) {
	var s models.Submission
	result := DB.First(&s, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &s, nil
}
