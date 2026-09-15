package database

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/dormitory-bot/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func jsonAccessFromRegistry(modules ...string) string {
	access := make(domain.JSONAccessMap)
	for _, m := range modules {
		access["resource."+m] = domain.ResourcePermissions{"*": true}
	}
	data, _ := json.Marshal(access)
	return string(data)
}

func jsonAccessReadOnly(modules ...string) string {
	access := make(domain.JSONAccessMap)
	for _, m := range modules {
		access["resource."+m] = domain.ResourcePermissions{"read": true}
	}
	data, _ := json.Marshal(access)
	return string(data)
}

func jsonAccessCRUD(modules ...string) string {
	access := make(domain.JSONAccessMap)
	for _, m := range modules {
		access["resource."+m] = domain.ResourcePermissions{
			"read": true, "create": true, "update": true, "delete": true,
		}
	}
	data, _ := json.Marshal(access)
	return string(data)
}

func SeedRoles(db *gorm.DB) {
	roles := []struct {
		Name        string
		DisplayName string
		Priority    int
		IsSystem    bool
		TargetType  string
		JSONAccess  string
	}{
		{
			Name:        "director",
			DisplayName: "Директор",
			Priority:    100,
			IsSystem:    true,
			TargetType:  "employee",
			JSONAccess:  jsonAccessFromRegistry("laundry", "room", "cleaning", "chat_link", "reference", "module", "staff", "role"),
		},
		{
			Name:        "commandant",
			DisplayName: "Заведующий общежитием",
			Priority:    80,
			IsSystem:    true,
			TargetType:  "employee",
			JSONAccess:  jsonAccessFromRegistry("laundry", "room", "cleaning", "chat_link", "reference", "module", "staff", "role"),
		},
		{
			Name:        "duty_officer",
			DisplayName: "Дежурный",
			Priority:    60,
			IsSystem:    true,
			TargetType:  "employee",
			JSONAccess:  jsonAccessReadOnly("laundry", "room", "cleaning", "chat_link", "reference"),
		},
		{
			Name:        "chairman",
			DisplayName: "Председатель",
			Priority:    40,
			IsSystem:    true,
			TargetType:  "employee",
			JSONAccess:  jsonAccessCRUD("room", "cleaning"),
		},
		{
			Name:        "starosta",
			DisplayName: "Староста",
			Priority:    20,
			IsSystem:    true,
			TargetType:  "employee",
			JSONAccess:  jsonAccessCRUD("cleaning"),
		},
		{
			Name:        "assistant",
			DisplayName: "Помощник заведующего",
			Priority:    50,
			IsSystem:    false,
			TargetType:  "employee",
			JSONAccess:  jsonAccessReadOnly("laundry", "room", "cleaning", "chat_link", "reference"),
		},
	}

	for _, r := range roles {
		row := db.Raw(`SELECT count(*) as cnt FROM role WHERE name = ?`, r.Name)
		var cnt int64
		row.Scan(&cnt)

		if cnt == 0 {
			db.Exec(`INSERT INTO role (id, name, display_name, priority, is_system, target_type, json_access, created_at, updated_at)
				VALUES (gen_random_uuid(), ?, ?, ?, ?, ?, ?::jsonb, now(), now())`,
				r.Name, r.DisplayName, r.Priority, r.IsSystem, r.TargetType, r.JSONAccess)
		} else {
			db.Exec(`UPDATE role SET json_access = ?::jsonb, display_name = ?, priority = ?, is_system = ?, target_type = ?, updated_at = now() WHERE name = ?`,
				r.JSONAccess, r.DisplayName, r.Priority, r.IsSystem, r.TargetType, r.Name)
		}
	}
	log.Println("SeedRoles completed")
}

