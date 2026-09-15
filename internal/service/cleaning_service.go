package service

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/dormitory-bot/internal/domain"
	"github.com/dormitory-bot/internal/infrastructure/cache"
	"github.com/dormitory-bot/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CleaningService struct {
	dutyRepo      *repository.BaseRepository[domain.CleaningDuty]
	penaltyRepo   *repository.BaseRepository[domain.PenaltyCleaning]
	exemptionRepo *repository.BaseRepository[domain.CleaningExemption]
	roomRepo      *repository.BaseRepository[domain.Room]
	residentRepo  *repository.BaseRepository[domain.Resident]
	txManager     *TransactionManager
	redisClient   *cache.RedisClient
}

func (s *CleaningService) DB() *gorm.DB {
	return s.dutyRepo.DB()
}

func NewCleaningService(
	dutyRepo *repository.BaseRepository[domain.CleaningDuty],
	penaltyRepo *repository.BaseRepository[domain.PenaltyCleaning],
	exemptionRepo *repository.BaseRepository[domain.CleaningExemption],
	roomRepo *repository.BaseRepository[domain.Room],
	residentRepo *repository.BaseRepository[domain.Resident],
	txManager *TransactionManager,
) *CleaningService {
	return &CleaningService{
		dutyRepo:      dutyRepo,
		penaltyRepo:   penaltyRepo,
		exemptionRepo: exemptionRepo,
		roomRepo:      roomRepo,
		residentRepo:  residentRepo,
		txManager:     txManager,
	}
}

func (s *CleaningService) SetRedisClient(redisClient *cache.RedisClient) {
	s.redisClient = redisClient
}

func (s *CleaningService) GetResidentSchedule(roomID uuid.UUID, month int, year int) ([]domain.CleaningDuty, error) {
	startDate := fmt.Sprintf("%d-%02d-01", year, month)
	nextMonth := month + 1
	nextYear := year
	if nextMonth > 12 {
		nextMonth = 1
		nextYear = year + 1
	}
	endDate := fmt.Sprintf("%d-%02d-01", nextYear, nextMonth)

	var duties []domain.CleaningDuty
	err := s.dutyRepo.DB().Where("room_id = ? AND duty_date >= ? AND duty_date < ?",
		roomID, startDate, endDate).
		Order("duty_date ASC").Find(&duties).Error
	return duties, err
}

func (s *CleaningService) GetDormitorySchedule(dormitoryID uuid.UUID, month int, year int) ([]domain.CleaningDuty, error) {
	startDate := fmt.Sprintf("%d-%02d-01", year, month)
	nextMonth := month + 1
	nextYear := year
	if nextMonth > 12 {
		nextMonth = 1
		nextYear = year + 1
	}
	endDate := fmt.Sprintf("%d-%02d-01", nextYear, nextMonth)

	var duties []domain.CleaningDuty
	err := s.dutyRepo.DB().Where("dormitory_id = ? AND duty_date >= ? AND duty_date < ?", dormitoryID, startDate, endDate).
		Order("duty_date ASC").Find(&duties).Error
	return duties, err
}

func (s *CleaningService) generateLockKey(dormitoryID uuid.UUID, month, year int) string {
	return fmt.Sprintf("dorm:locking:cleaning_generate:%s:%d:%d", dormitoryID.String(), month, year)
}

