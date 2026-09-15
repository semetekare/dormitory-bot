package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/dormitory-bot/internal/auth"
	"github.com/dormitory-bot/internal/domain"
	"github.com/dormitory-bot/internal/infrastructure/cache"
	"github.com/dormitory-bot/internal/infrastructure/rabbitmq"
	"github.com/dormitory-bot/internal/mock"
	"github.com/dormitory-bot/internal/repository"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Dependencies struct {
	DB          *gorm.DB
	RmqClient   *rabbitmq.EISClient
	Redis       *cache.RedisClient
	EISProvider mock.EISDataProvider // nil в production
}

type DBService struct {
	db           *gorm.DB
	eisProvider  mock.EISDataProvider
	UniBroker    *UniversityBroker
	cacheService *CacheService
	txManager    *TransactionManager
	rbac         *RBACService
	Guards       *Guards

	userRepo                  *repository.BaseRepository[domain.User]
	residentRepo              *repository.BaseRepository[domain.Resident]
	employeeRepo              *repository.BaseRepository[domain.Employee]
	employeeDormitoryRoleRepo *repository.EmployeeDormitoryRoleRepository
	userDormitoryRoleRepo     *repository.UserDormitoryRoleRepository
	dormitoryRepo             *repository.BaseRepository[domain.Dormitory]
	floorRepo                 *repository.BaseRepository[domain.Floor]
	roomRepo                  *repository.BaseRepository[domain.Room]
	roleRepo                  *repository.BaseRepository[domain.Role]

	washingMachineRepo        *repository.BaseRepository[domain.WashingMachine]
	laundryBookingRepo        *repository.LaundryBookingRepository
	laundrySettingsRepo       *repository.BaseRepository[domain.LaundrySettings]
	materialRepo              *repository.BaseRepository[domain.MaterialResponsibility]
	referenceRepo             *repository.BaseRepository[domain.ReferenceMaterial]
	chatLinkRepo              *repository.BaseRepository[domain.ChatLink]
	cleaningDutyRepo          *repository.BaseRepository[domain.CleaningDuty]
	penaltyCleaningRepo       *repository.BaseRepository[domain.PenaltyCleaning]
	cleaningExemptionRepo     *repository.BaseRepository[domain.CleaningExemption]
	sanitaryControlRepo       *repository.BaseRepository[domain.SanitaryControl]
	disciplineControlRepo     *repository.BaseRepository[domain.DisciplineControl]
	cohabitationControlRepo   *repository.BaseRepository[domain.CohabitationControl]
	residentFinancialDebtRepo *repository.BaseRepository[domain.ResidentFinancialDebt]

	LaundryService      *LaundryService
	RoomService         *RoomService
	ReferenceService    *ReferenceService
	ChatLinkService     *ChatLinkService
	CleaningService     *CleaningService
	BotModuleService    *BotModuleService
	ControlService      *ControlService
	ResidentService     *ResidentService
	NotificationService *NotificationService


}
// GetDormitoryRepo exposes the dormitory repository for API handlers.
func (s *DBService) GetDormitoryRepo() *repository.BaseRepository[domain.Dormitory] {
	return s.dormitoryRepo
}

// GetReferenceRepo exposes the reference materials repository for API handlers.
func (s *DBService) GetReferenceRepo() *repository.BaseRepository[domain.ReferenceMaterial] {
	return s.referenceRepo
}

// GetChatLinkRepo exposes the chat link repository for API handlers.
func (s *DBService) GetChatLinkRepo() *repository.BaseRepository[domain.ChatLink] {
	return s.chatLinkRepo
}

