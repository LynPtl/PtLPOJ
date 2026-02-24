package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SubmissionStatus defines the current state of a code evaluation
type SubmissionStatus string

const (
	StatusPending SubmissionStatus = "PENDING"
	StatusRunning SubmissionStatus = "RUNNING"
	StatusAC      SubmissionStatus = "AC"  // Accepted
	StatusWA      SubmissionStatus = "WA"  // Wrong Answer
	StatusTLE     SubmissionStatus = "TLE" // Time Limit Exceeded
	StatusOLE     SubmissionStatus = "OLE" // Output Limit Exceeded
	StatusRE      SubmissionStatus = "RE"  // Runtime Error
	StatusCE      SubmissionStatus = "CE"  // Compile Error (System error in our case)
)

// Submission represents a code submission made by a user for a specific problem
type Submission struct {
	ID              uuid.UUID        `gorm:"type:uuid;primaryKey"`
	UserID          uuid.UUID        `gorm:"type:uuid;index;not null"`
	ProblemID       int              `gorm:"index;not null"`
	Code            string           `gorm:"type:text;not null"`
	Status          SubmissionStatus `gorm:"type:string;default:'PENDING';index"`
	ExecutionTimeMs int              `gorm:"default:0"` // Runtime execution time in milliseconds
	MemoryPeakKb    int              `gorm:"default:0"` // Peak memory used by Docker container in KB
	FailedAtCase    int              `gorm:"default:0"` // The case number that caused WA/RE. 0 if AC
	Message         string           `gorm:"type:text"`
	CreatedAt       time.Time        `gorm:"index"`
	UpdatedAt       time.Time
}

// BeforeCreate will set a UUID rather than numeric ID.
func (s *Submission) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return
}