func (s *CleaningService) GenerateSchedule(dormitoryID uuid.UUID, month int, year int) error {
	lockKey := s.generateLockKey(dormitoryID, month, year)
	if s.redisClient != nil {
		acquired, err := s.redisClient.SetNX(context.Background(), lockKey, "1", 30*time.Second)
		if err != nil || !acquired {
			if err != nil {
				return fmt.Errorf("failed to acquire lock: %w", err)
			}
			return fmt.Errorf("график уже генерируется, попробуйте позже")
		}
		defer s.redisClient.Delete(context.Background(), lockKey)
	}

	return s.txManager.ExecuteInTransaction(nil, func(tx *gorm.DB) error {
		startDate := fmt.Sprintf("%d-%02d", year, month)
		daysInMonth := DaysIn(month, year)

		nextM, nextY := month+1, year
		if nextM > 12 {
			nextM = 1
			nextY = year + 1
		}

		s.dutyRepo.DB().Where("dormitory_id = ? AND duty_date >= ? AND duty_date < ? AND status = ?",
			dormitoryID, startDate+"-01", fmt.Sprintf("%d-%02d-01", nextY, nextM), "scheduled").
			Delete(&domain.CleaningDuty{})

		rooms, err := s.roomRepo.FindAllByField("dormitory_id", dormitoryID)
		if err != nil || len(rooms) == 0 {
			return fmt.Errorf("no active rooms")
		}

		var activeRooms []domain.Room
		for _, r := range rooms {
			if r.IsActive {
				activeRooms = append(activeRooms, r)
			}
		}

		if len(activeRooms) == 0 {
			return fmt.Errorf("no active rooms")
		}

		monthStart := fmt.Sprintf("%s-01", startDate)
		monthEnd := fmt.Sprintf("%s-%02d", startDate, daysInMonth)

		exemptions := make([]domain.CleaningExemption, 0)
		s.exemptionRepo.DB().Joins("JOIN room ON room.id = cleaning_exemption.room_id").
			Where("room.dormitory_id = ?", dormitoryID).
			Where("cleaning_exemption.end_date >= ?", monthStart).
			Where("cleaning_exemption.start_date <= ?", monthEnd).
			Find(&exemptions)
		exemptRooms := make(map[uuid.UUID]bool)
		for _, e := range exemptions {
			exemptRooms[e.RoomID] = true
		}

		emptyRooms := make(map[uuid.UUID]bool)
		for _, r := range activeRooms {
			var count int64
			s.residentRepo.DB().Model(&domain.Resident{}).Where("room_id = ?", r.ID).Count(&count)
			if count == 0 {
				emptyRooms[r.ID] = true
			}
		}

		penalties := make([]domain.PenaltyCleaning, 0)
		s.penaltyRepo.DB().Joins("JOIN room ON room.id = penalty_cleaning.room_id").
			Where("room.dormitory_id = ?", dormitoryID).
			Where("penalty_cleaning.completed_count < penalty_cleaning.count").
			Find(&penalties)

		type penaltyDay struct {
			PenaltyID uuid.UUID
			RoomID    uuid.UUID
			Reason    string
		}
		var penaltyDays []penaltyDay
		for _, p := range penalties {
			remaining := p.Count - p.CompletedCount
			for i := 0; i < remaining; i++ {
				penaltyDays = append(penaltyDays, penaltyDay{
					PenaltyID: p.ID,
					RoomID:    p.RoomID,
					Reason:    p.Reason,
				})
			}
		}

		type floorRooms struct {
			FloorID uuid.UUID
			RoomIDs []uuid.UUID
		}
		roomsByFloor := make(map[uuid.UUID]*floorRooms)
		for _, r := range activeRooms {
			if emptyRooms[r.ID] {
				continue
			}
			if _, ok := roomsByFloor[r.FloorID]; !ok {
				roomsByFloor[r.FloorID] = &floorRooms{FloorID: r.FloorID}
			}
			roomsByFloor[r.FloorID].RoomIDs = append(roomsByFloor[r.FloorID].RoomIDs, r.ID)
		}

		rng := rand.New(rand.NewSource(time.Now().UnixNano()))

		occupied := make(map[string]map[uuid.UUID]bool)

		for floorID, fr := range roomsByFloor {
			roomIDs := fr.RoomIDs
			if len(roomIDs) == 0 {
				continue
			}
			for i := len(roomIDs) - 1; i > 0; i-- {
				j := rng.Intn(i + 1)
				roomIDs[i], roomIDs[j] = roomIDs[j], roomIDs[i]
			}

			roomIdx := 0
			for day := 1; day <= daysInMonth; day++ {
				date := fmt.Sprintf("%s-%02d", startDate, day)

				roomID := roomIDs[roomIdx]
				skipCount := 0
				for exemptRooms[roomID] && skipCount < len(roomIDs) {
					roomIdx = (roomIdx + 1) % len(roomIDs)
					roomID = roomIDs[roomIdx]
					skipCount++
				}

				if skipCount < len(roomIDs) {
					duty := &domain.CleaningDuty{
						DormitoryID: dormitoryID,
						FloorID:     floorID,
						RoomID:      roomID,
						DutyDate:    date,
						DutyType:    "regular",
						Status:      "scheduled",
					}
					if err := s.dutyRepo.Create(duty); err != nil {
						return err
					}
					if occupied[date] == nil {
						occupied[date] = make(map[uuid.UUID]bool)
					}
					occupied[date][roomID] = true
				}

				roomIdx = (roomIdx + 1) % len(roomIDs)
			}
		}

		day := 1
		roomMap := make(map[uuid.UUID]domain.Room)
		for _, r := range activeRooms {
			roomMap[r.ID] = r
		}
		for _, pd := range penaltyDays {
			if exemptRooms[pd.RoomID] {
				continue
			}

			for day <= daysInMonth {
				date := fmt.Sprintf("%s-%02d", startDate, day)
				if occupied[date] == nil {
					occupied[date] = make(map[uuid.UUID]bool)
				}
				if !occupied[date][pd.RoomID] {
					break
				}
				day++
			}

			if day > daysInMonth {
				break
			}

			date := fmt.Sprintf("%s-%02d", startDate, day)
			rm := roomMap[pd.RoomID]
			penaltyCleaningID := pd.PenaltyID
			duty := &domain.CleaningDuty{
				DormitoryID:       dormitoryID,
				FloorID:           rm.FloorID,
				RoomID:            pd.RoomID,
				DutyDate:          date,
				DutyType:          "penalty",
				Status:            "scheduled",
				PenaltyCleaningID: &penaltyCleaningID,
			}
			if err := s.dutyRepo.Create(duty); err != nil {
				return err
			}
			occupied[date][pd.RoomID] = true
			day++
		}

		return nil
	})
}

