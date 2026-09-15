package domain

import (
	"time"

	"github.com/google/uuid"
)

type LaundrySettings struct {
	ID                     uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	DormitoryID            uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_ls_dormitory_id" json:"dormitory_id"`
	BookingStartTime       string    `gorm:"column:booking_start_time;size:10;not null;default:'07:00:00'" json:"booking_start_time"`
	BookingEndTime         string    `gorm:"column:booking_end_time;size:10;not null;default:'23:00:00'" json:"booking_end_time"`
	AdvanceBookingEnabled  bool      `gorm:"not null;default:false" json:"advance_booking_enabled"`
	AdvanceBookingMaxDays  int       `gorm:"not null;default:1" json:"advance_booking_max_days"`
	DefaultDurationMinutes int       `gorm:"not null;default:60" json:"default_duration_minutes"`
	BlockedSlotStart       string    `gorm:"column:blocked_slot_start;size:10" json:"blocked_slot_start,omitempty"`
	BlockedSlotEnd         string    `gorm:"column:blocked_slot_end;size:10" json:"blocked_slot_end,omitempty"`
	BlockedSlotDays        string    `gorm:"size:50" json:"blocked_slot_days,omitempty"`
	BlockedSlotReason      string    `gorm:"size:200" json:"blocked_slot_reason,omitempty"`
	CastellanBookingEnabled bool     `gorm:"not null;default:true" json:"castellan_booking_enabled"`
	UpdatedAt              time.Time `json:"updated_at"`

	Dormitory Dormitory `gorm:"foreignKey:DormitoryID" json:"dormitory,omitempty"`
}

func (LaundrySettings) TableName() string { return "laundry_settings" }
