package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Dormitory struct {
	ID               uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name             string         `gorm:"size:200;not null" json:"name"`
	Address          string         `gorm:"type:text" json:"address,omitempty"`
	EISDormitoryCode string         `gorm:"size:50" json:"eis_dormitory_code,omitempty"`
	IsActive         bool           `gorm:"not null;default:true" json:"is_active"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (Dormitory) TableName() string { return "dormitory" }