func (s *CleaningService) MarkDutyCompleted(dutyID uuid.UUID, completedBy uuid.UUID) error {
	return s.txManager.ExecuteInTransaction(nil, func(tx *gorm.DB) error {
		var duty domain.CleaningDuty
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", dutyID).First(&duty).Error; err != nil {
			return err
		}
		if duty.Status != "scheduled" {
			return fmt.Errorf("дежурство уже имеет статус: %s", duty.Status)
		}

		now := time.Now()
		if err := tx.Model(&domain.CleaningDuty{}).
			Where("id = ?", dutyID).
			Updates(map[string]interface{}{
				"status":          "completed",
				"completed_by_id": completedBy,
				"completed_at":    now,
			}).Error; err != nil {
			return err
		}

		if duty.DutyType == "penalty" && duty.PenaltyCleaningID != nil {
			if err := tx.Model(&domain.PenaltyCleaning{}).
				Where("id = ?", *duty.PenaltyCleaningID).
				Where("completed_count < count").
				Update("completed_count", gorm.Expr("completed_count + 1")).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *CleaningService) MarkDutyMissed(dutyID uuid.UUID) error {
	return s.txManager.ExecuteInTransaction(nil, func(tx *gorm.DB) error {
		var duty domain.CleaningDuty
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", dutyID).First(&duty).Error; err != nil {
			return err
		}
		if duty.Status != "scheduled" {
			return fmt.Errorf("дежурство уже имеет статус: %s", duty.Status)
		}

		return tx.Model(&domain.CleaningDuty{}).
			Where("id = ?", dutyID).
			Update("status", "missed").Error
	})
}

func (s *CleaningService) CreatePenalty(roomID uuid.UUID, residentID *uuid.UUID, count int, reason string, issuedBy uuid.UUID) (*domain.PenaltyCleaning, error) {
	penalty := &domain.PenaltyCleaning{
		RoomID:         roomID,
		ResidentID:     residentID,
		Count:          count,
		CompletedCount: 0,
		Reason:         reason,
		IssuedByID:     issuedBy,
		IssuedAt:       time.Now().Format("2006-01-02"),
	}
	if err := s.penaltyRepo.Create(penalty); err != nil {
		return nil, err
	}
	return penalty, nil
}

func (s *CleaningService) DeletePenalty(penaltyID uuid.UUID) error {
	return s.penaltyRepo.Delete(penaltyID)
}

func (s *CleaningService) UpdatePenalty(penaltyID uuid.UUID, count int, reason string) error {
	var penalty domain.PenaltyCleaning
	if err := s.penaltyRepo.DB().Where("id = ?", penaltyID).First(&penalty).Error; err != nil {
		return err
	}
	if penalty.CompletedCount > 0 {
		return fmt.Errorf("нельзя изменить штраф, по которому уже есть отбытые дежурства")
	}
	return s.penaltyRepo.DB().Model(&penalty).
		Updates(map[string]interface{}{
			"count":  count,
			"reason": reason,
		}).Error
}

func (s *CleaningService) GetPenaltiesByDormitory(dormitoryID uuid.UUID) ([]domain.PenaltyCleaning, error) {
	var penalties []domain.PenaltyCleaning
	err := s.penaltyRepo.DB().Joins("JOIN room ON room.id = penalty_cleaning.room_id").
		Where("room.dormitory_id = ?", dormitoryID).
		Find(&penalties).Error
	return penalties, err
}

func (s *CleaningService) GetActivePenaltiesByRoom(roomID uuid.UUID) ([]domain.PenaltyCleaning, error) {
	var penalties []domain.PenaltyCleaning
	err := s.penaltyRepo.DB().Where("room_id = ? AND completed_count < count", roomID).
		Find(&penalties).Error
	return penalties, err
}

func (s *CleaningService) CreateExemption(roomID uuid.UUID, startDate string, endDate string, reason string, createdBy uuid.UUID) (*domain.CleaningExemption, error) {
	exemption := &domain.CleaningExemption{
		RoomID:      roomID,
		StartDate:   startDate,
		EndDate:     endDate,
		Reason:      reason,
		CreatedByID: createdBy,
	}
	if err := s.exemptionRepo.Create(exemption); err != nil {
		return nil, err
	}
	return exemption, nil
}

func (s *CleaningService) GetExemptionsByDormitory(dormitoryID uuid.UUID) ([]domain.CleaningExemption, error) {
	var exemptions []domain.CleaningExemption
	err := s.exemptionRepo.DB().Joins("JOIN room ON room.id = cleaning_exemption.room_id").
		Where("room.dormitory_id = ?", dormitoryID).
		Find(&exemptions).Error
	return exemptions, err
}

func (s *CleaningService) HasOverlappingExemption(roomID uuid.UUID, startDate, endDate string) (bool, error) {
	var count int64
	err := s.exemptionRepo.DB().Model(&domain.CleaningExemption{}).
		Where("room_id = ? AND end_date >= ? AND start_date <= ? AND deleted_at IS NULL",
			roomID, startDate, endDate).
		Count(&count).Error
	return count > 0, err
}

func (s *CleaningService) CreateExemptionSafe(roomID uuid.UUID, startDate string, endDate string, reason string, createdBy uuid.UUID) (*domain.CleaningExemption, error) {
	overlap, err := s.HasOverlappingExemption(roomID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	if overlap {
		return nil, fmt.Errorf("комната уже освобождена на пересекающийся период")
	}
	return s.CreateExemption(roomID, startDate, endDate, reason, createdBy)
}

func (s *CleaningService) TransferDuty(dutyID uuid.UUID, newRoomID uuid.UUID) error {
	return s.txManager.ExecuteInTransaction(nil, func(tx *gorm.DB) error {
		var duty domain.CleaningDuty
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", dutyID).First(&duty).Error; err != nil {
			return err
		}

		var newRoom domain.Room
		if err := tx.Where("id = ?", newRoomID).First(&newRoom).Error; err != nil {
			return fmt.Errorf("комната не найдена")
		}

		if err := tx.Model(&duty).Updates(map[string]interface{}{
			"room_id":  newRoomID,
			"floor_id": newRoom.FloorID,
		}).Error; err != nil {
			return err
		}

		return nil
	})
}

func (s *CleaningService) TransferDutyDate(dutyID uuid.UUID, newDate string) error {
	return s.txManager.ExecuteInTransaction(nil, func(tx *gorm.DB) error {
		var duty domain.CleaningDuty
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", dutyID).First(&duty).Error; err != nil {
			return err
		}
		if duty.Status != "scheduled" {
			return fmt.Errorf("можно переносить только запланированные дежурства")
		}

		return tx.Model(&duty).Update("duty_date", newDate).Error
	})
}

