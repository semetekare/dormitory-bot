package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/dormitory-bot/internal/domain"
	"github.com/dormitory-bot/internal/repository"
	"github.com/dormitory-bot/internal/timeutil"
	"github.com/google/uuid"
)

type LaundryService struct {
	machineRepo  *repository.BaseRepository[domain.WashingMachine]
	bookingRepo  *repository.LaundryBookingRepository
	settingsRepo *repository.BaseRepository[domain.LaundrySettings]
	txManager    *TransactionManager
}

type LaundrySlot struct {
	MachineID     string
	MachineNumber string
	SlotStart     string
	SlotEnd       string
}

type LaundryBookingResult struct {
	BookingID     string
	MachineNumber string
	SlotDate      string
	SlotStart     string
	SlotEnd       string
}

func NewLaundryService(
	machineRepo *repository.BaseRepository[domain.WashingMachine],
	bookingRepo *repository.LaundryBookingRepository,
	settingsRepo *repository.BaseRepository[domain.LaundrySettings],
	txManager *TransactionManager,
) *LaundryService {
	return &LaundryService{
		machineRepo:  machineRepo,
		bookingRepo:  bookingRepo,
		settingsRepo: settingsRepo,
		txManager:    txManager,
	}
}

func (s *LaundryService) GetAvailableSlots(dormitoryID uuid.UUID, date string) ([]LaundrySlot, error) {
	settings, err := s.settingsRepo.FindByField("dormitory_id", dormitoryID)
	var startTime string

	if err != nil || settings == nil {
		startTime = "07:00"
	} else {
		startTime = settings.BookingStartTime
	}

	endTimeStr := "23:00"
	if settings != nil && settings.BookingEndTime != "" {
		endTimeStr = settings.BookingEndTime
	}
	castellanEnabled := settings != nil && settings.CastellanBookingEnabled

	today := timeutil.Today()
	nowStr := timeutil.TimeStr()

	machines, err := s.machineRepo.FindAllByField("dormitory_id", dormitoryID)
	if err != nil {
		return nil, err
	}

	var activeMachines []domain.WashingMachine
	for _, m := range machines {
		if m.IsActive {
			activeMachines = append(activeMachines, m)
		}
	}

	if len(activeMachines) == 0 {
		return nil, nil
	}

	var allSlots []LaundrySlot

	for _, machine := range activeMachines {
		existingBookings, _ := s.bookingRepo.FindByMachineAndDate(machine.ID, date)

		// Check for custom slots first.
		var customSlots []domain.LaundrySlot
		s.machineRepo.DB().Where("machine_id = ? AND is_active = true", machine.ID).
			Order("slot_start ASC").Find(&customSlots)

		if len(customSlots) > 0 {
			for _, cs := range customSlots {
				if cs.DaysOfWeek != "" {
					weekday := int(time.Now().Weekday())
					if weekday == 0 {
						weekday = 7
					}
					found := strings.Contains(cs.DaysOfWeek, fmt.Sprintf("%d", weekday))
					if !found {
						continue
					}
				}

				conflict := false
				for _, b := range existingBookings {
					if b.SlotStart < cs.SlotEnd && b.SlotEnd > cs.SlotStart {
						conflict = true
						break
					}
				}
				isBlocked := castellanEnabled && machine.BlockedSlotStart != "" &&
					cs.SlotStart < machine.BlockedSlotEnd && cs.SlotEnd > machine.BlockedSlotStart

				if !conflict && !isBlocked {
					if date == today {
						if cs.SlotStart > nowStr {
							allSlots = append(allSlots, LaundrySlot{
								MachineID:     machine.ID.String(),
								MachineNumber: machine.MachineNumber,
								SlotStart:     cs.SlotStart,
								SlotEnd:       cs.SlotEnd,
							})
						}
					} else {
						allSlots = append(allSlots, LaundrySlot{
							MachineID:     machine.ID.String(),
							MachineNumber: machine.MachineNumber,
							SlotStart:     cs.SlotStart,
							SlotEnd:       cs.SlotEnd,
						})
					}
				}
			}
			continue
		}

		// Fallback: generate slots from DurationMinutes.
		duration := machine.DurationMinutes
		if duration == 0 {
			duration = 60
		}
		var defaultStartHour, defaultStartMin int
		fmt.Sscanf(startTime, "%d:%d", &defaultStartHour, &defaultStartMin)
		var endHour, endMin int
		fmt.Sscanf(endTimeStr, "%d:%d", &endHour, &endMin)
		slotStart := time.Date(0, 1, 1, defaultStartHour, defaultStartMin, 0, 0, time.UTC)
		slotEndTime := time.Date(0, 1, 1, endHour, endMin, 0, 0, time.UTC)

		bookedSlots := make(map[string]bool)
		for _, b := range existingBookings {
			bookedSlots[b.SlotStart] = true
		}

		for slotStart.Before(slotEndTime) {
			slotEnd := slotStart.Add(time.Duration(duration) * time.Minute)
			slotStartStr := slotStart.Format("15:04")
			slotEndStr := slotEnd.Format("15:04")

			if slotEnd.After(slotEndTime) {
				break
			}

			conflict := false
			for _, b := range existingBookings {
				if b.SlotStart < slotEndStr && b.SlotEnd > slotStartStr {
					conflict = true
					break
				}
			}

			isBlocked := castellanEnabled && machine.BlockedSlotStart != "" && slotStartStr < machine.BlockedSlotEnd && slotEndStr > machine.BlockedSlotStart

			if !conflict && !isBlocked {
				if date == today {
					if slotStartStr > nowStr {
						allSlots = append(allSlots, LaundrySlot{
							MachineID:     machine.ID.String(),
							MachineNumber: machine.MachineNumber,
							SlotStart:     slotStartStr,
							SlotEnd:       slotEndStr,
						})
					}
				} else {
					allSlots = append(allSlots, LaundrySlot{
						MachineID:     machine.ID.String(),
						MachineNumber: machine.MachineNumber,
						SlotStart:     slotStartStr,
						SlotEnd:       slotEndStr,
					})
				}
			}
			_ = bookedSlots

			slotStart = slotEnd
		}
	}

	return allSlots, nil
}

