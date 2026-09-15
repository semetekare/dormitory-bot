package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTService — выпуск и проверка JWT токенов для Mini App API.
type JWTService struct {
	secret []byte
}

// Claims — стандартные claims + наши поля.
type Claims struct {
	jwt.RegisteredClaims
	UserID      string `json:"user_id"`
	PersonType  string `json:"person_type"`
	DormitoryID string `json:"dormitory_id,omitempty"`
	EmployeeID  string `json:"employee_id,omitempty"`
}

// NewJWTService создаёт сервис с переданным секретом.
func NewJWTService(secret string) *JWTService {
	if secret == "" {
		panic("JWT: secret must not be empty")
	}
	return &JWTService{secret: []byte(secret)}
}

// IssueToken выпускает JWT с переданными claims. TTL = 24h.
func (s *JWTService) IssueToken(userID, personType, dormitoryID, employeeID string) (string, error) {
	return s.IssueTokenWithTTL(userID, personType, dormitoryID, employeeID, 24*time.Hour)
}

// IssueTokenWithTTL выпускает JWT с произвольным TTL.
// В dev-режиме TTL = 30 дней, в prod — 24h (через IssueToken).
func (s *JWTService) IssueTokenWithTTL(userID, personType, dormitoryID, employeeID string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
		UserID:      userID,
		PersonType:  personType,
		DormitoryID: dormitoryID,
		EmployeeID:  employeeID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("jwt sign: %w", err)
	}
	return tokenStr, nil
}

// ValidateToken парсит и валидирует JWT, возвращает claims.
func (s *JWTService) ValidateToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("jwt parse: %w", err)
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}