func DaysIn(month int, year int) int {
	if month == 2 {
		if year%4 == 0 && (year%100 != 0 || year%400 == 0) {
			return 29
		}
		return 28
	}
	if month == 4 || month == 6 || month == 9 || month == 11 {
		return 30
	}
	return 31
}

// GenerateScheduleMulti generates a cleaning schedule for multiple months starting from
// the current month. Uses a global lock per dormitory.
func (s *CleaningService) GenerateScheduleMulti(dormitoryID uuid.UUID, monthsCount int) error {
	lockKey := fmt.Sprintf("dorm:locking:cleaning_generate:%s", dormitoryID.String())
	if s.redisClient != nil {
		acquired, err := s.redisClient.SetNX(context.Background(), lockKey, "1", 60*time.Second)
		if err != nil || !acquired {
			if err != nil {
				return fmt.Errorf("failed to acquire lock: %w", err)
			}
			return fmt.Errorf("график уже генерируется, попробуйте позже")
		}
		defer s.redisClient.Delete(context.Background(), lockKey)
	}

	now := time.Now()
	for i := 0; i < monthsCount; i++ {
		m := int(now.Month()) + i
		y := now.Year()
		if m > 12 {
			m -= 12
			y = now.Year() + 1
		}
		if err := s.GenerateSchedule(dormitoryID, m, y); err != nil {
			// Don't fail completely on one month's error — log and continue.
			log.Printf("GenerateScheduleMulti: error for %d-%d: %v", y, m, err)
		}
	}
	return nil
}

