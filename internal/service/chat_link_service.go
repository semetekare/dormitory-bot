package service

import (
	"github.com/dormitory-bot/internal/domain"
	"github.com/dormitory-bot/internal/repository"
	"github.com/google/uuid"
)

type ChatLinkService struct {
	repo *repository.BaseRepository[domain.ChatLink]
}

func NewChatLinkService(repo *repository.BaseRepository[domain.ChatLink]) *ChatLinkService {
	return &ChatLinkService{repo: repo}
}

func (s *ChatLinkService) GetForResident(dormitoryID uuid.UUID, floorID uuid.UUID, wingID *uuid.UUID) ([]domain.ChatLink, error) {
	var links []domain.ChatLink
	query := s.repo.DB().Where("dormitory_id = ? AND is_active = true", dormitoryID)

	query = query.Where(
		"(link_type = 'dormitory') OR (link_type = 'floor' AND (floor_id IS NULL OR floor_id = ?)) OR (link_type = 'wing' AND (wing_id IS NULL OR wing_id = ?))",
		floorID, wingID,
	)

	err := query.Order("link_type ASC").Find(&links).Error
	return links, err
}

func (s *ChatLinkService) List(dormitoryID uuid.UUID) ([]domain.ChatLink, error) {
	return s.repo.FindAllByField("dormitory_id", dormitoryID)
}

func (s *ChatLinkService) Create(link *domain.ChatLink) error {
	return s.repo.Create(link)
}

func (s *ChatLinkService) Update(link *domain.ChatLink) error {
	return s.repo.Update(link)
}

func (s *ChatLinkService) Delete(id uuid.UUID) error {
	return s.repo.Delete(id)
}

func (s *ChatLinkService) GetByID(id uuid.UUID) (*domain.ChatLink, error) {
	return s.repo.GetByID(id)
}
