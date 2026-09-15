package repository

import (
	"github.com/dormitory-bot/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserDormitoryRoleRepository struct {
	*BaseRepository[domain.UserDormitoryRole]
}

func NewUserDormitoryRoleRepository(db *gorm.DB) *UserDormitoryRoleRepository {
	return &UserDormitoryRoleRepository{
		BaseRepository: NewBaseRepository[domain.UserDormitoryRole](db),
	}
}

func (r *UserDormitoryRoleRepository) GetByUserAndDormitory(userID, dormitoryID uuid.UUID) ([]domain.UserDormitoryRole, error) {
	var roles []domain.UserDormitoryRole
	err := r.DB().Where("user_id = ? AND dormitory_id = ?", userID, dormitoryID).Find(&roles).Error
	return roles, err
}

func (r *UserDormitoryRoleRepository) GetByUser(userID uuid.UUID) ([]domain.UserDormitoryRole, error) {
	var roles []domain.UserDormitoryRole
	err := r.DB().Where("user_id = ?", userID).Find(&roles).Error
	return roles, err
}

func (r *UserDormitoryRoleRepository) GetByDormitory(dormitoryID uuid.UUID) ([]domain.UserDormitoryRole, error) {
	var roles []domain.UserDormitoryRole
	err := r.DB().Where("dormitory_id = ?", dormitoryID).Find(&roles).Error
	return roles, err
}

func (r *UserDormitoryRoleRepository) GetByDormitoryWithUser(dormitoryID uuid.UUID) ([]domain.UserDormitoryRoleWithUser, error) {
	var results []domain.UserDormitoryRoleWithUser
	err := r.DB().
		Table("user_dormitory_role").
		Select("user_dormitory_role.*, \"user\".first_name as user_first_name, \"user\".last_name as user_last_name, \"user\".middle_name as user_middle_name, \"user\".phone as user_phone, \"user\".person_type as person_type, \"user\".platform_user_id as platform_user_id").
		Joins("JOIN \"user\" ON \"user\".id = user_dormitory_role.user_id").
		Where("user_dormitory_role.dormitory_id = ?", dormitoryID).
		Order("user_dormitory_role.created_at DESC").
		Find(&results).Error
	return results, err
}

func (r *UserDormitoryRoleRepository) GetByUserIDAndDormitory(userID, dormitoryID uuid.UUID) ([]domain.UserDormitoryRole, error) {
	return r.GetByUserAndDormitory(userID, dormitoryID)
}

func (r *UserDormitoryRoleRepository) DeleteByUserAndDormitory(userID, dormitoryID uuid.UUID) error {
	return r.DB().Where("user_id = ? AND dormitory_id = ?", userID, dormitoryID).Delete(&domain.UserDormitoryRole{}).Error
}

func (r *UserDormitoryRoleRepository) CountByUserAndDormitory(userID, dormitoryID uuid.UUID) (int64, error) {
	var count int64
	err := r.DB().Model(&domain.UserDormitoryRole{}).Where("user_id = ? AND dormitory_id = ?", userID, dormitoryID).Count(&count).Error
	return count, err
}
