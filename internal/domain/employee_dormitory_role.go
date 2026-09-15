package domain

import (
	"time"

	"github.com/google/uuid"
)

type EmployeeDormitoryRole struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	EmployeeID  uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_edr_employee_dormitory" json:"employee_id"`
	DormitoryID uuid.UUID `gorm:"type:uuid;not null" json:"dormitory_id"`
	Role        string    `gorm:"size:50;not null" json:"role"`
	CreatedAt   time.Time `json:"created_at"`

	Employee  Employee  `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Dormitory Dormitory `gorm:"foreignKey:DormitoryID" json:"dormitory,omitempty"`
}

func (EmployeeDormitoryRole) TableName() string { return "employee_dormitory_role" }
