package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserDormitoryRole struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;index:idx_udr_user_dormitory" json:"user_id"`
	DormitoryID uuid.UUID `gorm:"type:uuid;not null;index:idx_udr_user_dormitory" json:"dormitory_id"`
	Role        string    `gorm:"size:100;not null" json:"role"`
	AssignedBy  uuid.UUID `gorm:"type:uuid" json:"assigned_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	User         User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Dormitory    Dormitory `gorm:"foreignKey:DormitoryID" json:"dormitory,omitempty"`
	AssignedUser Employee  `gorm:"foreignKey:AssignedBy" json:"assigned_user,omitempty"`
}

func (UserDormitoryRole) TableName() string {
	return "user_dormitory_role"
}

type UserDormitoryRoleWithUser struct {
	UserDormitoryRole
	UserFirstName  string `json:"user_first_name"`
	UserLastName   string `json:"user_last_name"`
	UserMiddleName string `json:"user_middle_name,omitempty"`
	UserPhone      string `json:"user_phone"`
	PersonType     string `json:"person_type"`
	Source         string `json:"source"`
	PlatformUserID string `json:"platform_user_id"`
}
