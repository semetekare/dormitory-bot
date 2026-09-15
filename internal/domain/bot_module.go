package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BotModule struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	DormitoryID   uuid.UUID      `gorm:"type:uuid;not null" json:"dormitory_id"`
	Platform      string         `gorm:"size:20;not null" json:"platform"`
	BotToken      string         `gorm:"size:500;not null" json:"bot_token"`
	BotName       string         `gorm:"size:200" json:"bot_name,omitempty"`
	Status        string         `gorm:"size:20;not null;default:'stopped'" json:"status"`
	LastStartedAt *time.Time     `json:"last_started_at,omitempty"`
	LastStoppedAt *time.Time     `json:"last_stopped_at,omitempty"`
	ErrorMessage  string         `gorm:"type:text" json:"error_message,omitempty"`
	StatsJSON     JSONStatsMap   `gorm:"type:jsonb" json:"stats_json,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Dormitory Dormitory `gorm:"foreignKey:DormitoryID" json:"dormitory,omitempty"`
}

func (BotModule) TableName() string { return "bot_module" }

type JSONStatsMap map[string]interface{}

func (j JSONStatsMap) Value() (driver.Value, error) {
	if j == nil {
		return "{}", nil
	}
	return json.Marshal(j)
}

func (j *JSONStatsMap) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONStatsMap)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to unmarshal JSONB value: not bytes")
	}
	return json.Unmarshal(bytes, j)
}