// GetScheduleForDate returns all scheduled duties for a given date.
func (s *CleaningService) GetScheduleForDate(dormitoryID uuid.UUID, date string) ([]domain.CleaningDuty, error) {
	return s.dutyRepo.FindAllByField("dormitory_id", dormitoryID)
}

// GetResidentDutiesForMonth returns all duties for a resident's room in a given month.
func (s *CleaningService) GetResidentDutiesForMonth(roomID uuid.UUID, month, year int) ([]domain.CleaningDuty, error) {
	startDate := fmt.Sprintf("%d-%02d-01", year, month)
	nextM, nextY := month+1, year
	if nextM > 12 {
		nextM = 1
		nextY = year + 1
	}
	endDate := fmt.Sprintf("%d-%02d-01", nextY, nextM)

	var duties []domain.CleaningDuty
	err := s.dutyRepo.DB().
		Where("room_id = ? AND duty_date >= ? AND duty_date < ?", roomID, startDate, endDate).
		Order("duty_date ASC").Find(&duties).Error
	return duties, err
}

// GetOrCreateCleaningSettings returns settings for a dormitory, creating defaults if absent.
func (s *CleaningService) GetOrCreateCleaningSettings(dormitoryID uuid.UUID) (*domain.CleaningSettings, error) {
	var cs domain.CleaningSettings
	err := s.dutyRepo.DB().Where("dormitory_id = ?", dormitoryID).First(&cs).Error
	if err == nil {
		return &cs, nil
	}
	cs = domain.CleaningSettings{
		DormitoryID:    dormitoryID,
		DutyStartTime:  "18:00",
		DutyDuration:   60,
		ReminderBefore: 60,
		ActiveDays:     "1,2,3,4,5,6,7",
	}
	if err := s.dutyRepo.DB().Create(&cs).Error; err != nil {
		return nil, err
	}
	return &cs, nil
}

// UpdateCleaningSettings upserts cleaning settings.
func (s *CleaningService) UpdateCleaningSettings(cs *domain.CleaningSettings) error {
	var existing domain.CleaningSettings
	err := s.dutyRepo.DB().Where("dormitory_id = ?", cs.DormitoryID).First(&existing).Error
	if err != nil {
		return s.dutyRepo.DB().Create(cs).Error
	}
	cs.ID = existing.ID
	return s.dutyRepo.DB().Save(cs).Error
}