func New(deps *Dependencies) *DBService {
	svc := &DBService{
		db:                        deps.DB,
		eisProvider:               deps.EISProvider,
		txManager:                 NewTransactionManager(deps.DB),
		userRepo:                  repository.NewBaseRepository[domain.User](deps.DB),
		residentRepo:              repository.NewBaseRepository[domain.Resident](deps.DB),
		employeeRepo:              repository.NewBaseRepository[domain.Employee](deps.DB),
		employeeDormitoryRoleRepo: repository.NewEmployeeDormitoryRoleRepository(deps.DB),
		userDormitoryRoleRepo:     repository.NewUserDormitoryRoleRepository(deps.DB),
		dormitoryRepo:             repository.NewBaseRepository[domain.Dormitory](deps.DB),
		floorRepo:                 repository.NewBaseRepository[domain.Floor](deps.DB),
		roomRepo:                  repository.NewBaseRepository[domain.Room](deps.DB),
		roleRepo:                  repository.NewBaseRepository[domain.Role](deps.DB),

		washingMachineRepo:        repository.NewBaseRepository[domain.WashingMachine](deps.DB),
		laundryBookingRepo:        repository.NewLaundryBookingRepository(deps.DB),
		laundrySettingsRepo:       repository.NewBaseRepository[domain.LaundrySettings](deps.DB),
		materialRepo:              repository.NewBaseRepository[domain.MaterialResponsibility](deps.DB),
		referenceRepo:             repository.NewBaseRepository[domain.ReferenceMaterial](deps.DB),
		chatLinkRepo:              repository.NewBaseRepository[domain.ChatLink](deps.DB),
		cleaningDutyRepo:          repository.NewBaseRepository[domain.CleaningDuty](deps.DB),
		penaltyCleaningRepo:       repository.NewBaseRepository[domain.PenaltyCleaning](deps.DB),
		cleaningExemptionRepo:     repository.NewBaseRepository[domain.CleaningExemption](deps.DB),
		sanitaryControlRepo:       repository.NewBaseRepository[domain.SanitaryControl](deps.DB),
		disciplineControlRepo:     repository.NewBaseRepository[domain.DisciplineControl](deps.DB),
		cohabitationControlRepo:   repository.NewBaseRepository[domain.CohabitationControl](deps.DB),
		residentFinancialDebtRepo: repository.NewBaseRepository[domain.ResidentFinancialDebt](deps.DB),
	}

	svc.LaundryService = NewLaundryService(svc.washingMachineRepo, svc.laundryBookingRepo, svc.laundrySettingsRepo, svc.txManager)
	svc.RoomService = NewRoomService(svc.roomRepo, svc.residentRepo, svc.materialRepo)
	svc.ReferenceService = NewReferenceService(svc.referenceRepo)
	svc.ChatLinkService = NewChatLinkService(svc.chatLinkRepo)
	svc.CleaningService = NewCleaningService(svc.cleaningDutyRepo, svc.penaltyCleaningRepo, svc.cleaningExemptionRepo, svc.roomRepo, svc.residentRepo, svc.txManager)
	if deps.Redis != nil {
		svc.CleaningService.SetRedisClient(deps.Redis)
	}
	svc.BotModuleService = NewBotModuleService(repository.NewBaseRepository[domain.BotModule](deps.DB))
	svc.ControlService = NewControlService(svc.sanitaryControlRepo, svc.disciplineControlRepo, svc.cohabitationControlRepo)
	svc.ResidentService = NewResidentService(svc.residentRepo, svc.userRepo, svc.roomRepo, svc.dormitoryRepo, svc.floorRepo, svc.residentFinancialDebtRepo, svc.cleaningDutyRepo, svc.ControlService)
	svc.NotificationService = NewNotificationService(deps.Redis)

	if deps.Redis != nil {
		svc.cacheService = NewCacheService(deps.Redis)
	}
	svc.UniBroker = NewUniversityBroker(
		deps.RmqClient,
		deps.Redis,
		svc.userRepo,
		svc.residentRepo,
		svc.employeeRepo,
		svc.dormitoryRepo,
		svc.roomRepo,
		svc.floorRepo,
		svc.employeeDormitoryRoleRepo,
	)

	svc.rbac = NewRBACService(svc.roleRepo, svc.employeeRepo, svc.employeeDormitoryRoleRepo, svc.userDormitoryRoleRepo, svc)
	svc.Guards = NewGuards(svc)

	return svc
}

type VerifyUserResult struct {
	Found         bool
	UserID        string
	FirstName     string
	LastName      string
	MiddleName    string
	EISVerified   bool
	EISPersonID   string
	PersonType    string
	DormitoryID   string
	DormitoryName string
	Resident      *ResidentInfo
	Employee      *EmployeeFullInfo
}

type ResidentInfo struct {
	ResidentID     string
	DormitoryID    string
	DormitoryName  string
	RoomID         string
	RoomNumber     string
	FloorNumber    int
	ContractNumber string
	ContractStart  string
	ContractEnd    string
	ContractStatus int
}

type EmployeeFullInfo struct {
	EmployeeID  string
	Role        string
	RoleDisplay string
	Position    string
}

