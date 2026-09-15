package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Floor struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	DormitoryID uuid.UUID      `gorm:"type:uuid;not null" json:"dormitory_id"`
	FloorNumber int            `gorm:"not null" json:"floor_number"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Dormitory Dormitory `gorm:"foreignKey:DormitoryID" json:"dormitory,omitempty"`
}

func (Floor) TableName() string { return "floor" }