func SeedDormitories(db *gorm.DB) {
	var count int64
	db.Model(&domain.Dormitory{}).Count(&count)
	if count > 0 {
		log.Println("SeedDormitories: data already exists, skipping")
		return
	}

	dorm := domain.Dormitory{
		ID:               uuid.MustParse("a0000000-0000-0000-0000-000000000001"),
		Name:             "Общежитие №8",
		Address:          "ул. Лермонтова, д. 80",
		EISDormitoryCode: "EIS-DORM-8",
		IsActive:         true,
	}
	if err := db.Create(&dorm).Error; err != nil {
		log.Printf("SeedDormitories error: %v", err)
		return
	}
	log.Printf("Created dormitory: %s (%s)", dorm.Name, dorm.ID)

	floors := []domain.Floor{
		{ID: uuid.MustParse("b0000000-0000-0000-0000-000000000001"), DormitoryID: dorm.ID, FloorNumber: 1},
		{ID: uuid.MustParse("b0000000-0000-0000-0000-000000000002"), DormitoryID: dorm.ID, FloorNumber: 2},
		{ID: uuid.MustParse("b0000000-0000-0000-0000-000000000003"), DormitoryID: dorm.ID, FloorNumber: 3},
	}
	for _, f := range floors {
		db.Create(&f)
	}
	log.Printf("Created %d floors", len(floors))

	wings := []domain.Wing{
		{ID: uuid.MustParse("c0000000-0000-0000-0000-000000000001"), FloorID: floors[0].ID, Name: "Левое"},
		{ID: uuid.MustParse("c0000000-0000-0000-0000-000000000002"), FloorID: floors[0].ID, Name: "Правое"},
	}
	for _, w := range wings {
		db.Create(&w)
	}
	log.Printf("Created %d wings", len(wings))

	rooms := []domain.Room{
		{ID: uuid.MustParse("d0000000-0000-0000-0000-000000000101"), DormitoryID: dorm.ID, FloorID: floors[0].ID, WingID: &wings[0].ID, RoomNumber: "101", Capacity: 3, IsActive: true},
		{ID: uuid.MustParse("d0000000-0000-0000-0000-000000000102"), DormitoryID: dorm.ID, FloorID: floors[0].ID, WingID: &wings[0].ID, RoomNumber: "102", Capacity: 2, IsActive: true},
		{ID: uuid.MustParse("d0000000-0000-0000-0000-000000000103"), DormitoryID: dorm.ID, FloorID: floors[0].ID, WingID: &wings[0].ID, RoomNumber: "103", Capacity: 4, IsActive: true},
		{ID: uuid.MustParse("d0000000-0000-0000-0000-000000000201"), DormitoryID: dorm.ID, FloorID: floors[1].ID, WingID: &wings[1].ID, RoomNumber: "201", Capacity: 2, IsActive: true},
		{ID: uuid.MustParse("d0000000-0000-0000-0000-000000000202"), DormitoryID: dorm.ID, FloorID: floors[1].ID, WingID: &wings[1].ID, RoomNumber: "202", Capacity: 3, IsActive: true},
		{ID: uuid.MustParse("d0000000-0000-0000-0000-000000000301"), DormitoryID: dorm.ID, FloorID: floors[2].ID, RoomNumber: "301", Capacity: 2, IsActive: true},
	}
	for _, r := range rooms {
		db.Create(&r)
	}
	log.Printf("Created %d rooms", len(rooms))

	washingMachines := []domain.WashingMachine{
		{ID: uuid.MustParse("e0000000-0000-0000-0000-000000000001"), DormitoryID: dorm.ID, FloorID: floors[0].ID, MachineNumber: "1", IsActive: true, DurationMinutes: 60},
		{ID: uuid.MustParse("e0000000-0000-0000-0000-000000000002"), DormitoryID: dorm.ID, FloorID: floors[1].ID, MachineNumber: "2", IsActive: true, DurationMinutes: 60},
	}
	for _, m := range washingMachines {
		db.Create(&m)
	}
	log.Printf("Created %d washing machines", len(washingMachines))

	laundrySettings := domain.LaundrySettings{
		ID:                     uuid.MustParse("f0000000-0000-0000-0000-000000000001"),
		DormitoryID:            dorm.ID,
		BookingStartTime:       "07:00",
		DefaultDurationMinutes: 60,
	}
	db.Create(&laundrySettings)
	log.Println("Created laundry settings")

	materialItems := []domain.MaterialResponsibility{
		{ID: uuid.MustParse("fa000000-0000-0000-0000-000000000001"), RoomID: rooms[0].ID, ItemName: "Кровать", InventoryNumber: "INV-101-001", Quantity: 3, Condition: "хорошее"},
		{ID: uuid.MustParse("fa000000-0000-0000-0000-000000000002"), RoomID: rooms[0].ID, ItemName: "Стол письменный", InventoryNumber: "INV-101-002", Quantity: 3, Condition: "хорошее"},
		{ID: uuid.MustParse("fa000000-0000-0000-0000-000000000003"), RoomID: rooms[1].ID, ItemName: "Кровать", InventoryNumber: "INV-102-001", Quantity: 2, Condition: "удовл."},
	}
	for _, item := range materialItems {
		db.Create(&item)
	}
	log.Println("Created material responsibility items")

	log.Println("SeedDormitories completed")
}

