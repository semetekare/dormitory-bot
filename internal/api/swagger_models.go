// Package api — Swagger response models.
// Each type is referenced by @Success annotations in router.go.
package api

// AuthVerifyResponse — POST /auth/verify
type AuthVerifyResponse struct {
	LaunchToken string    `json:"launch_token" example:"abc-uuid-token"`
	User        AuthUser `json:"user"`
}
type AuthUser struct {
	UserID     string `json:"user_id" example:"abc6efcd-487b-4b42-a30c-3dbb0a60cfdd"`
	PersonType string `json:"person_type" example:"student"`
	FirstName  string `json:"first_name" example:"Артём"`
	Phone      string `json:"phone" example:"79001234501"`
}

// AuthMiniappResponse — POST /auth/miniapp
type AuthMiniappResponse struct {
	Token string      `json:"token" example:"eyJhbG..."`
	User  AuthMiniUser `json:"user"`
}
type AuthMiniUser struct {
	UserID     string `json:"user_id" example:"abc6efcd-487b-4b42-a30c-3dbb0a60cfdd"`
	PersonType string `json:"person_type" example:"student"`
}

// AuthMeResponse — GET /auth/me
type AuthMeResponse struct {
	UserID     string `json:"user_id" example:"abc6efcd-487b-4b42-a30c-3dbb0a60cfdd"`
	PersonType string `json:"person_type" example:"student"`
	FirstName  string `json:"first_name" example:"Артём"`
	LastName   string `json:"last_name" example:"Иванов"`
	MiddleName string `json:"middle_name" example:"Сергеевич"`

	// Информация о договоре и общежитии (только для жителя)
	ContractNumber    *string  `json:"contract_number,omitempty"`
	ContractStartDate *string  `json:"contract_start_date,omitempty"`
	ContractEndDate   *string  `json:"contract_end_date,omitempty"`
	Dormitory         *string  `json:"dormitory,omitempty" example:"Общежитие №8"`
	RoomNumber        *string  `json:"room_number,omitempty"`
}

// ── Dormitories ──

type DormitoryListResponse struct {
	Data []DormitoryItem `json:"data"`
}
type DormitoryItem struct {
	ID               string `json:"id" example:"b0000000-0000-0000-0000-000000000001"`
	Name             string `json:"name" example:"Общежитие №8"`
	Address          string `json:"address" example:"ул. Лермонтова, д. 80"`
	EISDormitoryCode string `json:"eis_dormitory_code" example:"EIS-BLD-8"`
	IsActive         bool   `json:"is_active" example:"true"`
}

type FloorListResponse struct {
	Data []FloorItem `json:"data"`
}
type FloorItem struct {
	ID          string `json:"id" example:"b0000000-0000-0000-0000-000000010001"`
	DormitoryID string `json:"dormitory_id" example:"b0000000-0000-0000-0000-000000000001"`
	FloorNumber int    `json:"floor_number" example:"1"`
}

type StaffListResponse struct {
	Data []StaffItem `json:"data"`
}
type StaffItem struct {
	UserID      string `json:"user_id" example:"abc6efcd-487b-4b42-a30c-3dbb0a60cfdd"`
	FirstName   string `json:"first_name" example:"Данил"`
	LastName    string `json:"last_name" example:"Высоких"`
	Phone       string `json:"phone" example:"79025643215"`
	Role        string `json:"role" example:"commandant"`
	DisplayName string `json:"display_name" example:"Заведующий"`
}

// ── Rooms ──

type RoomListResponse struct {
	Data []RoomItem `json:"data"`
}
type RoomItem struct {
	ID          string `json:"id" example:"b0000000-0000-0000-0000-000000030001"`
	DormitoryID string `json:"dormitory_id" example:"b0000000-0000-0000-0000-000000000001"`
	FloorID     string `json:"floor_id" example:"b0000000-0000-0000-0000-000000010001"`
	RoomNumber  string `json:"room_number" example:"101"`
	Capacity    int    `json:"capacity" example:"3"`
	IsActive    bool   `json:"is_active" example:"true"`
}

type MaterialItemListResponse struct {
	Data []MaterialItem `json:"data"`
}
type MaterialItemResponse struct {
	Data MaterialItem `json:"data"`
}
type MaterialItem struct {
	ID              string `json:"id" example:"b0000000-0000-0000-0000-000000040001"`
	RoomID          string `json:"room_id" example:"b0000000-0000-0000-0000-000000030001"`
	ItemName        string `json:"item_name" example:"Кровать"`
	InventoryNumber string `json:"inventory_number" example:"MOCK-INV-101-001"`
	Quantity        int    `json:"quantity" example:"3"`
	Condition       string `json:"condition" example:"хорошее"`
}

