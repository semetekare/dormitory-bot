package mock

import (
	"log"

	"github.com/dormitory-bot/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ── Mock UUID namespace ──
// Используем префикс b (mock), чтобы избежать конфликтов с seed-данными (a).
// ВАЖНО: префикс должен быть hex-цифрой [0-9a-f], иначе uuid.MustParse паникует.
var (
	mockDormID8   = uuid.MustParse("b0000000-0000-0000-0000-000000000001")
	mockDormID5   = uuid.MustParse("b0000000-0000-0000-0000-000000000002")

	// Dorm 8 floors
	mockFloor8_1 = uuid.MustParse("b0000000-0000-0000-0000-000000010001")
	mockFloor8_2 = uuid.MustParse("b0000000-0000-0000-0000-000000010002")
	mockFloor8_3 = uuid.MustParse("b0000000-0000-0000-0000-000000010003")

	// Dorm 5 floors
	mockFloor5_1 = uuid.MustParse("b0000000-0000-0000-0000-000000020001")
	mockFloor5_2 = uuid.MustParse("b0000000-0000-0000-0000-000000020002")

	// Wings (Dorm 8, floor 1)
	mockWingLeft  = uuid.MustParse("b0000000-0000-0000-0000-000000100001")
	mockWingRight = uuid.MustParse("b0000000-0000-0000-0000-000000100002")

	// Rooms Dorm 8
	mockRoom8_101 = uuid.MustParse("b0000000-0000-0000-0000-000000110101")
	mockRoom8_102 = uuid.MustParse("b0000000-0000-0000-0000-000000110102")
	mockRoom8_103 = uuid.MustParse("b0000000-0000-0000-0000-000000110103")
	mockRoom8_201 = uuid.MustParse("b0000000-0000-0000-0000-000000110201")
	mockRoom8_202 = uuid.MustParse("b0000000-0000-0000-0000-000000110202")
	mockRoom8_301 = uuid.MustParse("b0000000-0000-0000-0000-000000110301")

	// Rooms Dorm 5
	mockRoom5_101 = uuid.MustParse("b0000000-0000-0000-0000-000000120101")
	mockRoom5_102 = uuid.MustParse("b0000000-0000-0000-0000-000000120102")
	mockRoom5_201 = uuid.MustParse("b0000000-0000-0000-0000-000000120201")
	mockRoom5_202 = uuid.MustParse("b0000000-0000-0000-0000-000000120202")

	// Washing machines
	mockWM8_1 = uuid.MustParse("b0000000-0000-0000-0000-000000200001")
	mockWM8_2 = uuid.MustParse("b0000000-0000-0000-0000-000000200002")

	// Laundry settings
	mockLS8 = uuid.MustParse("b0000000-0000-0000-0000-000000300001")
	mockLS5 = uuid.MustParse("b0000000-0000-0000-0000-000000300002")

	// Material items
	mockMat1 = uuid.MustParse("b0000000-0000-0000-0000-000000400001")
	mockMat2 = uuid.MustParse("b0000000-0000-0000-0000-000000400002")
	mockMat3 = uuid.MustParse("b0000000-0000-0000-0000-000000400003")

	// Chat links
	mockCL1 = uuid.MustParse("b0000000-0000-0000-0000-000000500001")
	mockCL2 = uuid.MustParse("b0000000-0000-0000-0000-000000500002")
	mockCL3 = uuid.MustParse("b0000000-0000-0000-0000-000000500003")

	// Reference materials
	mockRef1 = uuid.MustParse("b0000000-0000-0000-0000-000000600001")
	mockRef2 = uuid.MustParse("b0000000-0000-0000-0000-000000600002")
	mockRef3 = uuid.MustParse("b0000000-0000-0000-0000-000000600003")
)

// LoadMockSyncData загружает все мок-данные в БД бота.
// Вызывается один раз при старте в mock-режиме.
// Идемпотентна — проверяет существование записей перед созданием.
func LoadMockSyncData(db *gorm.DB) error {
	var count int64
	db.Model(&domain.Dormitory{}).Where("id IN ?", []uuid.UUID{mockDormID8, mockDormID5}).Count(&count)
	if count > 0 {
		log.Println("[mock] LoadMockSyncData: data already exists, skipping")
		return nil
	}

	log.Println("[mock] Loading mock data into database...")

	// 1. Dormitories
	dorms := []domain.Dormitory{
		{ID: mockDormID8, Name: "Общежитие №8", Address: "ул. Лермонтова, д. 80", EISDormitoryCode: "EIS-BLD-8", IsActive: true},
		{ID: mockDormID5, Name: "Общежитие №5", Address: "ул. Лермонтова, д. 75", EISDormitoryCode: "EIS-BLD-5", IsActive: true},
	}
	for _, d := range dorms {
		db.Create(&d)
	}
	log.Printf("[mock] Created %d dormitories", len(dorms))

	// 2. Floors
	floors := []domain.Floor{
		{ID: mockFloor8_1, DormitoryID: mockDormID8, FloorNumber: 1},
		{ID: mockFloor8_2, DormitoryID: mockDormID8, FloorNumber: 2},
		{ID: mockFloor8_3, DormitoryID: mockDormID8, FloorNumber: 3},
		{ID: mockFloor5_1, DormitoryID: mockDormID5, FloorNumber: 1},
		{ID: mockFloor5_2, DormitoryID: mockDormID5, FloorNumber: 2},
	}
	for _, f := range floors {
		db.Create(&f)
	}
	log.Printf("[mock] Created %d floors", len(floors))

	// 3. Wings (only Dorm 8 floor 1 has wings)
	wings := []domain.Wing{
		{ID: mockWingLeft, FloorID: mockFloor8_1, Name: "Левое"},
		{ID: mockWingRight, FloorID: mockFloor8_1, Name: "Правое"},
	}
	for _, w := range wings {
		db.Create(&w)
	}
	log.Printf("[mock] Created %d wings", len(wings))

	// 4. Rooms
	rooms := []domain.Room{
		// Dorm 8
		{ID: mockRoom8_101, DormitoryID: mockDormID8, FloorID: mockFloor8_1, WingID: &mockWingLeft, RoomNumber: "101", Capacity: 3, IsActive: true},
		{ID: mockRoom8_102, DormitoryID: mockDormID8, FloorID: mockFloor8_1, WingID: &mockWingLeft, RoomNumber: "102", Capacity: 2, IsActive: true},
		{ID: mockRoom8_103, DormitoryID: mockDormID8, FloorID: mockFloor8_1, WingID: &mockWingLeft, RoomNumber: "103", Capacity: 4, IsActive: true},
		{ID: mockRoom8_201, DormitoryID: mockDormID8, FloorID: mockFloor8_2, WingID: &mockWingRight, RoomNumber: "201", Capacity: 2, IsActive: true},
		{ID: mockRoom8_202, DormitoryID: mockDormID8, FloorID: mockFloor8_2, WingID: &mockWingRight, RoomNumber: "202", Capacity: 3, IsActive: true},
		{ID: mockRoom8_301, DormitoryID: mockDormID8, FloorID: mockFloor8_3, WingID: nil, RoomNumber: "301", Capacity: 2, IsActive: true},
		// Dorm 5
		{ID: mockRoom5_101, DormitoryID: mockDormID5, FloorID: mockFloor5_1, WingID: nil, RoomNumber: "101", Capacity: 3, IsActive: true},
		{ID: mockRoom5_102, DormitoryID: mockDormID5, FloorID: mockFloor5_1, WingID: nil, RoomNumber: "102", Capacity: 2, IsActive: true},
		{ID: mockRoom5_201, DormitoryID: mockDormID5, FloorID: mockFloor5_2, WingID: nil, RoomNumber: "201", Capacity: 2, IsActive: true},
		{ID: mockRoom5_202, DormitoryID: mockDormID5, FloorID: mockFloor5_2, WingID: nil, RoomNumber: "202", Capacity: 3, IsActive: true},
	}
	for _, r := range rooms {
		db.Create(&r)
	}
	log.Printf("[mock] Created %d rooms", len(rooms))

	// 5. Washing machines
	washingMachines := []domain.WashingMachine{
		{ID: mockWM8_1, DormitoryID: mockDormID8, FloorID: mockFloor8_1, MachineNumber: "1", IsActive: true, DurationMinutes: 60},
		{ID: mockWM8_2, DormitoryID: mockDormID8, FloorID: mockFloor8_2, MachineNumber: "2", IsActive: true, DurationMinutes: 60},
	}
	for _, m := range washingMachines {
		db.Create(&m)
	}
	log.Printf("[mock] Created %d washing machines", len(washingMachines))

	// 6. Laundry settings
	laundrySettings := []domain.LaundrySettings{
		{ID: mockLS8, DormitoryID: mockDormID8, BookingStartTime: "07:00:00", BookingEndTime: "23:00:00", DefaultDurationMinutes: 60},
		{ID: mockLS5, DormitoryID: mockDormID5, BookingStartTime: "07:00:00", BookingEndTime: "23:00:00", DefaultDurationMinutes: 60},
	}
	for _, ls := range laundrySettings {
		db.Create(&ls)
	}
	log.Printf("[mock] Created %d laundry settings", len(laundrySettings))

	// 7. Material responsibility items
	materialItems := []domain.MaterialResponsibility{
		{ID: mockMat1, RoomID: mockRoom8_101, ItemName: "Кровать", InventoryNumber: "MOCK-INV-101-001", Quantity: 3, Condition: "хорошее"},
		{ID: mockMat2, RoomID: mockRoom8_101, ItemName: "Стол письменный", InventoryNumber: "MOCK-INV-101-002", Quantity: 3, Condition: "хорошее"},
		{ID: mockMat3, RoomID: mockRoom8_102, ItemName: "Кровать", InventoryNumber: "MOCK-INV-102-001", Quantity: 2, Condition: "удовл."},
	}
	for _, item := range materialItems {
		db.Create(&item)
	}
	log.Printf("[mock] Created %d material items", len(materialItems))

	// 8. Create mock users with employees and residents FIRST
	// (chat_links and reference_materials need a valid employee FK)
	loadMockUsersAndRoles(db)

	// 9. Chat links (requires existing employee for CreatedBy FK)
	var createdByEmp domain.Employee
	if err := db.Where("role = ?", "commandant").First(&createdByEmp).Error; err != nil {
		log.Printf("[mock] WARNING: no commandant employee found for chat_links FK, using uuid.Nil")
		createdByEmp.ID = uuid.Nil
	}
	chatLinks := []domain.ChatLink{
		{ID: mockCL1, DormitoryID: mockDormID8, LinkType: "dormitory", Platform: "telegram", URL: "https://t.me/dorm8_mock", Title: "Чат общежития №8", IsActive: true, CreatedBy: createdByEmp.ID},
		{ID: mockCL2, DormitoryID: mockDormID8, FloorID: &mockFloor8_1, LinkType: "floor", Platform: "telegram", URL: "https://t.me/dorm8_floor1_mock", Title: "Чат 1 этажа", IsActive: true, CreatedBy: createdByEmp.ID},
		{ID: mockCL3, DormitoryID: mockDormID5, LinkType: "dormitory", Platform: "telegram", URL: "https://t.me/dorm5_mock", Title: "Чат общежития №5", IsActive: true, CreatedBy: createdByEmp.ID},
	}
	for _, cl := range chatLinks {
		db.Create(&cl)
	}
	log.Printf("[mock] Created %d chat links", len(chatLinks))

	// 10. Reference materials
	refMaterials := []domain.ReferenceMaterial{
		{ID: mockRef1, DormitoryID: mockDormID8, Name: "Правила проживания", Description: "Правила внутреннего распорядка общежития. Запрещено: курение, распитие алкоголя, шум после 23:00.", Category: "Правила", Ordinal: 1, CreatedBy: createdByEmp.ID},
		{ID: mockRef2, DormitoryID: mockDormID8, Name: "Как пользоваться прачкой", Description: "Запись на стирку производится через бота в разделе «Стирка». Одна стирка — 60 минут. Отмена — не позднее чем за 15 минут до начала.", Category: "Прачка", Ordinal: 2, CreatedBy: createdByEmp.ID},
		{ID: mockRef3, DormitoryID: mockDormID8, Name: "График дежурств", Description: "Дежурства назначаются автоматически в начале каждого месяца. Комнаты дежурят по очереди. За невыполнение — штрафное дежурство.", Category: "Дежурства", Ordinal: 3, CreatedBy: createdByEmp.ID},
	}
	for _, m := range refMaterials {
		db.Create(&m)
	}
	log.Printf("[mock] Created %d reference materials", len(refMaterials))

	log.Println("[mock] Mock data loaded successfully")
	return nil
}

// mockUserEntry описывает одного пользователя для загрузки.
type mockUserEntry struct {
	Phone       string
	LastName    string
	FirstName   string
	MiddleName  string
	PersonType  string // "employee" or "student"
	EISPersonID string
	// Employee fields (if PersonType == "employee")
	Position   string
	Department string
	Role       string
	BuildingID int // EIS building ID for dormitory resolution
	// Student fields (if PersonType == "student")
	RoomID        uuid.UUID
	DormitoryID   uuid.UUID
	ContractNum   string
	ContractStart string
	ContractEnd   string
	DebtAmount    float64
	DebtPeriod    string
	DebtDesc      string
}

func loadMockUsersAndRoles(db *gorm.DB) {
	entries := []mockUserEntry{
		// ── Employees ──
		{
			Phone: "79025643215", LastName: "Высоких", FirstName: "Данил", MiddleName: "Александрович",
			PersonType: "employee", EISPersonID: "1001",
			Position: "Заведующий общежитием", Department: "Администрация студгородка", Role: "commandant", BuildingID: 8,
		},
		{
			Phone: "79140001122", LastName: "Кузнецова", FirstName: "Елена", MiddleName: "Петровна",
			PersonType: "employee", EISPersonID: "1002",
			Position: "Дежурный", Department: "Общежитие №8", Role: "duty_officer", BuildingID: 8,
		},
		{
			Phone: "79140001133", LastName: "Морозов", FirstName: "Игорь", MiddleName: "Викторович",
			PersonType: "employee", EISPersonID: "1003",
			Position: "Заведующий общежитием", Department: "Общежитие №5", Role: "commandant", BuildingID: 5,
		},
		{
			Phone: "79140001144", LastName: "Фёдорова", FirstName: "Анна", MiddleName: "Сергеевна",
			PersonType: "employee", EISPersonID: "1004",
			Position: "Председатель студсовета", Department: "Общежитие №8", Role: "chairman", BuildingID: 8,
		},
		{
			Phone: "79140001155", LastName: "Григорьев", FirstName: "Павел", MiddleName: "Денисович",
			PersonType: "employee", EISPersonID: "1005",
			Position: "Староста", Department: "Общежитие №8", Role: "starosta", BuildingID: 8,
		},
		// ── Students (Dorm 8) ──
		{
			Phone: "79001234501", LastName: "Иванов", FirstName: "Артём", MiddleName: "Сергеевич",
			PersonType: "student", EISPersonID: "1010",
			RoomID: mockRoom8_101, DormitoryID: mockDormID8,
			ContractNum: "A-2026-001", ContractStart: "2025-09-01", ContractEnd: "2026-08-31",
			DebtAmount: 1500.00, DebtPeriod: "2026-06", DebtDesc: "Задолженность за проживание (июнь)",
		},
		{
			Phone: "79001234502", LastName: "Петров", FirstName: "Максим", MiddleName: "Алексеевич",
			PersonType: "student", EISPersonID: "1011",
			RoomID: mockRoom8_101, DormitoryID: mockDormID8,
			ContractNum: "A-2026-002", ContractStart: "2025-09-01", ContractEnd: "2026-08-31",
		},
		{
			Phone: "79001234503", LastName: "Смирнова", FirstName: "Анна", MiddleName: "Дмитриевна",
			PersonType: "student", EISPersonID: "1012",
			RoomID: mockRoom8_102, DormitoryID: mockDormID8,
			ContractNum: "A-2026-003", ContractStart: "2025-09-01", ContractEnd: "2026-08-31",
			DebtAmount: 3200.00, DebtPeriod: "2026-05/2026-06", DebtDesc: "Задолженность за проживание (май-июнь)",
		},
		{
			Phone: "79001234504", LastName: "Кузнецов", FirstName: "Дмитрий", MiddleName: "Игоревич",
			PersonType: "student", EISPersonID: "1013",
			RoomID: mockRoom8_103, DormitoryID: mockDormID8,
			ContractNum: "A-2026-004", ContractStart: "2025-09-01", ContractEnd: "2026-08-31",
		},
		{
			Phone: "79001234505", LastName: "Попов", FirstName: "Алексей", MiddleName: "Владимирович",
			PersonType: "student", EISPersonID: "1014",
			RoomID: mockRoom8_103, DormitoryID: mockDormID8,
			ContractNum: "A-2026-005", ContractStart: "2025-09-01", ContractEnd: "2026-08-31",
		},
		{
			Phone: "79001234506", LastName: "Соколов", FirstName: "Никита", MiddleName: "Павлович",
			PersonType: "student", EISPersonID: "1015",
			RoomID: mockRoom8_201, DormitoryID: mockDormID8,
			ContractNum: "A-2026-006", ContractStart: "2025-09-01", ContractEnd: "2026-08-31",
			DebtAmount: 5000.00, DebtPeriod: "2026-04/2026-06", DebtDesc: "Просроченная задолженность (апрель-июнь)",
		},
		{
			Phone: "79001234507", LastName: "Васильева", FirstName: "Ольга", MiddleName: "Игоревна",
			PersonType: "student", EISPersonID: "1016",
			RoomID: mockRoom8_202, DormitoryID: mockDormID8,
			ContractNum: "A-2026-007", ContractStart: "2025-09-01", ContractEnd: "2026-08-31",
		},
		// ── Students (Dorm 5) ──
		{
			Phone: "79001234508", LastName: "Белов", FirstName: "Сергей", MiddleName: "Николаевич",
			PersonType: "student", EISPersonID: "1017",
			RoomID: mockRoom5_101, DormitoryID: mockDormID5,
			ContractNum: "A-2025-008", ContractStart: "2024-09-01", ContractEnd: "2025-08-31",
		},
		{
			Phone: "79001234509", LastName: "Морозова", FirstName: "Ирина", MiddleName: "Александровна",
			PersonType: "student", EISPersonID: "1018",
			RoomID: mockRoom5_102, DormitoryID: mockDormID5,
			ContractNum: "A-2026-009", ContractStart: "2025-09-01", ContractEnd: "2026-08-31",
		},
		// Student WITHOUT contract — will be verified but won't have a resident record
		// until the contract is added. Person exists in the dataset but no contract.
	}

	createdUsers := 0
	createdEmployees := 0
	createdResidents := 0
	createdEDR := 0
	createdDebts := 0

	for _, e := range entries {
		// Check if user already exists
		var existing domain.User
		if err := db.Where("phone = ?", e.Phone).First(&existing).Error; err == nil {
			// User exists, skip
			continue
		}

		user := domain.User{
			FirstName:   e.FirstName,
			LastName:    e.LastName,
			MiddleName:  e.MiddleName,
			Phone:       e.Phone,
			Platform:    "max",
			PlatformUserID: "mock_" + e.Phone,
			EISVerified: true,
			EISPersonID: e.EISPersonID,
			PersonType:  e.PersonType,
		}
		if err := db.Create(&user).Error; err != nil {
			log.Printf("[mock] Failed to create user %s: %v", e.Phone, err)
			continue
		}
		createdUsers++

		if e.PersonType == "employee" {
			var dormID uuid.UUID
			switch e.BuildingID {
			case 5:
				dormID = mockDormID5
			default:
				dormID = mockDormID8
			}

			emp := domain.Employee{
				UserID:             user.ID,
				Position:           e.Position,
				Department:         e.Department,
				PrimaryDormitoryID: &dormID,
				Role:               e.Role,
			}
			if err := db.Create(&emp).Error; err != nil {
				log.Printf("[mock] Failed to create employee for user %s: %v", e.Phone, err)
				continue
			}
			createdEmployees++

			edr := domain.EmployeeDormitoryRole{
				EmployeeID:  emp.ID,
				DormitoryID: dormID,
				Role:        e.Role,
			}
			if err := db.Create(&edr).Error; err != nil {
				log.Printf("[mock] Failed to create EDR for employee %s: %v", emp.ID, err)
			} else {
				createdEDR++
			}
		} else if e.PersonType == "student" && e.RoomID != uuid.Nil {
			resident := domain.Resident{
				UserID:      user.ID,
				DormitoryID: e.DormitoryID,
				RoomID:      e.RoomID,
				IsActive:    true,
			}
			if e.ContractNum != "" {
				resident.ContractNumber = &e.ContractNum
			}
			if e.ContractStart != "" {
				resident.ContractStartDate = &e.ContractStart
			}
			if e.ContractEnd != "" {
				resident.ContractEndDate = &e.ContractEnd
			}
			if err := db.Create(&resident).Error; err != nil {
				log.Printf("[mock] Failed to create resident for user %s: %v", e.Phone, err)
				continue
			}
			createdResidents++

			if e.DebtAmount > 0 {
				debt := domain.ResidentFinancialDebt{
					ResidentID:  resident.ID,
					Amount:      e.DebtAmount,
					Period:      e.DebtPeriod,
					Description: e.DebtDesc,
					Source:      "EIS",
				}
				if err := db.Create(&debt).Error; err != nil {
					log.Printf("[mock] Failed to create debt: %v", err)
				} else {
					createdDebts++
				}
			}
		}
	}

	log.Printf("[mock] Users: created=%d employees=%d residents=%d edr=%d debts=%d",
		createdUsers, createdEmployees, createdResidents, createdEDR, createdDebts)
}
