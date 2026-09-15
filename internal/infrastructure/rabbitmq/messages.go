package rabbitmq

// ContractVersion is the current version of all message contracts.
const ContractVersion = 1

// ── Verify ──────────────────────────────────────────

type VerifyRequest struct {
	Version       int    `json:"version"`
	CorrelationID string `json:"correlation_id"`
	Phone         string `json:"phone"`
	MaxUserID     int64  `json:"max_user_id"`
	Platform      string `json:"platform"`
	Timestamp     string `json:"timestamp"`
}

type PersonInfo struct {
	PersonID   uint   `json:"person_id"`
	LastName   string `json:"last_name"`
	FirstName  string `json:"first_name"`
	MiddleName string `json:"middle_name"`
	Phone      string `json:"phone"`
}

type StudentInfo struct {
	BuildingID   uint   `json:"building_id"`
	BuildingName string `json:"building_name"`
	RoomName     string `json:"room_name"`
	RoomCapacity uint   `json:"room_capacity"`
	FloorID      uint   `json:"floor_id"`
	FloorName    string `json:"floor_name,omitempty"`
	StatusID     int    `json:"status_id"`
	StatusName   string `json:"status_name"`
}

type EmployeeInfo struct {
	EmployeeID   uint   `json:"employee_id"`
	PositionName string `json:"position_name"`
	PositionID   uint   `json:"position_id"`
	DepartmentID uint   `json:"department_id"`
	Department   string `json:"department"`
	BuildingID   *uint  `json:"building_id"`
}

type ContractInfo struct {
	Number    string `json:"number"`
	Status    int    `json:"status"`
	DateIn    string `json:"date_in,omitempty"`
	DateEnd   string `json:"date_end,omitempty"`
	NoDebt    *int   `json:"nodebt,omitempty"`
	Paused    *int   `json:"paused,omitempty"`
	PauseDate string `json:"pause_date,omitempty"`
}

type VerifyResponse struct {
	Version       int            `json:"version"`
	CorrelationID string         `json:"correlation_id"`
	Found         bool           `json:"found"`
	Person        *PersonInfo    `json:"person,omitempty"`
	Student       *StudentInfo   `json:"student,omitempty"`
	Employee      *EmployeeInfo  `json:"employee,omitempty"`
	Contract      *ContractInfo  `json:"contract,omitempty"`
	Error         string         `json:"error,omitempty"`
}

// ── Sync: Residents ─────────────────────────────────

type ResidentSyncEntry struct {
	PersonID    uint          `json:"person_id"`
	LastName    string        `json:"last_name"`
	FirstName   string        `json:"first_name"`
	MiddleName  string        `json:"middle_name"`
	Phone       string        `json:"phone"`
	RoomEISCode string        `json:"room_eis_code"`
	IsActive    bool          `json:"is_active"`
	Contract    *ContractInfo `json:"contract,omitempty"`
}

type SyncResidentsPush struct {
	Version           int                 `json:"version"`
	SyncID            string              `json:"sync_id"`
	UniversityID      string              `json:"university_id"`
	DormitoryEISCode  string              `json:"dormitory_eis_code"`
	Timestamp         string              `json:"timestamp"`
	Residents         []ResidentSyncEntry `json:"residents"`
	DeactivatedPhones []string            `json:"deactivated_phones"`
}

// ── Sync: Employees ─────────────────────────────────

type EmployeeSyncEntry struct {
	PersonID     uint   `json:"person_id"`
	LastName     string `json:"last_name"`
	FirstName    string `json:"first_name"`
	MiddleName   string `json:"middle_name"`
	Phone        string `json:"phone"`
	PositionName string `json:"position_name"`
	Department   string `json:"department"`
	Role         string `json:"role"`
}

type SyncEmployeesPush struct {
	Version          int                 `json:"version"`
	SyncID           string              `json:"sync_id"`
	UniversityID     string              `json:"university_id"`
	DormitoryEISCode string              `json:"dormitory_eis_code"`
	Timestamp        string              `json:"timestamp"`
	Employees        []EmployeeSyncEntry `json:"employees"`
}

// ── Sync: Dormitories ───────────────────────────────

type RoomSyncEntry struct {
	EISCode  string `json:"eis_code"`
	Name     string `json:"name"`
	Capacity uint   `json:"capacity"`
	Floor    int    `json:"floor"`
}

type DormitorySyncEntry struct {
	EISCode    string          `json:"eis_code"`
	Name       string          `json:"name"`
	BuildingID uint            `json:"building_id"`
	Floors     []int           `json:"floors"`
	Rooms      []RoomSyncEntry `json:"rooms"`
}

type SyncDormitoriesPush struct {
	Version      int                  `json:"version"`
	SyncID       string               `json:"sync_id"`
	UniversityID string               `json:"university_id"`
	Timestamp    string               `json:"timestamp"`
	Dormitories  []DormitorySyncEntry `json:"dormitories"`
}

// ── SyncTrigger (bot → adapter) ────────────────────

type SyncTrigger struct {
	Version   int    `json:"version"`
	TriggerID string `json:"trigger_id"`
	// MySQL connection parameters for the EIS database
	MySQLHost     string `json:"mysql_host"`
	MySQLPort     int    `json:"mysql_port"`
	MySQLUser     string `json:"mysql_user"`
	MySQLPassword string `json:"mysql_password"`
	MySQLDBName   string `json:"mysql_dbname"`
}

func (r *SyncTrigger) ValidateVersion() bool { return r.Version == ContractVersion }

// ── Helpers ─────────────────────────────────────────

func (r *VerifyResponse) ValidateVersion() bool      { return r.Version == ContractVersion }
func (r *SyncResidentsPush) ValidateVersion() bool    { return r.Version == ContractVersion }
func (r *SyncEmployeesPush) ValidateVersion() bool    { return r.Version == ContractVersion }
func (r *SyncDormitoriesPush) ValidateVersion() bool  { return r.Version == ContractVersion }
