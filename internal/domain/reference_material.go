package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReferenceMaterial struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	DormitoryID uuid.UUID      `gorm:"type:uuid;not null" json:"dormitory_id"`
	Name        string         `gorm:"size:300;not null" json:"name"`
	Description string         `gorm:"type:text;not null" json:"description"`
	Category    string         `gorm:"size:100;not null" json:"category"`
	ModuleKey   *string        `gorm:"size:50;index" json:"module_key,omitempty"`
	Ordinal     int            `gorm:"not null;default:0" json:"ordinal"`
	DocsURL     string         `gorm:"type:text" json:"docs_url,omitempty"`
	CreatedBy   uuid.UUID      `gorm:"type:uuid;not null" json:"created_by"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Dormitory Dormitory `gorm:"foreignKey:DormitoryID" json:"dormitory,omitempty"`
	Creator   Employee  `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

func (ReferenceMaterial) TableName() string { return "reference_material" }

const (
	ModuleLaundry    = "laundry"
	ModuleRoom       = "my_room"
	ModuleCleaning   = "cleaning"
	ModuleChatLinks  = "chat_links"
	ModuleReferences = "references"
)