func SeedTestUser(db *gorm.DB, phone string, maxUserID int64, platform string) {
	dormitoryID := uuid.MustParse("a0000000-0000-0000-0000-000000000001")

	var user domain.User
	err := db.Where("phone = ?", phone).First(&user).Error
	if err != nil {
		user = domain.User{
			FirstName:      "Данил",
			LastName:       "Высоких",
			MiddleName:     "Александрович",
			Phone:          phone,
			Platform:       platform,
			PlatformUserID: fmt.Sprintf("%d", maxUserID),
			EISVerified:    true,
			EISPersonID:    "1",
			PersonType:     "employee",
		}
		if err := db.Create(&user).Error; err != nil {
			log.Printf("SeedTestUser error: %v", err)
			return
		}
		log.Printf("Created test user: %s %s %s (phone: %s, id: %s, type: employee)",
			user.LastName, user.FirstName, user.MiddleName, user.Phone, user.ID)
	} else {
		user.FirstName = "Данил"
		user.LastName = "Высоких"
		user.MiddleName = "Александрович"
		user.PersonType = "employee"
		user.PlatformUserID = fmt.Sprintf("%d", maxUserID)
		user.EISVerified = true
		user.EISPersonID = "1"
		db.Save(&user)
		log.Printf("SeedTestUser: updated existing user %s to employee/commandant", user.ID)
	}

	var employee domain.Employee
	err = db.Where("user_id = ?", user.ID).First(&employee).Error
	if err != nil {
		employee = domain.Employee{
			UserID:             user.ID,
			Position:           "Заведующий общежитием",
			Department:         "Администрация студгородка",
			PrimaryDormitoryID: &dormitoryID,
			Role:               "commandant",
		}
		if err := db.Create(&employee).Error; err != nil {
			log.Printf("SeedTestUser employee error: %v", err)
			return
		}
		log.Printf("Created employee: commandant, dorm 1 (employee_id: %s)", employee.ID)
	} else {
		employee.Position = "Заведующий общежитием"
		employee.Department = "Администрация студгородка"
		employee.Role = "commandant"
		employee.PrimaryDormitoryID = &dormitoryID
		db.Save(&employee)
		log.Printf("SeedTestUser: updated employee %s to commandant", employee.ID)
	}

	var edr domain.EmployeeDormitoryRole
	err = db.Where("employee_id = ? AND dormitory_id = ?", employee.ID, dormitoryID).First(&edr).Error
	if err != nil {
		edr = domain.EmployeeDormitoryRole{
			EmployeeID:  employee.ID,
			DormitoryID: dormitoryID,
			Role:        "commandant",
		}
		if err := db.Create(&edr).Error; err != nil {
			log.Printf("SeedTestUser EDR error: %v", err)
			return
		}
		log.Printf("Created employee_dormitory_role: commandant in dorm 1")
	} else {
		edr.Role = "commandant"
		db.Save(&edr)
		log.Printf("SeedTestUser: updated EDR role to commandant")
	}

	var resident domain.Resident
	err = db.Where("user_id = ?", user.ID).First(&resident).Error
	if err != nil {
		room101 := uuid.MustParse("d0000000-0000-0000-0000-000000000101")
		resident = domain.Resident{
			UserID:      user.ID,
			DormitoryID: dormitoryID,
			RoomID:      room101,
		}
		if err := db.Create(&resident).Error; err != nil {
			log.Printf("SeedTestUser resident error: %v", err)
			return
		}
		log.Printf("Created resident for user %s: room 101 (resident_id: %s)", user.ID, resident.ID)
	} else {
		log.Printf("SeedTestUser: resident already exists (id: %s, room: %s)", resident.ID, resident.RoomID)
	}

	seedResidentsForRooms(db, dormitoryID, user.ID)
}

