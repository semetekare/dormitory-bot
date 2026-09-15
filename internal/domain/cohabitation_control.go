package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CohabitationControl struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	RoomID    uuid.UUID      `gorm:"type:uuid;not null" json:"room_id"`
	StartDate string         `gorm:"column:start_date;size:10;not null" json:"start_date"`
	EndDate   string         `gorm:"column:end_date;size:10;not null" json:"end_date"`
	Reason    string         `gorm:"type:text" json:"reason,omitempty"`
	Status    string         `gorm:"size:20;not null;default:'active'" json:"status"`
	Result    string         `gorm:"type:text" json:"result,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Room Room `gorm:"foreignKey:RoomID" json:"room,omitempty"`
}

func (CohabitationControl) TableName() string { return "cohabitation_control" }