type RoomResidentsResponse struct {
	Residents []RoomResidentItem `json:"residents"`
}
type RoomResidentItem struct {
	ID        string `json:"id" example:"resident-uuid"`
	UserID    string `json:"user_id" example:"abc6efcd-487b-4b42-a30c-3dbb0a60cfdd"`
	RoomID    string `json:"room_id" example:"b0000000-0000-0000-0000-000000030001"`
	FirstName string `json:"first_name" example:"Артём"`
	LastName  string `json:"last_name" example:"Иванов"`
	Phone     string `json:"phone" example:"79001234501"`
}

// ── Roles ──

type RoleListResponse struct {
	Data []RoleItem `json:"data"`
}
type RoleItem struct {
	ID          string `json:"id" example:"r-commandant"`
	Name        string `json:"name" example:"commandant"`
	DisplayName string `json:"display_name" example:"Заведующий"`
	TargetType  string `json:"target_type" example:"employee"`
	Priority    int    `json:"priority" example:"100"`
	IsSystem    bool   `json:"is_system" example:"true"`
}

// ── Residents ──

type ResidentProfileResponse struct {
	Data ResidentProfile `json:"data"`
}
type ResidentProfile struct {
	ID            string `json:"id" example:"resident-uuid"`
	UserID        string `json:"user_id" example:"abc6efcd-487b-4b42-a30c-3dbb0a60cfdd"`
	RoomID        string `json:"room_id" example:"b0000000-0000-0000-0000-000000030001"`
	DormitoryID   string `json:"dormitory_id" example:"b0000000-0000-0000-0000-000000000001"`
	ContractNum   string `json:"contract_num" example:"K-2024-001"`
	ContractStart string `json:"contract_start" example:"2024-09-01"`
	ContractEnd   string `json:"contract_end" example:"2025-06-30"`
}

type ResidentDebtResponse struct {
	Data []ResidentDebt `json:"data"`
}
type ResidentDebt struct {
	ID          string  `json:"id" example:"debt-uuid"`
	ResidentID  string  `json:"resident_id" example:"resident-uuid"`
	Amount      float64 `json:"amount" example:"1500.00"`
	Period      string  `json:"period" example:"2025-01"`
	Description string  `json:"description" example:"Задолженность за проживание"`
}

// ── Employees ──

type EmployeeInfoResponse struct {
	Data EmployeeInfo `json:"data"`
}
type EmployeeInfo struct {
	ID                 string `json:"id" example:"eb008023-1c16-4d89-a9ef-aeea6b7be792"`
	UserID             string `json:"user_id" example:"abc6efcd-487b-4b42-a30c-3dbb0a60cfdd"`
	Position           string `json:"position" example:"Заведующий общежитием"`
	Role               string `json:"role" example:"commandant"`
	PrimaryDormitoryID string `json:"primary_dormitory_id" example:"b0000000-0000-0000-0000-000000000001"`
}

type EmployeeDormitoriesResponse struct {
	Data []EmployeeDormitory `json:"data"`
}
type EmployeeDormitory struct {
	DormitoryID   string `json:"dormitory_id" example:"b0000000-0000-0000-0000-000000000001"`
	DormitoryName string `json:"dormitory_name" example:"Общежитие №8"`
	Role          string `json:"role" example:"commandant"`
}

// ── Laundry ──

type LaundrySlotListResponse struct {
	Data []LaundrySlot `json:"data"`
}
type LaundrySlot struct {
	MachineID     string `json:"machine_id" example:"b0000000-0000-0000-0000-000000200001"`
	MachineNumber string `json:"machine_number" example:"1"`
	SlotStart     string `json:"slot_start" example:"07:00"`
	SlotEnd       string `json:"slot_end" example:"08:00"`
}

type WashingMachineListResponse struct {
	Data []WashingMachineItem `json:"data"`
}
type WashingMachineItem struct {
	ID              string `json:"id" example:"b0000000-0000-0000-0000-000000200001"`
	DormitoryID     string `json:"dormitory_id" example:"b0000000-0000-0000-0000-000000000001"`
	FloorID         string `json:"floor_id" example:"b0000000-0000-0000-0000-000000010001"`
	MachineNumber   string `json:"machine_number" example:"1"`
	IsActive        bool   `json:"is_active" example:"true"`
	DurationMinutes int    `json:"duration_minutes" example:"60"`
}