func DefaultSeedConfig(mode string) string {
	return mode
}

func SeedMode(mode string) string {
	return mode
}

func seedResidentsForRooms(db *gorm.DB, dormitoryID uuid.UUID, excludeUserID uuid.UUID) {
	rooms := []uuid.UUID{
		uuid.MustParse("d0000000-0000-0000-0000-000000000101"),
		uuid.MustParse("d0000000-0000-0000-0000-000000000102"),
		uuid.MustParse("d0000000-0000-0000-0000-000000000103"),
		uuid.MustParse("d0000000-0000-0000-0000-000000000201"),
		uuid.MustParse("d0000000-0000-0000-0000-000000000202"),
		uuid.MustParse("d0000000-0000-0000-0000-000000000301"),
	}

	type testResident struct {
		lastName   string
		firstName  string
		middleName string
		phone      string
		roomID     uuid.UUID
		debt       float64
		debtDesc   string
	}

	phoneCounter := 1000
	var testUsers []testResident

	for _, roomID := range rooms {
		switch roomID {
		case uuid.MustParse("d0000000-0000-0000-0000-000000000101"):
			testUsers = append(testUsers,
				testResident{lastName: "Иванов", firstName: "Артём", middleName: "Сергеевич", phone: fmt.Sprintf("7999%06d", phoneCounter), roomID: roomID, debt: 1500.00, debtDesc: "Задолженность за проживание (июнь)"},
				testResident{lastName: "Петров", firstName: "Максим", middleName: "Алексеевич", phone: fmt.Sprintf("7999%06d", phoneCounter+1), roomID: roomID},
			)
			phoneCounter += 2
		case uuid.MustParse("d0000000-0000-0000-0000-000000000102"):
			testUsers = append(testUsers,
				testResident{lastName: "Смирнова", firstName: "Анна", middleName: "Дмитриевна", phone: fmt.Sprintf("7999%06d", phoneCounter), roomID: roomID, debt: 3200.00, debtDesc: "Задолженность за проживание (май-июнь)"},
			)
			phoneCounter++
		case uuid.MustParse("d0000000-0000-0000-0000-000000000103"):
			testUsers = append(testUsers,
				testResident{lastName: "Кузнецов", firstName: "Дмитрий", middleName: "Игоревич", phone: fmt.Sprintf("7999%06d", phoneCounter), roomID: roomID},
				testResident{lastName: "Попов", firstName: "Алексей", middleName: "Владимирович", phone: fmt.Sprintf("7999%06d", phoneCounter+1), roomID: roomID},
				testResident{lastName: "Соколов", firstName: "Никита", middleName: "Павлович", phone: fmt.Sprintf("7999%06d", phoneCounter+2), roomID: roomID, debt: 5000.00, debtDesc: "Просроченная задолженность (апрель-июнь)"},
			)
			phoneCounter += 3
		}
	}

	for _, tr := range testUsers {
		var count int64
		db.Model(&domain.User{}).Where("phone = ?", tr.phone).Count(&count)
		if count > 0 {
			continue
		}

		user := domain.User{
			FirstName:      tr.firstName,
			LastName:       tr.lastName,
			MiddleName:     tr.middleName,
			Phone:          tr.phone,
			Platform:       "max",
			PlatformUserID: fmt.Sprintf("seed_%s", tr.phone),
			EISVerified:    true,
			EISPersonID:    fmt.Sprintf("seed_%d", phoneCounter),
			PersonType:     "student",
		}
		if err := db.Create(&user).Error; err != nil {
			continue
		}

		resident := domain.Resident{
			UserID:      user.ID,
			DormitoryID: dormitoryID,
			RoomID:      tr.roomID,
		}
		db.Create(&resident)

		if tr.debt > 0 {
			debt := domain.ResidentFinancialDebt{
				ResidentID:  resident.ID,
				Amount:      tr.debt,
				Period:      "2026",
				Description: tr.debtDesc,
				Source:      "EIS",
			}
			db.Create(&debt)
		}
	}

	log.Printf("SeedResidents: created %d test residents across %d rooms", len(testUsers), len(rooms))
}

