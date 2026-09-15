package repository

import (
	"github.com/dormitory-bot/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type EmployeeDormitoryRoleRepository struct {
	*BaseRepository[domain.EmployeeDormitoryRole]
}

func NewEmployeeDormitoryRoleRepository(db *gorm.DB) *EmployeeDormitoryRoleRepository {
	return &EmployeeDormitoryRoleRepository{
		BaseRepository: NewBaseRepository[domain.EmployeeDormitoryRole](db),
	}
}

func (r *EmployeeDormitoryRoleRepository) CreateOrIgnore(entity *domain.EmployeeDormitoryRole) (bool, error) {
	result := r.DB().Clauses(clause.OnConflict{DoNothing: true}).Create(entity)
	return result.RowsAffected > 0, result.Error
}

func (r *EmployeeDormitoryRoleRepository) GetByEmployeeAndDormitory(employeeID uuid.UUID, dormitoryID uuid.UUID) ([]domain.EmployeeDormitoryRole, error) {
	var roles []domain.EmployeeDormitoryRole
	err := r.DB().Where("employee_id = ? AND dormitory_id = ?", employeeID, dormitoryID).Find(&roles).Error
	return roles, err
}

func (r *EmployeeDormitoryRoleRepository) GetByEmployee(employeeID uuid.UUID) ([]domain.EmployeeDormitoryRole, error) {
	var roles []domain.EmployeeDormitoryRole
	err := r.DB().Where("employee_id = ?", employeeID).Find(&roles).Error
	return roles, err
}

func (r *EmployeeDormitoryRoleRepository) GetByDormitory(dormitoryID uuid.UUID) ([]domain.EmployeeDormitoryRole, error) {
	var roles []domain.EmployeeDormitoryRole
	err := r.DB().Where("dormitory_id = ?", dormitoryID).Find(&roles).Error
	return roles, err
}