func (s *LaundryService) CreateBooking(userID uuid.UUID, bookerType string, machineID uuid.UUID, date string, slotStart string, slotEnd string) (*LaundryBookingResult, error) {
	conflict, err := s.bookingRepo.FindConflict(machineID, date, slotStart, slotEnd)
	if err == nil && conflict != nil {
		return nil, fmt.Errorf("slot already booked")
	}

	existing, _ := s.bookingRepo.FindActiveByUser(userID)
	if len(existing) > 0 {
		return nil, fmt.Errorf("you already have an active booking")
	}

	booking := &domain.LaundryBooking{
		MachineID:  machineID,
		UserID:     userID,
		BookerType: bookerType,
		SlotDate:   date,
		SlotStart:  slotStart,
		SlotEnd:    slotEnd,
		Status:     "active",
	}

	if err := s.bookingRepo.Create(booking); err != nil {
		return nil, err
	}

	machine, err := s.machineRepo.GetByID(machineID)
	var machineNumber string
	if err == nil {
		machineNumber = machine.MachineNumber
	}

	return &LaundryBookingResult{
		BookingID:     booking.ID.String(),
		MachineNumber: machineNumber,
		SlotDate:      date,
		SlotStart:     slotStart,
		SlotEnd:       slotEnd,
	}, nil
}

func (s *LaundryService) CancelBooking(bookingID uuid.UUID, userID uuid.UUID) error {
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return fmt.Errorf("booking not found")
	}

	if booking.UserID != userID {
		return fmt.Errorf("not your booking")
	}

	if booking.SlotDate == timeutil.Today() {
		bookingTime, _ := time.Parse("15:04", booking.SlotStart)
		cutoff := timeutil.Now().Add(15 * time.Minute)
		if cutoff.Format("15:04") >= bookingTime.Format("15:04") {
			return fmt.Errorf("отмена невозможна: до начала стирки менее 15 минут")
		}
	}

	return s.bookingRepo.CancelBooking(bookingID, "resident")
}

func (s *LaundryService) GetUserBookings(userID uuid.UUID) ([]domain.LaundryBooking, error) {
	return s.bookingRepo.FindActiveByUser(userID)
}

func (s *LaundryService) GetUserAllBookings(userID uuid.UUID) ([]domain.LaundryBooking, error) {
	return s.bookingRepo.FindByUser(userID)
}

