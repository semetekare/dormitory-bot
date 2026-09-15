package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PenaltyCleaning struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	RoomID         uuid.UUID      `gorm:"type:uuid;not null" json:"room_id"`
	ResidentID     *uuid.UUID     `gorm:"type:uuid" json:"resident_id,omitempty"`
	Count          int            `gorm:"not null;default:1" json:"count"`
	CompletedCount int            `gorm:"not null;default:0" json:"completed_count"`
	Reason         string         `gorm:"type:text;not null" json:"reason"`
	IssuedByID     uuid.UUID      `gorm:"type:uuid;not null" json:"issued_by_id"`
	IssuedAt       string         `gorm:"column:issued_at;size:10;not null" json:"issued_at"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Room     Room     `gorm:"foreignKey:RoomID" json:"room,omitempty"`
	Resident Resident `gorm:"foreignKey:ResidentID" json:"resident,omitempty"`
	IssuedBy Employee `gorm:"foreignKey:IssuedByID" json:"issued_by,omitempty"`
}

func (PenaltyCleaning) TableName() string { return "penalty_cleaning" }
