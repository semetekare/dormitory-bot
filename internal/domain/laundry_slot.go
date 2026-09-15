package domain

import (
	"github.com/google/uuid"
)

type LaundrySlot struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	MachineID   uuid.UUID `gorm:"type:uuid;not null;index" json:"machine_id"`
	SlotStart   string    `gorm:"size:10;not null" json:"slot_start"` // "07:00"
	SlotEnd     string    `gorm:"size:10;not null" json:"slot_end"`   // "07:40"
	DurationMin int       `gorm:"not null;default:40" json:"duration_min"`
	DaysOfWeek  string    `gorm:"size:20" json:"days_of_week,omitempty"` // "1,2,3,4,5" or "" = all
	IsActive    bool      `gorm:"not null;default:true" json:"is_active"`

	Machine WashingMachine `gorm:"foreignKey:MachineID" json:"machine,omitempty"`
}

func (LaundrySlot) TableName() string { return "laundry_slots" }