func (s *LaundryService) GetDormitoryBookings(dormitoryID uuid.UUID, date string) ([]domain.LaundryBooking, error) {
	s.bookingRepo.DB().Model(&domain.LaundryBooking{}).
		Where("slot_date = ? AND slot_end <= ? AND status = 'active'", date, timeutil.TimeStr()).
		Update("status", "completed")
	return s.bookingRepo.FindActiveByDormitoryAndDate(dormitoryID, date)
}

func (s *LaundryService) GetAllDormitoryBookings(dormitoryID uuid.UUID, date string) ([]domain.LaundryBooking, error) {
	return s.bookingRepo.FindAllByDormitoryAndDate(dormitoryID, date)
}

func (s *LaundryService) GetWashingMachines(dormitoryID uuid.UUID) ([]domain.WashingMachine, error) {
	var machines []domain.WashingMachine
	if err := s.machineRepo.DB().Where("dormitory_id = ?", dormitoryID).
		Order("floor_id ASC, machine_number ASC").
		Find(&machines).Error; err != nil {
		return nil, err
	}
	return machines, nil
}

func (s *LaundryService) CreateWashingMachine(dormitoryID uuid.UUID, floorID uuid.UUID, number string, duration int) (*domain.WashingMachine, error) {
	machine := &domain.WashingMachine{
		DormitoryID:     dormitoryID,
		FloorID:         floorID,
		MachineNumber:   number,
		DurationMinutes: duration,
		IsActive:        true,
	}
	if err := s.machineRepo.Create(machine); err != nil {
		return nil, err
	}
	return machine, nil
}

func (s *LaundryService) UpdateWashingMachine(machineID uuid.UUID, floorID uuid.UUID, number string, duration int, isActive bool) error {
	machine, err := s.machineRepo.GetByID(machineID)
	if err != nil {
		return err
	}
	machine.FloorID = floorID
	machine.MachineNumber = number
	machine.DurationMinutes = duration
	machine.IsActive = isActive
	return s.machineRepo.Update(machine)
}

func (s *LaundryService) UpdateAllMachinesDuration(dormitoryID uuid.UUID, duration int) error {
	return s.machineRepo.DB().Model(&domain.WashingMachine{}).
		Where("dormitory_id = ?", dormitoryID).
		Update("duration_minutes", duration).Error
}

func (s *LaundryService) SetWashingMachineBlockedSlot(machineID uuid.UUID, start, end, reason string) error {
	machine, err := s.machineRepo.GetByID(machineID)
	if err != nil {
		return err
	}
	machine.BlockedSlotStart = start
	machine.BlockedSlotEnd = end
	machine.BlockedSlotReason = reason
	return s.machineRepo.Update(machine)
}

func (s *LaundryService) AdminCancelBooking(bookingID uuid.UUID) error {
	if _, err := s.bookingRepo.GetByID(bookingID); err != nil {
		return fmt.Errorf("booking not found")
	}
	return s.bookingRepo.CancelBooking(bookingID, "employee")
}

func (s *LaundryService) CleanupExpiredBookings() error {
	now := timeutil.Now()
	today := now.Format("2006-01-02")
	nowStr := now.Format("15:04")

	if err := s.bookingRepo.DB().Model(&domain.LaundryBooking{}).
		Where("slot_date < ? AND status = 'active'", today).
		Update("status", "completed").Error; err != nil {
		return err
	}

	return s.bookingRepo.DB().Model(&domain.LaundryBooking{}).
		Where("slot_date = ? AND slot_end <= ? AND status = 'active'", today, nowStr).
		Update("status", "completed").Error
}

func (s *LaundryService) DeleteWashingMachine(id uuid.UUID) error {
	return s.machineRepo.Delete(id)
}

func (s *LaundryService) GetLaundrySettings(dormitoryID uuid.UUID) (*domain.LaundrySettings, error) {
	settings, err := s.settingsRepo.FindByField("dormitory_id", dormitoryID)
	if err != nil {
		return nil, err
	}
	return settings, nil
}

func (s *LaundryService) UpdateLaundrySettings(settings *domain.LaundrySettings) error {
	existing, err := s.settingsRepo.FindByField("dormitory_id", settings.DormitoryID)
	if err != nil {
		return s.settingsRepo.Create(settings)
	}
	settings.ID = existing.ID
	settings.DormitoryID = existing.DormitoryID
	return s.settingsRepo.Update(settings)
}
