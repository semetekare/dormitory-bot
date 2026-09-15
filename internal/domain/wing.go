package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Wing struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	FloorID   uuid.UUID      `gorm:"type:uuid;not null" json:"floor_id"`
	Name      string         `gorm:"size:100;not null" json:"name"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Floor Floor `gorm:"foreignKey:FloorID" json:"floor,omitempty"`
}

func (Wing) TableName() string { return "wing" }