// SeedContent creates chat_links and reference_materials that depend on
// a valid employee FK. Must be called AFTER SeedDormitories AND SeedTestUser.
func SeedContent(db *gorm.DB) {
	dormitoryID := uuid.MustParse("a0000000-0000-0000-0000-000000000001")

	// Check if already seeded
	var count int64
	db.Model(&domain.ChatLink{}).Where("dormitory_id = ?", dormitoryID).Count(&count)
	if count > 0 {
		log.Println("SeedContent: data already exists, skipping")
		return
	}

	// Resolve a valid employee for CreatedBy FK
	var createdBy domain.Employee
	if err := db.Where("role = ?", "commandant").First(&createdBy).Error; err != nil {
		log.Printf("SeedContent: no commandant employee found for FK, using uuid.Nil")
		createdBy.ID = uuid.Nil
	}

	chatLinks := []domain.ChatLink{
		{ID: uuid.MustParse("fb000000-0000-0000-0000-000000000001"), DormitoryID: dormitoryID, LinkType: "dormitory", Platform: "telegram", URL: "https://t.me/dorm8_chat", Title: "Чат общежития №8", IsActive: true, CreatedBy: createdBy.ID},
		{ID: uuid.MustParse("fb000000-0000-0000-0000-000000000002"), DormitoryID: dormitoryID, FloorID: ptrUUID(uuid.MustParse("b0000000-0000-0000-0000-000000000001")), LinkType: "floor", Platform: "telegram", URL: "https://t.me/dorm8_floor1", Title: "Чат 1 этажа", IsActive: true, CreatedBy: createdBy.ID},
	}
	for _, l := range chatLinks {
		db.Create(&l)
	}
	log.Println("Created chat links")

	refMaterials := []domain.ReferenceMaterial{
		{ID: uuid.MustParse("fc000000-0000-0000-0000-000000000001"), DormitoryID: dormitoryID, Name: "Правила проживания", Description: "Правила внутреннего распорядка общежития. Запрещено: курение, распитие алкоголя, шум после 23:00.", Category: "Правила", Ordinal: 1, CreatedBy: createdBy.ID},
		{ID: uuid.MustParse("fc000000-0000-0000-0000-000000000002"), DormitoryID: dormitoryID, Name: "Как пользоваться прачкой", Description: "Запись на стирку производится через бота в разделе «Стирка». Одна стирка — 60 минут. Отмена — не позднее чем за 15 минут до начала.", Category: "Прачка", Ordinal: 2, CreatedBy: createdBy.ID},
		{ID: uuid.MustParse("fc000000-0000-0000-0000-000000000003"), DormitoryID: dormitoryID, Name: "График дежурств", Description: "Дежурства назначаются автоматически в начале каждого месяца. Комнаты дежурят по очереди. За невыполнение — штрафное дежурство.", Category: "Дежурства", Ordinal: 3, CreatedBy: createdBy.ID},
	}
	for _, m := range refMaterials {
		db.Create(&m)
	}
	log.Println("Created reference materials")
}

func ptrUUID(id uuid.UUID) *uuid.UUID { return &id }
