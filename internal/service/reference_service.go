package service

import (
	"github.com/dormitory-bot/internal/domain"
	"github.com/dormitory-bot/internal/repository"
	"github.com/google/uuid"
)

type ReferenceService struct {
	refRepo *repository.BaseRepository[domain.ReferenceMaterial]
}

func NewReferenceService(refRepo *repository.BaseRepository[domain.ReferenceMaterial]) *ReferenceService {
	return &ReferenceService{refRepo: refRepo}
}

func (s *ReferenceService) GetByDormitory(dormitoryID uuid.UUID) ([]domain.ReferenceMaterial, error) {
	materials, err := s.refRepo.FindAllByField("dormitory_id", dormitoryID)
	if err != nil {
		return nil, err
	}
	return materials, nil
}

func (s *ReferenceService) GetByCategory(dormitoryID uuid.UUID, category string) ([]domain.ReferenceMaterial, error) {
	var materials []domain.ReferenceMaterial
	err := s.refRepo.DB().Where("dormitory_id = ? AND category = ?", dormitoryID, category).
		Order("ordinal ASC").Find(&materials).Error
	return materials, err
}

func (s *ReferenceService) GetCategories(dormitoryID uuid.UUID) ([]string, error) {
	var categories []string
	err := s.refRepo.DB().Model(&domain.ReferenceMaterial{}).
		Where("dormitory_id = ?", dormitoryID).
		Distinct("category").Pluck("category", &categories).Error
	return categories, err
}

func (s *ReferenceService) GetByModule(dormitoryID uuid.UUID, moduleKey string) ([]domain.ReferenceMaterial, error) {
	var materials []domain.ReferenceMaterial
	err := s.refRepo.DB().
		Where("dormitory_id = ? AND module_key = ?", dormitoryID, moduleKey).
		Order("ordinal ASC").Find(&materials).Error
	return materials, err
}

func (s *ReferenceService) Create(material *domain.ReferenceMaterial) error {
	return s.refRepo.Create(material)
}

func (s *ReferenceService) Update(material *domain.ReferenceMaterial) error {
	return s.refRepo.Update(material)
}

func (s *ReferenceService) Delete(id uuid.UUID) error {
	return s.refRepo.Delete(id)
}

func (s *ReferenceService) GetByID(id uuid.UUID) (*domain.ReferenceMaterial, error) {
	return s.refRepo.GetByID(id)
}