type LaundryBookingResponse struct {
	Data LaundryBooking `json:"data"`
}
type LaundryBooking struct {
	BookingID     string `json:"booking_id" example:"booking-uuid"`
	MachineNumber string `json:"machine_number" example:"1"`
	SlotDate      string `json:"slot_date" example:"2026-08-06"`
	SlotStart     string `json:"slot_start" example:"10:00"`
	SlotEnd       string `json:"slot_end" example:"11:00"`
}

type LaundrySettingsResponse struct {
	Data LaundrySettingsItem `json:"data"`
}
type LaundrySettingsItem struct {
	ID                     string `json:"id" example:"b0000000-0000-0000-0000-000000300001"`
	DormitoryID            string `json:"dormitory_id" example:"b0000000-0000-0000-0000-000000000001"`
	BookingStartTime       string `json:"booking_start_time" example:"07:00:00"`
	BookingEndTime         string `json:"booking_end_time" example:"23:00:00"`
	AdvanceBookingEnabled  bool   `json:"advance_booking_enabled" example:"false"`
	AdvanceBookingMaxDays  int    `json:"advance_booking_max_days" example:"1"`
	DefaultDurationMinutes int    `json:"default_duration_minutes" example:"60"`
}

// ── Cleaning ──

type CleaningScheduleResponse struct {
	Data []CleaningDuty `json:"data"`
}
type CleaningDuty struct {
	ID          string `json:"id" example:"duty-uuid"`
	DormitoryID string `json:"dormitory_id" example:"b0000000-0000-0000-0000-000000000001"`
	RoomID      string `json:"room_id" example:"b0000000-0000-0000-0000-000000030001"`
	RoomNumber  string `json:"room_number" example:"101"`
	DutyDate    string `json:"duty_date" example:"2026-08-15"`
	Status      string `json:"status" example:"pending"`
}

type CleaningPenaltyResponse struct {
	Data []CleaningPenalty `json:"data"`
}
type CleaningPenalty struct {
	ID     string `json:"id" example:"penalty-uuid"`
	RoomID string `json:"room_id" example:"b0000000-0000-0000-0000-000000030001"`
	Count  int    `json:"count" example:"1"`
	Reason string `json:"reason" example:"Невыполнение дежурства"`
	Status string `json:"status" example:"pending"`
}

type CleaningExemptionResponse struct {
	Data []CleaningExemption `json:"data"`
}
type CleaningExemption struct {
	ID        string `json:"id" example:"exemption-uuid"`
	RoomID    string `json:"room_id" example:"b0000000-0000-0000-0000-000000030001"`
	StartDate string `json:"start_date" example:"2026-08-01"`
	EndDate   string `json:"end_date" example:"2026-08-07"`
	Reason    string `json:"reason" example:"Ремонт"`
}

// ── References ──

type ReferenceMaterialListResponse struct {
	Data []ReferenceMaterialItem `json:"data"`
}
type ReferenceMaterialItem struct {
	ID          string `json:"id" example:"b0000000-0000-0000-0000-000000600001"`
	DormitoryID string `json:"dormitory_id" example:"b0000000-0000-0000-0000-000000000001"`
	Name        string `json:"name" example:"Правила проживания"`
	Description string `json:"description" example:"Правила внутреннего распорядка."`
	Category    string `json:"category" example:"Правила"`
	Ordinal     int    `json:"ordinal" example:"1"`
}

// ── Chat Links ──

type ChatLinkListResponse struct {
	Data []ChatLinkItem `json:"data"`
}
type ChatLinkItem struct {
	ID          string `json:"id" example:"b0000000-0000-0000-0000-000000500001"`
	DormitoryID string `json:"dormitory_id" example:"b0000000-0000-0000-0000-000000000001"`
	FloorID     string `json:"floor_id,omitempty" example:"b0000000-0000-0000-0000-000000010001"`
	LinkType    string `json:"link_type" example:"dormitory"`
	Platform    string `json:"platform" example:"telegram"`
	URL         string `json:"url" example:"https://t.me/dorm8_mock"`
	Title       string `json:"title" example:"Чат общежития №8"`
	IsActive    bool   `json:"is_active" example:"true"`
}

// ── Common ──

type ErrorResponse struct {
	Error   string `json:"error" example:"dormitory not found"`
	Success bool   `json:"success" example:"false"`
	Code    int    `json:"code" example:"404"`
}

type MessageResponse struct {
	Message string `json:"message" example:"booking cancelled"`
}
