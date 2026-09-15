package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CleaningExemption struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	RoomID      uuid.UUID      `gorm:"type:uuid;not null" json:"room_id"`
	StartDate   string         `gorm:"column:start_date;size:10;not null" json:"start_date"`
	EndDate     string         `gorm:"column:end_date;size:10;not null" json:"end_date"`
	Reason      string         `gorm:"type:text;not null" json:"reason"`
	CreatedByID uuid.UUID      `gorm:"type:uuid;not null" json:"created_by_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Room      Room     `gorm:"foreignKey:RoomID" json:"room,omitempty"`
	CreatedBy Employee `gorm:"foreignKey:CreatedByID" json:"created_by,omitempty"`
}

func (CleaningExemption) TableName() string { return "cleaning_exemption" }
