package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MaterialResponsibility struct {
	ID                   uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	RoomID               uuid.UUID      `gorm:"type:uuid;not null" json:"room_id"`
	ItemName             string         `gorm:"size:300;not null" json:"item_name"`
	InventoryNumber      string         `gorm:"size:100;not null" json:"inventory_number"`
	Quantity             int            `gorm:"not null;default:1" json:"quantity"`
	Condition            string         `gorm:"size:100;not null;default:'хорошее'" json:"condition"`
	AcceptedByResidentID *uuid.UUID     `gorm:"type:uuid" json:"accepted_by_resident_id,omitempty"`
	AcceptedAt           *string        `gorm:"column:accepted_at;size:10" json:"accepted_at,omitempty"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	DeletedAt            gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Room               Room     `gorm:"foreignKey:RoomID" json:"room,omitempty"`
	AcceptedByResident Resident `gorm:"foreignKey:AcceptedByResidentID" json:"accepted_by_resident,omitempty"`
}

func (MaterialResponsibility) TableName() string { return "material_responsibility" }