func (s *DBService) VerifyUser(phone string, maxUserID int64, platform string) (*VerifyUserResult, error) {
	existingUser, err := s.userRepo.FindByField("phone", phone)
	if err == nil && existingUser != nil {

		if platform == "max" && maxUserID > 0 {
			realID := fmt.Sprintf("%d", maxUserID)
			if existingUser.PlatformUserID != realID {
				oldID := existingUser.PlatformUserID
				existingUser.PlatformUserID = realID
				existingUser.Platform = platform
				s.userRepo.Update(existingUser)
				log.Printf("Updated PlatformUserID for user %s: %s → %s", existingUser.ID, oldID, realID)
			}
		}

		result := &VerifyUserResult{
			Found:       true,
			UserID:      existingUser.ID.String(),
			FirstName:   existingUser.FirstName,
			LastName:    existingUser.LastName,
			MiddleName:  existingUser.MiddleName,
			EISVerified: existingUser.EISVerified,
			EISPersonID: existingUser.EISPersonID,
			PersonType:  existingUser.PersonType,
		}

		if existingUser.PersonType == "student" {
			resident, err := s.residentRepo.FindByField("user_id", existingUser.ID)
			if err == nil && resident != nil {
				if resident.IsActive {
					result.Resident = s.buildResidentInfo(resident)
				} else {
					log.Printf("User %d: resident found but INACTIVE (dorm=%s, room=%s)",
						maxUserID, resident.DormitoryID, resident.RoomID)
				}
			} else {
				log.Printf("User %d: PersonType=student but no resident record found (user_id=%s)",
					maxUserID, existingUser.ID)
			}
		} else if existingUser.PersonType == "employee" {
			employee, err := s.employeeRepo.FindByField("user_id", existingUser.ID)
			if err == nil && employee != nil {
				dormitories, _ := s.GetEmployeeDormitories(employee.ID.String())
				if len(dormitories) > 0 {
					result.Employee = s.buildEmployeeInfo(employee)
					if len(dormitories) == 1 {
						result.DormitoryID = dormitories[0].DormitoryID
					}
				} else {
					log.Printf("User %d: PersonType=employee but no dormitory roles found (employee_id=%s)",
						maxUserID, employee.ID)
					resident, err := s.residentRepo.FindByField("user_id", existingUser.ID)
					if err == nil && resident != nil {
						if resident.IsActive {
							result.Resident = s.buildResidentInfo(resident)
						} else {
							log.Printf("User %d: employee fallback resident INACTIVE (dorm=%s, room=%s)",
								maxUserID, resident.DormitoryID, resident.RoomID)
						}
					}
				}
			}
		}

		return result, nil
	}

	log.Printf("Stage 2: Looking up via EIS RPC for phone %s", phone)

	if s.UniBroker != nil && s.UniBroker.IsRMQConnected() {
		person, employee, student, contract, err := s.UniBroker.VerifyUserRPC(context.Background(), phone)
		if err != nil {
			log.Printf("UniBroker RPC verify error: %v", err)
			return &VerifyUserResult{Found: false}, nil
		}
		if person == nil {
			return &VerifyUserResult{Found: false}, nil
		}

		return s.createUserFromEIS(person, employee, student, contract, platform, maxUserID)
	}

	// Stage 2.5: Mock provider fallback
	if s.eisProvider != nil {
		log.Printf("Stage 2.5: Falling back to mock EIS provider for phone %s", phone)
		person, employee, student, contract, mockErr := s.eisProvider.VerifyPerson(context.Background(), phone)
		if mockErr != nil {
			log.Printf("Mock provider error: %v", mockErr)
		}
		if person != nil {
			// Convert mock types to rabbitmq types for shared create* methods
			return s.createUserFromEIS(
				person.ToRabbitMQ(),
				employee.ToRabbitMQ(),
				student.ToRabbitMQ(),
				contract.ToRabbitMQ(),
				platform, maxUserID)
		}
	}

	return &VerifyUserResult{Found: false}, nil
}

// createUserFromEIS — общий код создания user/resident/employee из данных ЕИС.
// Используется и RPC-веткой, и mock-веткой.
func (s *DBService) createUserFromEIS(person *rabbitmq.PersonInfo, employee *rabbitmq.EmployeeInfo, student *rabbitmq.StudentInfo, contract *rabbitmq.ContractInfo, platform string, maxUserID int64) (*VerifyUserResult, error) {
	normalizedPhone := NormalizePhone(person.Phone)
	user := &domain.User{
		FirstName:      person.FirstName,
		LastName:       person.LastName,
		MiddleName:     person.MiddleName,
		Phone:          normalizedPhone,
		Platform:       platform,
		PlatformUserID: fmt.Sprintf("%d", maxUserID),
		EISVerified:    true,
		EISPersonID:    fmt.Sprintf("%d", person.PersonID),
	}
	if employee != nil {
		user.PersonType = "employee"
	} else if student != nil {
		user.PersonType = "student"
	}
	if err := s.userRepo.Create(user); err != nil {
		log.Printf("Failed to create user from EIS: %v", err)
		return &VerifyUserResult{Found: false}, nil
	}

	result := &VerifyUserResult{
		Found:       true,
		UserID:      user.ID.String(),
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		MiddleName:  user.MiddleName,
		EISVerified: true,
		EISPersonID: user.EISPersonID,
		PersonType:  user.PersonType,
	}

	s.createResidentFromEIS(student, contract, user.ID, result)

	if employee != nil {
		s.createEmployeeFromEIS(employee, user.ID, result)
	}

	return result, nil
}

