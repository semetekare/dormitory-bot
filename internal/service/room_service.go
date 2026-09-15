package service

import (
	"github.com/dormitory-bot/internal/domain"
	"github.com/dormitory-bot/internal/repository"
	"github.com/google/uuid"
)

type RoomService struct {
	roomRepo     *repository.BaseRepository[domain.Room]
	residentRepo *repository.BaseRepository[domain.Resident]
	materialRepo *repository.BaseRepository[domain.MaterialResponsibility]
}

func NewRoomService(
	roomRepo *repository.BaseRepository[domain.Room],
	residentRepo *repository.BaseRepository[domain.Resident],
	materialRepo *repository.BaseRepository[domain.MaterialResponsibility],
) *RoomService {
	return &RoomService{
		roomRepo:     roomRepo,
		residentRepo: residentRepo,
		materialRepo: materialRepo,
	}
}

func (s *RoomService) GetRoomItems(roomID uuid.UUID) ([]domain.MaterialResponsibility, error) {
	return s.materialRepo.FindAllByField("room_id", roomID)
}

func (s *RoomService) AddRoomItem(roomID uuid.UUID, itemName string, invNumber string, quantity int, condition string) (*domain.MaterialResponsibility, error) {
	item := &domain.MaterialResponsibility{
		RoomID:          roomID,
		ItemName:        itemName,
		InventoryNumber: invNumber,
		Quantity:        quantity,
		Condition:       condition,
	}
	if err := s.materialRepo.Create(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *RoomService) UpdateRoomItem(itemID uuid.UUID, itemName string, invNumber string, quantity int, condition string) error {
	item, err := s.materialRepo.GetByID(itemID)
	if err != nil {
		return err
	}
	item.ItemName = itemName
	item.InventoryNumber = invNumber
	item.Quantity = quantity
	item.Condition = condition
	return s.materialRepo.Update(item)
}

func (s *RoomService) DeleteRoomItem(itemID uuid.UUID) error {
	return s.materialRepo.Delete(itemID)
}

func (s *RoomService) GetRoomResidents(roomID uuid.UUID) ([]domain.Resident, error) {
	var residents []domain.Resident
	err := s.residentRepo.DB().Preload("User").Where("room_id = ?", roomID).Find(&residents).Error
	return residents, err
}

func (s *RoomService) GetRoomInfo(roomID uuid.UUID) (*domain.Room, error) {
	return s.roomRepo.GetByID(roomID)
}

func (s *RoomService) GetResident(residentID uuid.UUID) (*domain.Resident, error) {
	return s.residentRepo.GetByID(residentID)
}

// SearchRooms ищет комнаты по номеру (частичное совпадение).
func (s *RoomService) SearchRooms(query string) ([]domain.Room, error) {
	var rooms []domain.Room
	err := s.roomRepo.DB().
		Where("room_number ILIKE ?", "%"+query+"%").
		Limit(50).
		Find(&rooms).Error
	return rooms, err
}
