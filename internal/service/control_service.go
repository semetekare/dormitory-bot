package service

import (
	"fmt"
	"time"

	"github.com/dormitory-bot/internal/domain"
	"github.com/dormitory-bot/internal/repository"
	"github.com/google/uuid"
)

type ControlService struct {
	sanitaryCtrl     *repository.BaseRepository[domain.SanitaryControl]
	disciplineCtrl   *repository.BaseRepository[domain.DisciplineControl]
	cohabitationCtrl *repository.BaseRepository[domain.CohabitationControl]
}

func NewControlService(
	sanitaryRepo *repository.BaseRepository[domain.SanitaryControl],
	disciplineRepo *repository.BaseRepository[domain.DisciplineControl],
	cohabitationRepo *repository.BaseRepository[domain.CohabitationControl],
) *ControlService {
	return &ControlService{
		sanitaryCtrl:     sanitaryRepo,
		disciplineCtrl:   disciplineRepo,
		cohabitationCtrl: cohabitationRepo,
	}
}

type ControlResult struct {
	Sanitary     []domain.SanitaryControl
	Discipline   []domain.DisciplineControl
	Cohabitation []domain.CohabitationControl
}

func (s *ControlService) GetRoomControls(roomID uuid.UUID) (*ControlResult, error) {
	result := &ControlResult{}
	sanitary, _ := s.sanitaryCtrl.FindAllByField("room_id", roomID)
	result.Sanitary = sanitary
	discipline, _ := s.disciplineCtrl.FindAllByField("room_id", roomID)
	result.Discipline = discipline
	cohabitation, _ := s.cohabitationCtrl.FindAllByField("room_id", roomID)
	result.Cohabitation = cohabitation
	return result, nil
}

func (s *ControlService) SetSanitaryControl(roomID uuid.UUID, startDate string, endDate string, reason string, inspectorID uuid.UUID) (*domain.SanitaryControl, error) {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, fmt.Errorf("неверный формат даты начала")
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, fmt.Errorf("неверный формат даты окончания")
	}
	today := time.Now().Truncate(24 * time.Hour)
	if start.Before(today) {
		return nil, fmt.Errorf("дата начала не может быть раньше сегодняшнего дня")
	}
	if end.Before(start) {
		return nil, fmt.Errorf("дата окончания не может быть раньше даты начала")
	}

	control := &domain.SanitaryControl{
		RoomID:      roomID,
		StartDate:   startDate,
		EndDate:     endDate,
		Reason:      reason,
		InspectorID: inspectorID,
		Status:      "active",
	}
	if err := s.sanitaryCtrl.Create(control); err != nil {
		return nil, err
	}
	return control, nil
}

func (s *ControlService) CompleteSanitaryControl(controlID uuid.UUID, result string) error {
	return s.sanitaryCtrl.DB().Model(&domain.SanitaryControl{}).
		Where("id = ?", controlID).
		Updates(map[string]interface{}{"status": "completed", "result": result}).Error
}

func (s *ControlService) SetDisciplineControl(roomID uuid.UUID, residentID *uuid.UUID, startDate string, endDate string, violation string, responsibleID uuid.UUID) (*domain.DisciplineControl, error) {
	control := &domain.DisciplineControl{
		RoomID:        roomID,
		ResidentID:    residentID,
		StartDate:     startDate,
		EndDate:       endDate,
		Violation:     violation,
		ResponsibleID: responsibleID,
		Status:        "active",
	}
	if err := s.disciplineCtrl.Create(control); err != nil {
		return nil, err
	}
	return control, nil
}

func (s *ControlService) CompleteDisciplineControl(controlID uuid.UUID, result string) error {
	return s.disciplineCtrl.DB().Model(&domain.DisciplineControl{}).
		Where("id = ?", controlID).
		Updates(map[string]interface{}{"status": "completed", "result": result}).Error
}

func (s *ControlService) SetCohabitationControl(roomID uuid.UUID, startDate string, endDate string, reason string) (*domain.CohabitationControl, error) {
	control := &domain.CohabitationControl{
		RoomID:    roomID,
		StartDate: startDate,
		EndDate:   endDate,
		Reason:    reason,
		Status:    "active",
	}
	if err := s.cohabitationCtrl.Create(control); err != nil {
		return nil, err
	}
	return control, nil
}

func (s *ControlService) CompleteCohabitationControl(controlID uuid.UUID, result string) error {
	return s.cohabitationCtrl.DB().Model(&domain.CohabitationControl{}).
		Where("id = ?", controlID).
		Updates(map[string]interface{}{"status": "completed", "result": result}).Error
}