// createResidentFromEIS создаёт Resident/Room/Floor из данных студента.
func (s *DBService) createResidentFromEIS(student *rabbitmq.StudentInfo, contract *rabbitmq.ContractInfo, userID uuid.UUID, result *VerifyUserResult) {
	if student == nil {
		return
	}
	ctx := context.Background()
	dormID := s.ensureDormitoryFromName(ctx, student.BuildingName, student.BuildingID)
	roomName := student.RoomName
	if roomName == "" {
		roomName = fmt.Sprintf("Room-%s", userID)
	}
	floorNum := extractFloorFromRoomName(student.RoomName)
	if floorNum < 1 {
		floorNum = 1
	}
	floors, _ := s.floorRepo.FindAllByField("dormitory_id", dormID)
	var floorID uuid.UUID
	for _, f := range floors {
		if f.FloorNumber == floorNum {
			floorID = f.ID
			break
		}
	}
	if floorID == uuid.Nil {
		fl := &domain.Floor{DormitoryID: dormID, FloorNumber: floorNum}
		s.floorRepo.Create(fl)
		floorID = fl.ID
	}
	// Find the room by dormitory_id + room_number (multi-field query via raw GORM).
	// FindByField only supports single-field queries; roomRepo.DB() gives direct GORM access.
	var room domain.Room
	errDB := s.roomRepo.DB().Where("dormitory_id = ? AND room_number = ?", dormID, roomName).First(&room).Error
	if errors.Is(errDB, gorm.ErrRecordNotFound) {
		room = domain.Room{
			DormitoryID: dormID,
			FloorID:     floorID,
			RoomNumber:  roomName,
			Capacity:    int(student.RoomCapacity),
			IsActive:    true,
		}
		s.roomRepo.Create(&room)
	} else if errDB != nil {
		log.Printf("[WARN] createResidentFromEIS: DB error looking up room %s/%s: %v", dormID, roomName, errDB)
		return
	}
	resident := &domain.Resident{
		UserID:      userID,
		DormitoryID: dormID,
		RoomID:      room.ID,
		IsActive:    true,
	}
	s.residentRepo.Create(resident)

	if contract != nil {
		resident.ContractNumber = &contract.Number
		if contract.DateIn != "" {
			v := truncateDate(contract.DateIn)
			resident.ContractStartDate = &v
		}
		if contract.DateEnd != "" {
			v := truncateDate(contract.DateEnd)
			resident.ContractEndDate = &v
		}
	}
	result.Resident = s.buildResidentInfo(resident)
}

// createEmployeeFromEIS создаёт Employee/EmployeeDormitoryRole из данных сотрудника.
func (s *DBService) createEmployeeFromEIS(employee *rabbitmq.EmployeeInfo, userID uuid.UUID, result *VerifyUserResult) {
	roleName := resolveEmployeeRole(employee.PositionName, employee.Department)
	emp := &domain.Employee{
		UserID:     userID,
		Position:   employee.PositionName,
		Department: employee.Department,
		Role:       roleName,
	}
	s.employeeRepo.Create(emp)

	result.Employee = s.buildEmployeeInfo(emp)

	targetDorm := s.resolveDormitoryForEmployeeEIS(employee)
	if targetDorm != nil {
		result.DormitoryID = targetDorm.ID.String()
		s.employeeDormitoryRoleRepo.CreateOrIgnore(&domain.EmployeeDormitoryRole{
			EmployeeID:  emp.ID,
			DormitoryID: targetDorm.ID,
			Role:        roleName,
		})
	}
}




func truncateDate(s string) string {
	if len(s) > 10 {
		return s[:10]
	}
	return s
}




func extractDormNumber(text string) string {
	lower := strings.ToLower(text)

	if i := strings.Index(lower, "общежитие №"); i != -1 {
		rest := lower[i+len("общежитие №"):]
		return extractNumber(rest)
	}
	if i := strings.Index(lower, "общежитие n"); i != -1 {
		rest := lower[i+len("общежитие n"):]
		return extractNumber(rest)
	}
	if i := strings.Index(lower, "общежитие "); i != -1 {
		rest := lower[i+len("общежитие "):]
		return extractNumber(rest)
	}
	return ""
}

