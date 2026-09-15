package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DisciplineControl struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	RoomID        uuid.UUID      `gorm:"type:uuid;not null" json:"room_id"`
	ResidentID    *uuid.UUID     `gorm:"type:uuid" json:"resident_id,omitempty"`
	StartDate     string         `gorm:"column:start_date;size:10;not null" json:"start_date"`
	EndDate       string         `gorm:"column:end_date;size:10;not null" json:"end_date"`
	Violation     string         `gorm:"type:text;not null" json:"violation"`
	ResponsibleID uuid.UUID      `gorm:"type:uuid;not null" json:"responsible_id"`
	Status        string         `gorm:"size:20;not null;default:'active'" json:"status"`
	Result        string         `gorm:"type:text" json:"result,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Room        Room     `gorm:"foreignKey:RoomID" json:"room,omitempty"`
	Resident    Resident `gorm:"foreignKey:ResidentID" json:"resident,omitempty"`
	Responsible Employee `gorm:"foreignKey:ResponsibleID" json:"responsible,omitempty"`
}

func (DisciplineControl) TableName() string { return "discipline_control" }
