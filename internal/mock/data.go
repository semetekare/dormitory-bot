package mock

// ── Mock Dataset ──────────────────────────────────────────
//
// Полный набор тестовых данных для мок-режима:
// 2 общежития, 5 сотрудников, 10 студентов с контрактами и задолженностями.

var MockDataset = mockDataset{
	Dormitories: []DormitoryData{
		{
			EISCode:    "EIS-BLD-8",
			BuildingID: 8,
			Name:       "АОУ Студгородок общежитие 8",
			Address:    "ул. Лермонтова, д. 80",
			Floors: []FloorData{
				{
					Number: 1,
					Wings:  []string{"Левое", "Правое"},
					Rooms: []RoomData{
						{Number: "101", Capacity: 3},
						{Number: "102", Capacity: 2},
						{Number: "103", Capacity: 4},
					},
				},
				{
					Number: 2,
					Wings:  []string{"Левое", "Правое"},
					Rooms: []RoomData{
						{Number: "201", Capacity: 2},
						{Number: "202", Capacity: 3},
					},
				},
				{
					Number: 3,
					Rooms: []RoomData{
						{Number: "301", Capacity: 2},
					},
				},
			},
		},
		{
			EISCode:    "EIS-BLD-5",
			BuildingID: 5,
			Name:       "АОУ Студгородок общежитие 5",
			Address:    "ул. Лермонтова, д. 75",
			Floors: []FloorData{
				{
					Number: 1,
					Rooms: []RoomData{
						{Number: "101", Capacity: 3},
						{Number: "102", Capacity: 2},
					},
				},
				{
					Number: 2,
					Rooms: []RoomData{
						{Number: "201", Capacity: 2},
						{Number: "202", Capacity: 3},
					},
				},
			},
		},
	},

	// ── Persons (все люди в системе) ──
	Persons: []PersonData{
		// Employees
		{PersonID: 1001, FirstName: "Данил", LastName: "Высоких", MiddleName: "Александрович", Phone: "79025643215"},
		{PersonID: 1002, FirstName: "Елена", LastName: "Кузнецова", MiddleName: "Петровна", Phone: "79140001122"},
		{PersonID: 1003, FirstName: "Игорь", LastName: "Морозов", MiddleName: "Викторович", Phone: "79140001133"},
		{PersonID: 1004, FirstName: "Анна", LastName: "Фёдорова", MiddleName: "Сергеевна", Phone: "79140001144"},
		{PersonID: 1005, FirstName: "Павел", LastName: "Григорьев", MiddleName: "Денисович", Phone: "79140001155"},

		// Students (Dorm №8)
		{PersonID: 1010, FirstName: "Артём", LastName: "Иванов", MiddleName: "Сергеевич", Phone: "79001234501"},
		{PersonID: 1011, FirstName: "Максим", LastName: "Петров", MiddleName: "Алексеевич", Phone: "79001234502"},
		{PersonID: 1012, FirstName: "Анна", LastName: "Смирнова", MiddleName: "Дмитриевна", Phone: "79001234503"},
		{PersonID: 1013, FirstName: "Дмитрий", LastName: "Кузнецов", MiddleName: "Игоревич", Phone: "79001234504"},
		{PersonID: 1014, FirstName: "Алексей", LastName: "Попов", MiddleName: "Владимирович", Phone: "79001234505"},
		{PersonID: 1015, FirstName: "Никита", LastName: "Соколов", MiddleName: "Павлович", Phone: "79001234506"},
		{PersonID: 1016, FirstName: "Ольга", LastName: "Васильева", MiddleName: "Игоревна", Phone: "79001234507"},
		{PersonID: 1017, FirstName: "Сергей", LastName: "Белов", MiddleName: "Николаевич", Phone: "79001234508"},
		{PersonID: 1018, FirstName: "Ирина", LastName: "Морозова", MiddleName: "Александровна", Phone: "79001234509"},
		{PersonID: 1019, FirstName: "Владимир", LastName: "Ершов", MiddleName: "Петрович", Phone: "79001234510"},
	},

	// ── Employees ──
	Employees: []EmployeeData{
		// Dorm №8 staff
		{EmployeeID: 2001, PersonID: 1001, PositionName: "Заведующий общежитием", Department: "Администрация студгородка", BuildingID: uintPtr(8), Phone: "79025643215"},
		{EmployeeID: 2002, PersonID: 1002, PositionName: "Дежурный", Department: "Общежитие №8", BuildingID: uintPtr(8), Phone: "79140001122"},
		{EmployeeID: 2004, PersonID: 1004, PositionName: "Председатель студсовета", Department: "Общежитие №8", BuildingID: uintPtr(8), Phone: "79140001144"},
		{EmployeeID: 2005, PersonID: 1005, PositionName: "Староста", Department: "Общежитие №8", BuildingID: uintPtr(8), Phone: "79140001155"},

		// Dorm №5 staff
		{EmployeeID: 2003, PersonID: 1003, PositionName: "Заведующий общежитием", Department: "Общежитие №5", BuildingID: uintPtr(5), Phone: "79140001133"},
	},

	// ── Students ──
	Students: []StudentData{
		// Dorm №8 students
		{StudentID: 3001, PersonID: 1010, BuildingID: 8, BuildingName: "Общежитие 8", RoomName: "101", RoomCapacity: 3},
		{StudentID: 3002, PersonID: 1011, BuildingID: 8, BuildingName: "Общежитие 8", RoomName: "101", RoomCapacity: 3},
		{StudentID: 3003, PersonID: 1012, BuildingID: 8, BuildingName: "Общежитие 8", RoomName: "102", RoomCapacity: 2},
		{StudentID: 3004, PersonID: 1013, BuildingID: 8, BuildingName: "Общежитие 8", RoomName: "103", RoomCapacity: 4},
		{StudentID: 3005, PersonID: 1014, BuildingID: 8, BuildingName: "Общежитие 8", RoomName: "103", RoomCapacity: 4},
		{StudentID: 3006, PersonID: 1015, BuildingID: 8, BuildingName: "Общежитие 8", RoomName: "201", RoomCapacity: 2},
		{StudentID: 3007, PersonID: 1016, BuildingID: 8, BuildingName: "Общежитие 8", RoomName: "202", RoomCapacity: 3},

		// Dorm №5 students
		{StudentID: 3008, PersonID: 1017, BuildingID: 5, BuildingName: "Общежитие 5", RoomName: "101", RoomCapacity: 3},
		{StudentID: 3009, PersonID: 1018, BuildingID: 5, BuildingName: "Общежитие 5", RoomName: "102", RoomCapacity: 2},

		// Student without contract — person 1019 exists but NO contract
	},

	// ── Contracts ──
	Contracts: []ContractData{
		// Active students
		{ContractID: 4001, PersonID: 1010, Number: "A-2026-001", DateIn: "2025-09-01", DateEnd: "2026-08-31"},
		{ContractID: 4002, PersonID: 1011, Number: "A-2026-002", DateIn: "2025-09-01", DateEnd: "2026-08-31"},
		{ContractID: 4003, PersonID: 1012, Number: "A-2026-003", DateIn: "2025-09-01", DateEnd: "2026-08-31"},
		{ContractID: 4004, PersonID: 1013, Number: "A-2026-004", DateIn: "2025-09-01", DateEnd: "2026-08-31"},
		{ContractID: 4005, PersonID: 1014, Number: "A-2026-005", DateIn: "2025-09-01", DateEnd: "2026-08-31"},
		{ContractID: 4006, PersonID: 1015, Number: "A-2026-006", DateIn: "2025-09-01", DateEnd: "2026-08-31"},
		{ContractID: 4007, PersonID: 1016, Number: "A-2026-007", DateIn: "2025-09-01", DateEnd: "2026-08-31"},
		// Expired contract — date_end in the past
		{ContractID: 4008, PersonID: 1017, Number: "A-2025-008", DateIn: "2024-09-01", DateEnd: "2025-08-31"},
		// Active contract
		{ContractID: 4009, PersonID: 1018, Number: "A-2026-009", DateIn: "2025-09-01", DateEnd: "2026-08-31"},
		// Person 1019 — NO contract (contractless student)
	},

	// ── Resident debt ──
	// keyed by PersonID for lookup during sync loading
	Debts: map[uint]ResidentDebtData{
		1010: {Amount: 1500.00, Period: "2026-06", Desc: "Задолженность за проживание (июнь)"},
		1012: {Amount: 3200.00, Period: "2026-05/2026-06", Desc: "Задолженность за проживание (май-июнь)"},
		1015: {Amount: 5000.00, Period: "2026-04/2026-06", Desc: "Просроченная задолженность (апрель-июнь)"},
	},
}

// ── Types ────────────────────────────────────────────

type mockDataset struct {
	Dormitories []DormitoryData
	Persons     []PersonData
	Employees   []EmployeeData
	Students    []StudentData
	Contracts   []ContractData
	Debts       map[uint]ResidentDebtData
}

// ResidentDebtData — данные о задолженности проживающего.
type ResidentDebtData struct {
	Amount float64
	Period string
	Desc   string
}

func uintPtr(v uint) *uint { return &v }
