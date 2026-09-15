package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ChatLink struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	DormitoryID uuid.UUID      `gorm:"type:uuid;not null" json:"dormitory_id"`
	FloorID     *uuid.UUID     `gorm:"type:uuid" json:"floor_id,omitempty"`
	WingID      *uuid.UUID     `gorm:"type:uuid" json:"wing_id,omitempty"`
	LinkType    string         `gorm:"size:20;not null" json:"link_type"`
	Platform    string         `gorm:"size:20;not null" json:"platform"`
	URL         string         `gorm:"type:text;not null" json:"url"`
	Title       string         `gorm:"size:200;not null" json:"title"`
	IsActive    bool           `gorm:"not null;default:true" json:"is_active"`
	CreatedBy   uuid.UUID      `gorm:"type:uuid;not null" json:"created_by"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Dormitory Dormitory `gorm:"foreignKey:DormitoryID" json:"dormitory,omitempty"`
	Floor     Floor     `gorm:"foreignKey:FloorID" json:"floor,omitempty"`
	Wing      Wing      `gorm:"foreignKey:WingID" json:"wing,omitempty"`
	Creator   Employee  `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

func (ChatLink) TableName() string { return "chat_link" }