func extractNumber(s string) string {
	s = strings.TrimSpace(s)
	var num string
	for i := 0; i < len(s) && s[i] >= '0' && s[i] <= '9'; i++ {
		num += string(s[i])
	}
	return num
}



func extractFloorFromRoomName(roomName string) int {
	s := strings.TrimSpace(roomName)
	if idx := strings.IndexAny(s, "/-"); idx != -1 {
		s = strings.TrimSpace(s[:idx])
	}
	var digitChars string
	for _, c := range s {
		if c >= '0' && c <= '9' {
			digitChars += string(c)
		}
	}
	if len(digitChars) == 0 {
		return 1
	}
	floorNum := 1
	if len(digitChars) <= 2 {
		fmt.Sscanf(digitChars, "%d", &floorNum)
	} else {
		floorStr := digitChars[:len(digitChars)-2]
		fmt.Sscanf(floorStr, "%d", &floorNum)
	}
	if floorNum < 1 || floorNum > 16 {
		return 1
	}
	return floorNum
}

func (s *DBService) ensureDormitoryFromName(ctx context.Context, name string, buildingID uint) uuid.UUID {
	normalized := NormalizeDormitoryName(name)
	dorm, err := s.dormitoryRepo.FindByField("name", normalized)
	if err == nil && dorm != nil {
		if dorm.EISDormitoryCode == "" && buildingID != 0 {
			dorm.EISDormitoryCode = fmt.Sprintf("EIS-BLD-%d", buildingID)
			s.dormitoryRepo.Update(dorm)
		}
		return dorm.ID
	}
	newDorm := &domain.Dormitory{
		Name:     normalized,
		IsActive: true,
	}
	if buildingID != 0 {
		newDorm.EISDormitoryCode = fmt.Sprintf("EIS-BLD-%d", buildingID)
	}
	s.dormitoryRepo.Create(newDorm)
	return newDorm.ID
}

func (s *DBService) resolveDormitoryForEmployeeEIS(employee *rabbitmq.EmployeeInfo) *domain.Dormitory {
	if employee.BuildingID != nil && *employee.BuildingID != 0 {
		dormCode := fmt.Sprintf("EIS-BLD-%d", *employee.BuildingID)
		dorm, _ := s.dormitoryRepo.FindByField("eis_dormitory_code", dormCode)
		if dorm != nil {
			return dorm
		}
	}
	if employee.Department != "" {
		dormNum := extractDormNumber(employee.Department)
		if dormNum != "" {
			normalized := NormalizeDormitoryName(employee.Department)
			dorm, _ := s.dormitoryRepo.FindByField("name", normalized)
			if dorm == nil {
				dorm, _ = s.dormitoryRepo.FindByField("name", "Общежитие №"+dormNum)
			}
			if dorm != nil {
				return dorm
			}
		}
	}
	return nil
}



func (s *DBService) buildResidentInfo(resident *domain.Resident) *ResidentInfo {
	info := &ResidentInfo{
		ResidentID:  resident.ID.String(),
		DormitoryID: resident.DormitoryID.String(),
		RoomID:      resident.RoomID.String(),
	}

	dorm, err := s.dormitoryRepo.GetByID(resident.DormitoryID)
	if err == nil {
		info.DormitoryName = dorm.Name
	}

	room, err := s.roomRepo.GetByID(resident.RoomID)
	if err == nil {
		info.RoomNumber = room.RoomNumber
		floor, err := s.floorRepo.GetByID(room.FloorID)
		if err == nil {
			info.FloorNumber = floor.FloorNumber
		}
	}

	if resident.ContractNumber != nil {
		info.ContractNumber = *resident.ContractNumber
	}
	if resident.ContractStartDate != nil {
		info.ContractStart = *resident.ContractStartDate
	}
	if resident.ContractEndDate != nil {
		info.ContractEnd = *resident.ContractEndDate
	}

	return info
}

func (s *DBService) buildEmployeeInfo(employee *domain.Employee) *EmployeeFullInfo {
	roleDisplay := MapRoleDisplay(string(employee.Role))
	return &EmployeeFullInfo{
		EmployeeID:  employee.ID.String(),
		Role:        string(employee.Role),
		RoleDisplay: roleDisplay,
		Position:    employee.Position,
	}
}

func MapRoleDisplay(role string) string {
	switch role {
	case "director":
		return "Директор"
	case "commandant":
		return "Заведующий общежитием"
	case "duty_officer":
		return "Дежурный"
	case "chairman":
		return "Председатель"
	case "starosta":
		return "Староста"
	default:
		return role
	}
}

