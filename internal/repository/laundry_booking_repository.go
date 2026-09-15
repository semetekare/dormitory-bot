package repository

import (
	"time"

	"github.com/dormitory-bot/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LaundryBookingRepository struct {
	*BaseRepository[domain.LaundryBooking]
}

func NewLaundryBookingRepository(db *gorm.DB) *LaundryBookingRepository {
	return &LaundryBookingRepository{
		BaseRepository: NewBaseRepository[domain.LaundryBooking](db),
	}
}

func (r *LaundryBookingRepository) FindByUserAndDate(userID uuid.UUID, date string) ([]domain.LaundryBooking, error) {
	var bookings []domain.LaundryBooking
	err := r.DB().Where("user_id = ? AND slot_date = ? AND status = 'active'", userID, date).
		Order("slot_start ASC").Find(&bookings).Error
	return bookings, err
}

func (r *LaundryBookingRepository) FindByMachineAndDate(machineID uuid.UUID, date string) ([]domain.LaundryBooking, error) {
	var bookings []domain.LaundryBooking
	err := r.DB().Where("machine_id = ? AND slot_date = ? AND status = 'active'", machineID, date).
		Order("slot_start ASC").Find(&bookings).Error
	return bookings, err
}

func (r *LaundryBookingRepository) FindActiveByDormitoryAndDate(dormitoryID uuid.UUID, date string) ([]domain.LaundryBooking, error) {
	var bookings []domain.LaundryBooking
	err := r.DB().Joins("JOIN washing_machine ON washing_machine.id = laundry_booking.machine_id").
		Where("washing_machine.dormitory_id = ? AND laundry_booking.slot_date = ? AND laundry_booking.status = 'active'", dormitoryID, date).
		Order("laundry_booking.slot_start ASC").
		Find(&bookings).Error
	return bookings, err
}

func (r *LaundryBookingRepository) FindAllByDormitoryAndDate(dormitoryID uuid.UUID, date string) ([]domain.LaundryBooking, error) {
	var bookings []domain.LaundryBooking
	err := r.DB().Joins("JOIN washing_machine ON washing_machine.id = laundry_booking.machine_id").
		Where("washing_machine.dormitory_id = ? AND laundry_booking.slot_date = ?", dormitoryID, date).
		Order("laundry_booking.slot_start ASC").
		Find(&bookings).Error
	return bookings, err
}

func (r *LaundryBookingRepository) FindConflict(machineID uuid.UUID, date string, slotStart string, slotEnd string) (*domain.LaundryBooking, error) {
	var booking domain.LaundryBooking
	err := r.DB().Where("machine_id = ? AND slot_date = ? AND status = 'active' AND slot_start < ? AND slot_end > ?",
		machineID, date, slotEnd, slotStart).
		First(&booking).Error
	if err != nil {
		return nil, err
	}
	return &booking, nil
}

func (r *LaundryBookingRepository) CancelBooking(bookingID uuid.UUID, cancelledBy string) error {
	return r.DB().Model(&domain.LaundryBooking{}).
		Where("id = ? AND status = 'active'", bookingID).
		Updates(map[string]interface{}{
			"status":       "cancelled",
			"cancelled_by": cancelledBy,
		}).Error
}

func (r *LaundryBookingRepository) FindActiveByUser(userID uuid.UUID) ([]domain.LaundryBooking, error) {
	var bookings []domain.LaundryBooking
	now := time.Now()
	today := now.Format("2006-01-02")
	nowStr := now.Format("15:04")
	err := r.DB().Where(
		"user_id = ? AND status = 'active' AND (slot_date > ? OR (slot_date = ? AND slot_end > ?))",
		userID, today, today, nowStr,
	).Order("slot_date ASC, slot_start ASC").Find(&bookings).Error
	return bookings, err
}

func (r *LaundryBookingRepository) FindByUser(userID uuid.UUID) ([]domain.LaundryBooking, error) {
	var bookings []domain.LaundryBooking
	err := r.DB().Where("user_id = ?", userID).
		Order("slot_date DESC, slot_start DESC").Find(&bookings).Error
	return bookings, err
}
