package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LaundryBooking struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	MachineID   uuid.UUID      `gorm:"type:uuid;not null" json:"machine_id"`
	UserID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	BookerType  string         `gorm:"size:10;not null;default:'resident'" json:"booker_type"` // "resident" or "employee"
	SlotDate    string         `gorm:"column:slot_date;size:10;not null" json:"slot_date"`
	SlotStart   string         `gorm:"column:slot_start;size:10;not null" json:"slot_start"`
	SlotEnd     string         `gorm:"column:slot_end;size:10;not null" json:"slot_end"`
	Status      string         `gorm:"size:20;not null;default:'active'" json:"status"`
	CancelledBy string         `gorm:"size:20" json:"cancelled_by,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Machine WashingMachine `gorm:"foreignKey:MachineID" json:"machine,omitempty"`
	User    User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (LaundryBooking) TableName() string { return "laundry_booking" }
