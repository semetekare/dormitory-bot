package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/dormitory-bot/internal/domain"
	"github.com/dormitory-bot/internal/repository"
	"github.com/google/uuid"
)

type BotModuleService struct {
	repo *repository.BaseRepository[domain.BotModule]
}

func NewBotModuleService(repo *repository.BaseRepository[domain.BotModule]) *BotModuleService {
	return &BotModuleService{repo: repo}
}

func (s *BotModuleService) GetByDormitory(dormitoryID uuid.UUID) ([]domain.BotModule, error) {
	return s.repo.FindAllByField("dormitory_id", dormitoryID)
}

func (s *BotModuleService) Create(dormitoryID uuid.UUID, platform string, token string, name string) (*domain.BotModule, error) {
	encToken, err := encryptToken(token)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt token: %w", err)
	}

	existing, _ := s.repo.FindByField("dormitory_id", dormitoryID)
	_ = existing

	module := &domain.BotModule{
		DormitoryID: dormitoryID,
		Platform:    platform,
		BotToken:    encToken,
		BotName:     name,
		Status:      "stopped",
	}
	if err := s.repo.Create(module); err != nil {
		return nil, err
	}
	return module, nil
}

func (s *BotModuleService) UpdateToken(id uuid.UUID, token string) error {
	encToken, err := encryptToken(token)
	if err != nil {
		return err
	}
	return s.repo.DB().Model(&domain.BotModule{}).Where("id = ?", id).Update("bot_token", encToken).Error
}

func (s *BotModuleService) UpdateStatus(id uuid.UUID, status string) error {
	updates := map[string]interface{}{"status": status}
	now := time.Now()
	if status == "running" {
		updates["last_started_at"] = now
	} else if status == "stopped" {
		updates["last_stopped_at"] = now
	}
	if status == "error" {
		updates["error_message"] = "Unknown error"
	}
	return s.repo.DB().Model(&domain.BotModule{}).Where("id = ?", id).Updates(updates).Error
}

func (s *BotModuleService) Delete(id uuid.UUID) error {
	return s.repo.Delete(id)
}

func (s *BotModuleService) GetByID(id uuid.UUID) (*domain.BotModule, error) {
	return s.repo.GetByID(id)
}

func encryptToken(token string) (string, error) {
	keyStr := os.Getenv("ENCRYPTION_KEY")
	if keyStr == "" {
		keyStr = "0123456789abcdef0123456789abcdef"
	}
	key := []byte(keyStr)
	if len(key) != 32 {
		return "", fmt.Errorf("encryption key must be 32 bytes")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(token), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}