// MapRoleDisplayWithLookup tries MapRoleDisplay first, then falls back to a lookup map.
// The lookup map should be populated from ListRoles() and maps role.Name → role.DisplayName.
func MapRoleDisplayWithLookup(role string, lookup map[string]string) string {
	if display := MapRoleDisplay(role); display != role {
		return display
	}
	if d, ok := lookup[role]; ok && d != "" {
		return d
	}
	return role
}

func (s *DBService) ConfirmResident(residentUUID string) error {
	id, _ := uuid.Parse(residentUUID)
	resident, err := s.residentRepo.GetByID(id)
	if err != nil {
		return err
	}
	log.Printf("Resident %s confirmed", resident.ID)
	return nil
}

type InitResidentResult struct {
	ResidentID    string
	DormitoryID   string
	DormitoryName string
	RoomID        string
	RoomNumber    string
	FloorNumber   int
}

func (s *DBService) InitResident(userUUID string) (*InitResidentResult, error) {
	uid, _ := uuid.Parse(userUUID)
	resident, err := s.residentRepo.FindByField("user_id", uid)
	if err != nil {
		return nil, fmt.Errorf("no resident data for user %s", userUUID)
	}
	if !resident.IsActive {
		return nil, fmt.Errorf("resident contract is no longer active")
	}

	info := s.buildResidentInfo(resident)
	return &InitResidentResult{
		ResidentID:    info.ResidentID,
		DormitoryID:   info.DormitoryID,
		DormitoryName: info.DormitoryName,
		RoomID:        info.RoomID,
		RoomNumber:    info.RoomNumber,
		FloorNumber:   info.FloorNumber,
	}, nil
}

type InitEmployeeResult struct {
	EmployeeID  string
	Role        string
	RoleDisplay string
	Position    string
}

func (s *DBService) InitEmployee(userUUID string) (*InitEmployeeResult, error) {
	uid, _ := uuid.Parse(userUUID)
	employee, err := s.employeeRepo.FindByField("user_id", uid)
	if err != nil {
		return nil, fmt.Errorf("no employee data for user %s", userUUID)
	}

	info := s.buildEmployeeInfo(employee)
	return &InitEmployeeResult{
		EmployeeID: info.EmployeeID,
		Role:       info.Role,
		Position:   info.Position,
	}, nil
}

type DormitoryInfo struct {
	DormitoryID   string
	DormitoryName string
	Role          string
}

func (s *DBService) GetEmployeeDormitories(employeeUUID string) ([]DormitoryInfo, error) {
	eid, _ := uuid.Parse(employeeUUID)
	roles, err := s.employeeDormitoryRoleRepo.GetByEmployee(eid)
	if err != nil {
		return nil, err
	}

	var result []DormitoryInfo
	for _, r := range roles {
		dorm, err := s.dormitoryRepo.GetByID(r.DormitoryID)
		if err != nil {
			continue
		}
		result = append(result, DormitoryInfo{
			DormitoryID:   dorm.ID.String(),
			DormitoryName: dorm.Name,
			Role:          string(r.Role),
		})
	}

	if len(result) == 0 {
		employee, err := s.employeeRepo.GetByID(eid)
		if err != nil {
			return nil, err
		}
		if employee.PrimaryDormitoryID != nil {
			dorm, err := s.dormitoryRepo.GetByID(*employee.PrimaryDormitoryID)
			if err == nil {
				result = append(result, DormitoryInfo{
					DormitoryID:   dorm.ID.String(),
					DormitoryName: dorm.Name,
					Role:          string(employee.Role),
				})
			}
		}
	}

	return result, nil
}

func (s *DBService) GetDormitoryName(dormitoryID string) (string, error) {
	id, _ := uuid.Parse(dormitoryID)
	dorm, err := s.dormitoryRepo.GetByID(id)
	if err != nil {
		return "", err
	}
	return dorm.Name, nil
}

func (s *DBService) GetEmployeeRoleInDormitory(employeeUUID string, dormitoryID string) (string, error) {
	eid, _ := uuid.Parse(employeeUUID)
	did, _ := uuid.Parse(dormitoryID)

	roles, err := s.employeeDormitoryRoleRepo.GetByEmployeeAndDormitory(eid, did)
	if err == nil && len(roles) > 0 {
		return string(roles[0].Role), nil
	}

	employee, err := s.employeeRepo.GetByID(eid)
	if err != nil {
		return "", err
	}
	return string(employee.Role), nil
}






func (s *DBService) CheckPermission(dormitoryID uuid.UUID, employeeID uuid.UUID, resource, action string) bool {
	return s.rbac.CheckPermission(dormitoryID, employeeID, resource, action)
}

