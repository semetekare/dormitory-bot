package domain

import (
	"github.com/google/uuid"
)

type CleaningSettings struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	DormitoryID    uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_cs_dormitory_id" json:"dormitory_id"`
	DutyStartTime  string    `gorm:"size:10;not null;default:'18:00'" json:"duty_start_time"`
	DutyDuration   int       `gorm:"not null;default:60" json:"duty_duration"`
	ReminderBefore int       `gorm:"not null;default:60" json:"reminder_before"`
	ActiveDays     string    `gorm:"size:20;not null;default:'1,2,3,4,5,6,7'" json:"active_days"`

	Dormitory Dormitory `gorm:"foreignKey:DormitoryID" json:"dormitory,omitempty"`
}

func (CleaningSettings) TableName() string { return "cleaning_settings" }
