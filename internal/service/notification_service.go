package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/dormitory-bot/internal/infrastructure/cache"
	"github.com/dormitory-bot/internal/timeutil"
	"github.com/redis/go-redis/v9"
)

type NotificationService struct {
	redis *cache.RedisClient
}

func NewNotificationService(redis *cache.RedisClient) *NotificationService {
	return &NotificationService{redis: redis}
}

func SetReminderTimezone(name string) {
	timeutil.SetLocation(name)
}

func (s *NotificationService) ScheduleBookingReminders(ctx context.Context, machineNumber string, slotDate string, slotStart string, slotEnd string, userID int64) error {
	layout := "2006-01-02 15:04"
	startStr := slotDate + " " + slotStart
	endStr := slotDate + " " + slotEnd

	loc := timeutil.Location()
	startTime, err := time.ParseInLocation(layout, startStr, loc)
	if err != nil {
		return fmt.Errorf("failed to parse start time: %w", err)
	}
	endTime, err := time.ParseInLocation(layout, endStr, loc)
	if err != nil {
		return fmt.Errorf("failed to parse end time: %w", err)
	}

	now := timeutil.Now()

	remindBefore := startTime.Add(-15 * time.Minute)
	endRemind := endTime.Add(-10 * time.Minute)

	if remindBefore.After(now) {
		err := s.ScheduleReminder(ctx, fmt.Sprintf("remind_start_%s_%s_%d", slotDate, slotStart, userID), fmt.Sprintf("%d", userID),
			fmt.Sprintf("Через 15 минут ваша стирка. Машина №%s.", machineNumber), remindBefore)
		if err != nil {
			log.Printf("Failed to schedule start reminder: %v", err)
		}
	}

	if endRemind.After(now) {
		err := s.ScheduleReminder(ctx, fmt.Sprintf("remind_end_%s_%s_%d", slotDate, slotStart, userID), fmt.Sprintf("%d", userID),
			fmt.Sprintf("Стирка завершается через 10 минут. Машина №%s.", machineNumber), endRemind)
		if err != nil {
			log.Printf("Failed to schedule end reminder: %v", err)
		}
	}

	return nil
}

func (s *NotificationService) ScheduleReminder(ctx context.Context, bookingID string, residentID string, reminderText string, scheduledAt time.Time) error {
	key := fmt.Sprintf("dorm:reminders")

	reminderData := fmt.Sprintf("%s|%s|%s", bookingID, residentID, reminderText)
	score := float64(scheduledAt.Unix())

	client := s.redis.GetClient()
	if err := client.ZAdd(ctx, key, redis.Z{Score: score, Member: reminderData}).Err(); err != nil {
		return fmt.Errorf("failed to schedule reminder: %w", err)
	}

	log.Printf("Scheduled reminder %s for booking %s at %s", reminderText, bookingID, scheduledAt)
	return nil
}

func (s *NotificationService) CancelRemindersForBooking(ctx context.Context, bookingID string) error {
	client := s.redis.GetClient()

	members, err := client.ZRangeByScore(ctx, "dorm:reminders", &redis.ZRangeBy{
		Min: "-inf",
		Max: "+inf",
	}).Result()
	if err != nil {
		return err
	}

	for _, member := range members {
		if len(member) > len(bookingID) && member[:len(bookingID)] == bookingID {
			if err := client.ZRem(ctx, "dorm:reminders", member).Err(); err != nil {
				return err
			}
			log.Printf("Cancelled reminder for booking %s", bookingID)
		}
	}

	return nil
}

func (s *NotificationService) GetPendingReminders(ctx context.Context) ([]string, error) {
	now := float64(timeutil.Now().Unix())

	members, err := s.redis.GetClient().ZRangeByScore(ctx, "dorm:reminders", &redis.ZRangeBy{
		Min: "-inf",
		Max: fmt.Sprintf("%f", now),
	}).Result()
	if err != nil {
		return nil, err
	}

	for _, member := range members {
		s.redis.GetClient().ZRem(ctx, "dorm:reminders", member)
	}

	return members, nil
}

func (s *NotificationService) NotifyAllResidents(ctx context.Context, dormitoryID string, message string) error {
	log.Printf("Notifying all residents of dormitory %s: %s", dormitoryID, message)
	return nil
}
