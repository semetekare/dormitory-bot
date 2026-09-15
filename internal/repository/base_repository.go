package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BaseRepository[T any] struct {
	db *gorm.DB
}

func NewBaseRepository[T any](db *gorm.DB) *BaseRepository[T] {
	return &BaseRepository[T]{db: db}
}

func (r *BaseRepository[T]) DB() *gorm.DB {
	return r.db
}

func (r *BaseRepository[T]) Create(entity *T) error {
	return r.db.Create(entity).Error
}

func (r *BaseRepository[T]) GetByID(id uuid.UUID) (*T, error) {
	var entity T
	if err := r.db.Where("id = ?", id).First(&entity).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *BaseRepository[T]) Update(entity *T) error {
	return r.db.Save(entity).Error
}

func (r *BaseRepository[T]) Delete(id uuid.UUID) error {
	var entity T
	return r.db.Where("id = ?", id).Delete(&entity).Error
}

func (r *BaseRepository[T]) List(page, pageSize int, filters map[string]interface{}) ([]T, int64, error) {
	var entities []T
	var total int64

	query := r.db.Model(new(T))
	for key, value := range filters {
		query = query.Where(key, value)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return entities, total, nil
}

func (r *BaseRepository[T]) FindByField(field string, value interface{}) (*T, error) {
	var entity T
	if err := r.db.Where(field+" = ?", value).First(&entity).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *BaseRepository[T]) FindAllByField(field string, value interface{}) ([]T, error) {
	var entities []T
	if err := r.db.Where(field+" = ?", value).Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}
