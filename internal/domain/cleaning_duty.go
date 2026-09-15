package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CleaningDuty struct {
	ID                uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	DormitoryID       uuid.UUID      `gorm:"type:uuid;not null" json:"dormitory_id"`
	FloorID           uuid.UUID      `gorm:"type:uuid;not null" json:"floor_id"`
	RoomID            uuid.UUID      `gorm:"type:uuid;not null" json:"room_id"`
	DutyDate          string         `gorm:"column:duty_date;size:10;not null" json:"duty_date"`
	DutyType          string         `gorm:"size:20;not null" json:"duty_type"`
	Status            string         `gorm:"size:20;not null;default:'scheduled'" json:"status"`
	CompletedByID     *uuid.UUID     `gorm:"type:uuid" json:"completed_by_id,omitempty"`
	CompletedAt       *time.Time     `json:"completed_at,omitempty"`
	PenaltyCleaningID *uuid.UUID     `gorm:"type:uuid" json:"penalty_cleaning_id,omitempty"`
	Notes             string         `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Dormitory       Dormitory        `gorm:"foreignKey:DormitoryID" json:"dormitory,omitempty"`
	Floor           Floor            `gorm:"foreignKey:FloorID" json:"floor,omitempty"`
	Room            Room             `gorm:"foreignKey:RoomID" json:"room,omitempty"`
	CompletedBy     Resident         `gorm:"foreignKey:CompletedByID" json:"completed_by,omitempty"`
	PenaltyCleaning *PenaltyCleaning `gorm:"foreignKey:PenaltyCleaningID" json:"penalty_cleaning,omitempty"`
}

func (CleaningDuty) TableName() string { return "cleaning_duty" }
