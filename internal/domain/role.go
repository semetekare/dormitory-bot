package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Role struct {
	ID          uuid.UUID     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name        string        `gorm:"size:100;uniqueIndex;not null" json:"name"`
	DisplayName string        `gorm:"size:200;not null" json:"display_name"`
	Priority    int           `gorm:"default:0" json:"priority"`
	IsSystem    bool          `gorm:"default:false" json:"is_system"`
	TargetType  string        `gorm:"size:20;default:'employee'" json:"target_type"`
	JSONAccess  JSONAccessMap `gorm:"type:jsonb;not null;default:'{}'" json:"json_access"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

func (Role) TableName() string { return "role" }

type JSONAccessMap map[string]ResourcePermissions

func (j JSONAccessMap) Value() (driver.Value, error) {
	if j == nil {
		return "{}", nil
	}
	return json.Marshal(j)
}

func (j *JSONAccessMap) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONAccessMap)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to unmarshal JSONB value: not bytes")
	}
	return json.Unmarshal(bytes, j)
}

func (j JSONAccessMap) GetResourcePermissions(resourceType string) (ResourcePermissions, bool) {
	key := "resource." + resourceType
	perms, ok := j[key]
	if !ok {
		perms, ok = j[resourceType]
		return perms, ok
	}
	return perms, ok
}

type ResourcePermissions map[string]interface{}

func (r *Role) HasOperation(resourceType, operation string) ([]PermissionContext, bool) {
	perms, ok := r.JSONAccess.GetResourcePermissions(resourceType)
	if !ok {
		return nil, false
	}
	if wildcardValue, exists := perms["*"]; exists {
		return ParsePermissionValue(wildcardValue)
	}
	if opValue, exists := perms[operation]; exists {
		return ParsePermissionValue(opValue)
	}
	return nil, false
}

func (r *Role) CanAssignTo(personType string) bool {
	if r.TargetType == TargetTypeBoth {
		return true
	}
	if personType == "student" {
		return r.TargetType == TargetTypeResident
	}
	return r.TargetType == personType
}
