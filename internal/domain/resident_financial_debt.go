package domain

import (
	"time"

	"github.com/google/uuid"
)

type ResidentFinancialDebt struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ResidentID  uuid.UUID `gorm:"type:uuid;not null" json:"resident_id"`
	Amount      float64   `gorm:"type:decimal(10,2);not null" json:"amount"`
	Period      string    `gorm:"size:50;not null" json:"period"`
	Description string    `gorm:"type:text" json:"description,omitempty"`
	Source      string    `gorm:"size:100" json:"source,omitempty"`
	UpdatedAt   time.Time `json:"updated_at"`

	Resident Resident `gorm:"foreignKey:ResidentID" json:"resident,omitempty"`
}

func (ResidentFinancialDebt) TableName() string { return "resident_financial_debt" }
