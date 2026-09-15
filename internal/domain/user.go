package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	FirstName      string         `gorm:"size:100;not null" json:"first_name"`
	LastName       string         `gorm:"size:100;not null" json:"last_name"`
	MiddleName     string         `gorm:"size:100" json:"middle_name,omitempty"`
	Phone          string         `gorm:"size:20;not null;uniqueIndex:idx_user_phone" json:"phone"`
	Platform       string         `gorm:"size:20;not null" json:"platform"`
	PlatformUserID string         `gorm:"size:255;not null;uniqueIndex:idx_user_platform_user_id" json:"platform_user_id"`
	EISVerified    bool           `gorm:"not null;default:false" json:"eis_verified"`
	EISPersonID    string         `gorm:"size:50" json:"eis_person_id,omitempty"`
	PersonType     string         `gorm:"size:20" json:"person_type,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (User) TableName() string { return "user" }
