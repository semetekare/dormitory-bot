package service

import (
	"fmt"

	"github.com/dormitory-bot/internal/domain"
	"github.com/dormitory-bot/internal/repository"
	"github.com/google/uuid"
)

type ResidentService struct {
	residentRepo   *repository.BaseRepository[domain.Resident]
	userRepo       *repository.BaseRepository[domain.User]
	roomRepo       *repository.BaseRepository[domain.Room]
	dormitoryRepo  *repository.BaseRepository[domain.Dormitory]
	floorRepo      *repository.BaseRepository[domain.Floor]
	debtRepo       *repository.BaseRepository[domain.ResidentFinancialDebt]
	cleaningRepo   *repository.BaseRepository[domain.CleaningDuty]
	controlService *ControlService
}

func NewResidentService(
	residentRepo *repository.BaseRepository[domain.Resident],
	userRepo *repository.BaseRepository[domain.User],
	roomRepo *repository.BaseRepository[domain.Room],
	dormitoryRepo *repository.BaseRepository[domain.Dormitory],
	floorRepo *repository.BaseRepository[domain.Floor],
	debtRepo *repository.BaseRepository[domain.ResidentFinancialDebt],
	cleaningRepo *repository.BaseRepository[domain.CleaningDuty],
	controlService *ControlService,
) *ResidentService {
	return &ResidentService{
		residentRepo:   residentRepo,
		userRepo:       userRepo,
		roomRepo:       roomRepo,
		dormitoryRepo:  dormitoryRepo,
		floorRepo:      floorRepo,
		debtRepo:       debtRepo,
		cleaningRepo:   cleaningRepo,
		controlService: controlService,
	}
}

type ResidentProfile struct {
	Resident       domain.Resident
	User           domain.User
	Room           domain.Room
	Dormitory      domain.Dormitory
	Floor          domain.Floor
	Debts          []domain.ResidentFinancialDebt
	Controls       *ControlResult
	CleaningDuties []domain.CleaningDuty
	PlatformUserID string
}

func (s *ResidentService) GetFullProfile(residentID uuid.UUID) (*ResidentProfile, error) {
	resident, err := s.residentRepo.GetByID(residentID)
	if err != nil {
		return nil, err
	}

	profile := &ResidentProfile{Resident: *resident}

	user, err := s.userRepo.GetByID(resident.UserID)
	if err == nil {
		profile.User = *user
		profile.PlatformUserID = user.PlatformUserID
	}

	room, err := s.roomRepo.GetByID(resident.RoomID)
	if err == nil {
		profile.Room = *room
	}

	dorm, err := s.dormitoryRepo.GetByID(resident.DormitoryID)
	if err == nil {
		profile.Dormitory = *dorm
	}

	floor, err := s.floorRepo.GetByID(room.FloorID)
	if err == nil {
		profile.Floor = *floor
	}

	debts, _ := s.debtRepo.FindAllByField("resident_id", residentID)
	profile.Debts = debts

	controls, _ := s.controlService.GetRoomControls(resident.RoomID)
	profile.Controls = controls

	duties, _ := s.cleaningRepo.FindAllByField("completed_by_id", residentID)
	profile.CleaningDuties = duties

	return profile, nil
}

func (s *ResidentService) GetDebt(residentID uuid.UUID) ([]domain.ResidentFinancialDebt, error) {
	return s.debtRepo.FindAllByField("resident_id", residentID)
}

func (s *ResidentService) SearchRooms(dormitoryID uuid.UUID, query string) ([]domain.Room, error) {
	var rooms []domain.Room
	err := s.roomRepo.DB().Where("dormitory_id = ? AND room_number ILIKE ?", dormitoryID, "%"+query+"%").
		Find(&rooms).Error
	return rooms, err
}

func (s *ResidentService) GetFloorsByDormitory(dormitoryID uuid.UUID) ([]domain.Floor, error) {
	return s.floorRepo.FindAllByField("dormitory_id", dormitoryID)
}

func (s *ResidentService) GetResidentsWithUser(ids []uuid.UUID) (map[uuid.UUID]*domain.Resident, error) {
	if len(ids) == 0 {
		return map[uuid.UUID]*domain.Resident{}, nil
	}
	var residents []domain.Resident
	err := s.residentRepo.DB().Preload("User").Where("id IN ?", ids).Find(&residents).Error
	if err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]*domain.Resident)
	for i := range residents {
		result[residents[i].ID] = &residents[i]
	}
	return result, nil
}

func (s *ResidentService) GetRoomsByFloor(floorID uuid.UUID) ([]domain.Room, error) {
	return s.roomRepo.FindAllByField("floor_id", floorID)
}

func (s *ResidentService) GetRoomWithResidents(roomID uuid.UUID) ([]domain.Resident, error) {
	var residents []domain.Resident
	err := s.residentRepo.DB().Where("room_id = ?", roomID).Find(&residents).Error
	return residents, err
}

func (s *ResidentService) FormatResidentProfile(profile *ResidentProfile) string {
	text := fmt.Sprintf("👤 Житель: %s %s %s\n", profile.User.LastName, profile.User.FirstName, profile.User.MiddleName)
	text += fmt.Sprintf("📞 Телефон: %s\n", profile.User.Phone)
	text += fmt.Sprintf("🏠 Общежитие: %s\n", profile.Dormitory.Name)
	text += fmt.Sprintf("🚪 Комната: %s, этаж %d\n", profile.Room.RoomNumber, profile.Floor.FloorNumber)

	if len(profile.Debts) > 0 {
		text += "\n💰 Задолженности:\n"
		for _, d := range profile.Debts {
			text += fmt.Sprintf("  • %s: %.2f руб. (%s)\n", d.Period, d.Amount, d.Description)
		}
	}

	if profile.Controls != nil {
		for _, c := range profile.Controls.Sanitary {
			if c.Status == "active" {
				text += fmt.Sprintf("\n🧹 Санконтроль: %s – %s", c.StartDate, c.EndDate)
				if c.Reason != "" {
					text += fmt.Sprintf("\n   Причина: %s", c.Reason)
				}
			}
		}
		for _, c := range profile.Controls.Discipline {
			if c.Status == "active" {
				text += fmt.Sprintf("\n⚠️ Дисцип. контроль: %s – %s", c.StartDate, c.EndDate)
				if c.Violation != "" {
					text += fmt.Sprintf("\n   Нарушение: %s", c.Violation)
				}
			}
		}
		for _, c := range profile.Controls.Cohabitation {
			if c.Status == "active" {
				text += fmt.Sprintf("\n🏠 Контроль проживания: %s – %s", c.StartDate, c.EndDate)
				if c.Reason != "" {
					text += fmt.Sprintf("\n   Причина: %s", c.Reason)
				}
			}
		}
	}

	return text
}
