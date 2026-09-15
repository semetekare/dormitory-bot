package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Room struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	DormitoryID uuid.UUID      `gorm:"type:uuid;not null" json:"dormitory_id"`
	FloorID     uuid.UUID      `gorm:"type:uuid;not null" json:"floor_id"`
	WingID      *uuid.UUID     `gorm:"type:uuid" json:"wing_id,omitempty"`
	RoomNumber  string         `gorm:"size:20;not null" json:"room_number"`
	Capacity    int            `gorm:"not null;default:1" json:"capacity"`
	EISRoomCode string         `gorm:"size:50" json:"eis_room_code,omitempty"`
	IsActive    bool           `gorm:"not null;default:true" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Dormitory Dormitory `gorm:"foreignKey:DormitoryID" json:"dormitory,omitempty"`
	Floor     Floor     `gorm:"foreignKey:FloorID" json:"floor,omitempty"`
	Wing      Wing      `gorm:"foreignKey:WingID" json:"wing,omitempty"`
}

func (Room) TableName() string { return "room" }
