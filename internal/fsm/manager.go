package fsm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type UserContext struct {
	State           string                 `json:"state"`
	UserID          int64                  `json:"user_id"`
	Phone           string                 `json:"phone,omitempty"`
	FirstName       string                 `json:"first_name"`
	LastName        string                 `json:"last_name"`
	MiddleName      string                 `json:"middle_name,omitempty"`
	IsAuthorized    bool                   `json:"is_authorized"`
	PersonType      string                 `json:"person_type,omitempty"`
	EISVerified     bool                   `json:"eis_verified"`
	EISPersonID     string                 `json:"eis_person_id,omitempty"`
	UserUUID        string                 `json:"user_uuid,omitempty"`
	ResidentUUID    string                 `json:"resident_uuid,omitempty"`
	EmployeeUUID    string                 `json:"employee_uuid,omitempty"`
	DormitoryID     string                 `json:"dormitory_id,omitempty"`
	DormitoryName   string                 `json:"dormitory_name,omitempty"`
	RoomID          string                 `json:"room_id,omitempty"`
	RoomNumber      string                 `json:"room_number,omitempty"`
	FloorNumber     int                    `json:"floor_number,omitempty"`
	EmployeeRole    string                 `json:"employee_role,omitempty"`
	PreviousState   string                 `json:"previous_state,omitempty"`
	NavigationStack []string               `json:"navigation_stack,omitempty"`
	LastMessageID   string                 `json:"last_message_id,omitempty"`
	ContextData     map[string]interface{} `json:"context_data,omitempty"`
	StateExpiresAt  *time.Time             `json:"state_expires_at,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

type FSMManager struct {
	redis *redis.Client
	ttl   time.Duration
	mu    sync.Map
}

func NewFSMManager(redisClient *redis.Client) *FSMManager {
	return &FSMManager{
		redis: redisClient,
		ttl:   30 * 24 * time.Hour,
	}
}

func (m *FSMManager) fsmKey(userID int64) string {
	return fmt.Sprintf("dorm:fsm:%d", userID)
}

var stateTimeouts = map[string]time.Duration{}

func (m *FSMManager) SetStateWithTimeout(ctx context.Context, uc *UserContext, newState string) error {
	uc.PreviousState = uc.State
	uc.State = newState

	if duration, ok := stateTimeouts[newState]; ok {
		expiresAt := time.Now().Add(duration)
		uc.StateExpiresAt = &expiresAt
	} else {
		uc.StateExpiresAt = nil
	}

	return m.Set(ctx, uc)
}

func (m *FSMManager) getUserMutex(userID int64) *sync.Mutex {
	actual, _ := m.mu.LoadOrStore(userID, &sync.Mutex{})
	return actual.(*sync.Mutex)
}

func (m *FSMManager) Get(ctx context.Context, userID int64) (*UserContext, error) {
	return m.getInternal(ctx, userID)
}

func (m *FSMManager) getInternal(ctx context.Context, userID int64) (*UserContext, error) {
	key := m.fsmKey(userID)
	data, err := m.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return &UserContext{
			State:        StateUninitialized,
			UserID:       userID,
			IsAuthorized: false,
			ContextData:  make(map[string]interface{}),
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("redis get error: %w", err)
	}

	var uc UserContext
	if err := json.Unmarshal([]byte(data), &uc); err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}

	if uc.StateExpiresAt != nil && time.Now().After(*uc.StateExpiresAt) {
		uc.State = StateAuthorized
		uc.StateExpiresAt = nil
		uc.PreviousState = ""
		m.Set(context.Background(), &uc)
	}

	return &uc, nil
}

func (m *FSMManager) Set(ctx context.Context, uc *UserContext) error {
	if uc == nil {
		return fmt.Errorf("user context cannot be nil")
	}

	if uc.PreviousState != "" && uc.PreviousState != uc.State {
		ValidateStateTransition(uc.UserID, uc.PreviousState, uc.State)
	}

	uc.UpdatedAt = time.Now()
	data, err := json.Marshal(uc)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	key := m.fsmKey(uc.UserID)
	if err := m.redis.Set(ctx, key, data, m.ttl).Err(); err != nil {
		return fmt.Errorf("redis set error: %w", err)
	}

	return nil
}

func (m *FSMManager) UpdateFunc(ctx context.Context, userID int64, fn func(*UserContext) *UserContext) error {
	mu := m.getUserMutex(userID)
	mu.Lock()
	defer mu.Unlock()

	uc, err := m.getInternal(ctx, userID)
	if err != nil {
		return err
	}

	uc = fn(uc)
	return m.setInternal(ctx, uc)
}

func (m *FSMManager) setInternal(ctx context.Context, uc *UserContext) error {
	if uc == nil {
		return fmt.Errorf("user context cannot be nil")
	}

	uc.UpdatedAt = time.Now()
	data, err := json.Marshal(uc)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	key := m.fsmKey(uc.UserID)
	if err := m.redis.Set(ctx, key, data, m.ttl).Err(); err != nil {
		return fmt.Errorf("redis set error: %w", err)
	}

	return nil
}

func ValidateStateTransition(userID int64, fromState, toState string) {
	if fromState == "" || fromState == toState {
		return
	}
	if !IsValidTransition(fromState, toState) {
		log.Printf("WARNING: Invalid state transition for user %d: %s -> %s",
			userID, fromState, toState)
	}
}

func (uc *UserContext) PushState() {
	if len(uc.NavigationStack) == 0 || uc.NavigationStack[len(uc.NavigationStack)-1] != uc.State {
		if len(uc.NavigationStack) >= 10 {
			uc.NavigationStack = uc.NavigationStack[1:]
		}
		uc.NavigationStack = append(uc.NavigationStack, uc.State)
		uc.PreviousState = uc.State
	}
}

func (uc *UserContext) PopState() string {
	if len(uc.NavigationStack) == 0 {
		return StateAuthorized
	}

	prevState := uc.NavigationStack[len(uc.NavigationStack)-1]
	uc.NavigationStack = uc.NavigationStack[:len(uc.NavigationStack)-1]
	uc.PreviousState = prevState

	return prevState
}

func (uc *UserContext) ClearNavigationStack() {
	uc.NavigationStack = make([]string, 0)
	uc.PreviousState = ""
}

func (m *FSMManager) Delete(ctx context.Context, userID int64) error {
	key := m.fsmKey(userID)
	if err := m.redis.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("redis delete error: %w", err)
	}
	return nil
}