func (s *DBService) GetEffectivePermissions(dormitoryID uuid.UUID, employeeID uuid.UUID) (domain.JSONAccessMap, error) {
	return s.rbac.GetEffectivePermissions(dormitoryID, employeeID)
}

func (s *DBService) GetUserByPhone(phone string) (*domain.User, error) {
	return s.userRepo.FindByField("phone", phone)
}

func (s *DBService) GetResidentByUserID(userID uuid.UUID) (*domain.Resident, error) {
	var resident domain.Resident
	err := s.residentRepo.DB().Preload("User").Preload("Dormitory").Preload("Room").Where("user_id = ?", userID).First(&resident).Error
	if err != nil {
		return nil, err
	}
	return &resident, nil
}


func (s *DBService) GetStaffByDormitory(dormitoryID uuid.UUID) ([]domain.UserDormitoryRoleWithUser, error) {
	udrResults, _ := s.userDormitoryRoleRepo.GetByDormitoryWithUser(dormitoryID)
	for i := range udrResults {
		udrResults[i].Source = "udr"
	}

	edrResults, err := s.employeeDormitoryRoleRepo.GetByDormitory(dormitoryID)
	if err != nil {
		return udrResults, nil
	}

	seenUserIDs := make(map[uuid.UUID]bool)
	for _, r := range udrResults {
		seenUserIDs[r.UserID] = true
	}

	for _, edr := range edrResults {
		employee, empErr := s.employeeRepo.GetByID(edr.EmployeeID)
		if empErr != nil {
			continue
		}
		if seenUserIDs[employee.UserID] {
			continue
		}

		user, userErr := s.userRepo.GetByID(employee.UserID)
		if userErr != nil {
			continue
		}

		udrResults = append(udrResults, domain.UserDormitoryRoleWithUser{
			UserDormitoryRole: domain.UserDormitoryRole{
				ID:          edr.ID,
				UserID:      employee.UserID,
				DormitoryID: edr.DormitoryID,
				Role:        string(edr.Role),
				CreatedAt:   edr.CreatedAt,
			},
			UserFirstName:  user.FirstName,
			UserLastName:   user.LastName,
			UserMiddleName: user.MiddleName,
			UserPhone:      user.Phone,
			PersonType:     user.PersonType,
			Source:         "edr",
			PlatformUserID: user.PlatformUserID,
		})
		seenUserIDs[employee.UserID] = true
	}

	return udrResults, nil
}

func (s *DBService) GetUserDormitoryRoles(userID, dormitoryID uuid.UUID) ([]domain.UserDormitoryRole, error) {
	return s.userDormitoryRoleRepo.GetByUserAndDormitory(userID, dormitoryID)
}

func (s *DBService) AssignRoleToUser(userID, dormitoryID, assignedBy uuid.UUID, roleName string) error {
	if err := s.userDormitoryRoleRepo.DeleteByUserAndDormitory(userID, dormitoryID); err != nil {
		return fmt.Errorf("ошибка удаления старых UDR: %w", err)
	}

	employee, err := s.employeeRepo.FindByField("user_id", userID)
	if err == nil && employee != nil {
		result := s.employeeDormitoryRoleRepo.DB().
			Where("employee_id = ? AND dormitory_id = ?", employee.ID, dormitoryID).
			Delete(&domain.EmployeeDormitoryRole{})
		log.Printf("AssignRoleToUser: deleted %d EDR rows for employee %s dormitory %s",
			result.RowsAffected, employee.ID, dormitoryID)

		result = s.employeeRepo.DB().
			Model(&domain.Employee{}).
			Where("id = ?", employee.ID).
			Update("role", roleName)
		if result.Error != nil {
			log.Printf("AssignRoleToUser: failed to update employee.Role: %v", result.Error)
		} else {
			log.Printf("AssignRoleToUser: updated employee %s role %s -> %s (rows=%d)",
				employee.ID, employee.Role, roleName, result.RowsAffected)
		}
	}

	udr := &domain.UserDormitoryRole{
		UserID:      userID,
		DormitoryID: dormitoryID,
		Role:        roleName,
		AssignedBy:  assignedBy,
	}
	return s.userDormitoryRoleRepo.Create(udr)
}

