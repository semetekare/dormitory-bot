package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Resident struct {
	ID                uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID            uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_resident_user_id" json:"user_id"`
	DormitoryID       uuid.UUID      `gorm:"type:uuid;not null" json:"dormitory_id"`
	RoomID            uuid.UUID      `gorm:"type:uuid;not null" json:"room_id"`
	CheckInDate       *string        `gorm:"column:check_in_date;size:10" json:"check_in_date,omitempty"`
	IsActive          bool           `gorm:"not null;default:true;index" json:"is_active"`
	ContractNumber    *string        `gorm:"size:50" json:"contract_number,omitempty"`
	ContractStartDate *string        `gorm:"size:10" json:"contract_start_date,omitempty"`
	ContractEndDate   *string        `gorm:"size:10" json:"contract_end_date,omitempty"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	User      User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Dormitory Dormitory `gorm:"foreignKey:DormitoryID" json:"dormitory,omitempty"`
	Room      Room      `gorm:"foreignKey:RoomID" json:"room,omitempty"`
}

func (Resident) TableName() string { return "resident" }
