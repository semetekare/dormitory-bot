package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WashingMachine struct {
	ID                uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	DormitoryID       uuid.UUID      `gorm:"type:uuid;not null" json:"dormitory_id"`
	FloorID           uuid.UUID      `gorm:"type:uuid;not null" json:"floor_id"`
	MachineNumber     string         `gorm:"size:10;not null" json:"machine_number"`
	IsActive          bool           `gorm:"not null;default:true" json:"is_active"`
	DurationMinutes   int            `gorm:"not null;default:60" json:"duration_minutes"`
	BlockedSlotStart  string         `gorm:"column:blocked_slot_start;size:10" json:"blocked_slot_start,omitempty"`
	BlockedSlotEnd    string         `gorm:"column:blocked_slot_end;size:10" json:"blocked_slot_end,omitempty"`
	BlockedSlotReason string         `gorm:"column:blocked_slot_reason;size:200" json:"blocked_slot_reason,omitempty"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Dormitory Dormitory `gorm:"foreignKey:DormitoryID" json:"dormitory,omitempty"`
	Floor     Floor     `gorm:"foreignKey:FloorID" json:"floor,omitempty"`
}

func (WashingMachine) TableName() string { return "washing_machine" }