func (s *DBService) RemoveUserDormitoryRole(roleID, dormitoryID uuid.UUID) error {
	result := s.userDormitoryRoleRepo.DB().Where("id = ? AND dormitory_id = ?", roleID, dormitoryID).Delete(&domain.UserDormitoryRole{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("роль не найдена")
	}
	return nil
}

func (s *DBService) RemoveAllUserRoles(userID, dormitoryID uuid.UUID) error {
	return s.userDormitoryRoleRepo.DeleteByUserAndDormitory(userID, dormitoryID)
}

func (s *DBService) CreateRole(name, displayName, targetType string, priority int, jsonAccess domain.JSONAccessMap, isSystem bool) (*domain.Role, error) {
	role := &domain.Role{
		Name:        name,
		DisplayName: displayName,
		Priority:    priority,
		TargetType:  targetType,
		JSONAccess:  jsonAccess,
		IsSystem:    isSystem,
	}
	if err := s.roleRepo.Create(role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *DBService) UpdateRole(roleID uuid.UUID, displayName, targetType string, priority int, jsonAccess domain.JSONAccessMap) (*domain.Role, error) {
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		return nil, err
	}
	if role.IsSystem {
		return nil, fmt.Errorf("нельзя редактировать системную роль")
	}
	role.DisplayName = displayName
	role.TargetType = targetType
	role.Priority = priority
	role.JSONAccess = jsonAccess
	if err := s.roleRepo.Update(role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *DBService) DeleteRole(roleID uuid.UUID) error {
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		return err
	}
	if role.IsSystem {
		return fmt.Errorf("нельзя удалить системную роль")
	}
	return s.roleRepo.Delete(roleID)
}

func (s *DBService) GetRoleByID(roleID uuid.UUID) (*domain.Role, error) {
	return s.roleRepo.GetByID(roleID)
}

func (s *DBService) ListRoles() ([]domain.Role, error) {
	var roles []domain.Role
	err := s.roleRepo.DB().Order("priority DESC, name ASC").Find(&roles).Error
	return roles, err
}

func (s *DBService) GetUserByUserID(userID uuid.UUID) (*domain.User, error) {
	return s.userRepo.GetByID(userID)
}

func (s *DBService) GetUsersByIDs(ids []uuid.UUID) (map[uuid.UUID]*domain.User, error) {
	if len(ids) == 0 {
		return map[uuid.UUID]*domain.User{}, nil
	}
	var users []domain.User
	err := s.userRepo.DB().Where("id IN ?", ids).Find(&users).Error
	if err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]*domain.User)
	for i := range users {
		result[users[i].ID] = &users[i]
	}
	return result, nil
}

func (s *DBService) GetUserResidentRoleCount(userID, dormitoryID uuid.UUID) (int64, error) {
	return s.userDormitoryRoleRepo.CountByUserAndDormitory(userID, dormitoryID)
}

func (s *DBService) GetUserRolesForDormitory(userID, dormitoryID uuid.UUID) ([]domain.UserDormitoryRole, error) {
	return s.userDormitoryRoleRepo.GetByUserAndDormitory(userID, dormitoryID)
}

// ── Mini App Auth ──

// GenerateLaunchToken создаёт одноразовый токен для Mini App и сохраняет в Redis.
// TTL = 5 минут. Возвращает сам токен (hex-строка).
func (s *DBService) GenerateLaunchToken(userID, personType, dormitoryID, employeeID string) (string, error) {
	if s.cacheService == nil {
		return "", fmt.Errorf("cache service not available")
	}
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(tokenBytes)
	payload := map[string]string{
		"user_id":      userID,
		"person_type":  personType,
		"dormitory_id": dormitoryID,
		"employee_id":  employeeID,
	}
	key := "launch_token:" + token
	if err := s.cacheService.Set(context.Background(), key, payload, 5*time.Minute); err != nil {
		return "", err
	}
	return token, nil
}

// AuthorizeMiniApp обменивает launch token на JWT. Вызывается из POST /api/v1/auth/miniapp.
func (s *DBService) AuthorizeMiniApp(launchToken string, jwtSvc *auth.JWTService) (string, fiber.Map, error) {
	if s.cacheService == nil {
		return "", nil, fmt.Errorf("cache service not available")
	}
	key := "launch_token:" + launchToken
	var payload map[string]string
	found, err := s.cacheService.Get(context.Background(), key, &payload)
	if err != nil || !found {
		return "", nil, fmt.Errorf("invalid or expired launch token")
	}
	// Одноразовый — удаляем
	s.cacheService.Delete(context.Background(), key)

	jwtToken, err := jwtSvc.IssueToken(
		payload["user_id"],
		payload["person_type"],
		payload["dormitory_id"],
		payload["employee_id"],
	)
	if err != nil {
		return "", nil, fmt.Errorf("failed to issue JWT: %w", err)
	}

	profile := fiber.Map{
		"user_id":      payload["user_id"],
		"person_type":  payload["person_type"],
		"dormitory_id": payload["dormitory_id"],
		"employee_id":  payload["employee_id"],
	}
	return jwtToken, profile, nil
}
