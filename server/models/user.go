package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Role defines the access level of a user
type Role string

const (
	RoleAdmin   Role = "ADMIN"
	RoleStudent Role = "STUDENT"
)

// User represents an authenticated student or admin in the system
type User struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email     string    `gorm:"uniqueIndex;not null"`
	Role      Role      `gorm:"type:string;default:'STUDENT'"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return
}
