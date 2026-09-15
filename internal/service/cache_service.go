package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dormitory-bot/internal/infrastructure/cache"
	"github.com/google/uuid"
)

type CacheService struct {
	redisClient *cache.RedisClient
	defaultTTL  time.Duration
}

func NewCacheService(redisClient *cache.RedisClient) *CacheService {
	return &CacheService{
		redisClient: redisClient,
		defaultTTL:  5 * time.Minute,
	}
}

func (s *CacheService) Get(ctx context.Context, key string, dest interface{}) (bool, error) {
	val, err := s.redisClient.Get(ctx, key)
	if err != nil {
		return false, fmt.Errorf("failed to get from cache: %w", err)
	}
	if val == "" {
		return false, nil
	}
	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return false, fmt.Errorf("failed to unmarshal cache value: %w", err)
	}
	return true, nil
}

func (s *CacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = s.defaultTTL
	}
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value for cache: %w", err)
	}
	if err := s.redisClient.Set(ctx, key, string(data), ttl); err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}
	return nil
}

func (s *CacheService) Delete(ctx context.Context, key string) error {
	if err := s.redisClient.Delete(ctx, key); err != nil {
		return fmt.Errorf("failed to delete from cache: %w", err)
	}
	return nil
}

func (s *CacheService) DeleteByPattern(ctx context.Context, pattern string) error {
	fmt.Printf("Cache invalidation pattern: %s\n", pattern)
	return s.redisClient.DeleteByPattern(ctx, pattern)
}

func (s *CacheService) GetCacheKey(resource string, id uuid.UUID) string {
	return fmt.Sprintf("dorm:%s:%s", resource, id.String())
}

func (s *CacheService) GetListCacheKey(resource string, filters map[string]string) string {
	hash := ""
	for k, v := range filters {
		hash += fmt.Sprintf("%s=%s:", k, v)
	}
	return fmt.Sprintf("dorm:%s:list:%s", resource, hash)
}
