package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Employee struct {
	ID                 uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID             uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_employee_user_id" json:"user_id"`
	Position           string         `gorm:"size:200;not null" json:"position"`
	Department         string         `gorm:"size:200" json:"department,omitempty"`
	PrimaryDormitoryID *uuid.UUID     `gorm:"type:uuid" json:"primary_dormitory_id,omitempty"`
	Role               string         `gorm:"size:50;not null" json:"role"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	User             User                    `gorm:"foreignKey:UserID" json:"user,omitempty"`
	PrimaryDormitory Dormitory               `gorm:"foreignKey:PrimaryDormitoryID" json:"primary_dormitory,omitempty"`
	DormitoryRoles   []EmployeeDormitoryRole `gorm:"foreignKey:EmployeeID" json:"dormitory_roles,omitempty"`
}

func (Employee) TableName() string { return "employee" }
