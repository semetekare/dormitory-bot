package mock

import "context"

// EISDataProvider — абстракция над источником данных ЕИС.
// Реализации: MockProvider (мок-режим), RMQProvider (реальный адаптер через RabbitMQ).
type EISDataProvider interface {
	// VerifyPerson ищет человека по нормализованному телефону в ЕИС.
	// Возвращает связанные Person, Employee, Student, Contract или nil если не найден.
	VerifyPerson(ctx context.Context, phone string) (*PersonData, *EmployeeData, *StudentData, *ContractData, error)

	// GetDormitories возвращает все общежития из ЕИС.
	GetDormitories(ctx context.Context) ([]DormitoryData, error)

	// GetEmployeesByDormitory возвращает сотрудников общежития.
	GetEmployeesByDormitory(ctx context.Context, eisBuildingID int) ([]EmployeeData, error)

	// GetResidentsByBuilding возвращает проживающих в здании.
	GetResidentsByBuilding(ctx context.Context, eisBuildingID int) ([]ResidentData, error)
}

// PersonData — данные о человеке из ЕИС.
type PersonData struct {
	PersonID   uint   // eis persons.id
	FirstName  string
	LastName   string
	MiddleName string
	Phone      string
}

// EmployeeData — данные о сотруднике из ЕИС.
type EmployeeData struct {
	EmployeeID   uint
	PersonID     uint
	PositionName string // dolzn.name
	Department   string // departments.name
	BuildingID   *uint  // locality → auditories.building_id
	Phone        string
}

// StudentData — данные о студенте из ЕИС.
type StudentData struct {
	StudentID    uint
	PersonID     uint
	BuildingID   uint   // locality → auditories.building_id
	BuildingName string // buildings.name
	RoomName     string // auditories.name
	RoomCapacity int    // auditories.capacity
}

// ContractData — данные о контракте из ЕИС.
type ContractData struct {
	ContractID uint
	PersonID   uint
	Number     string
	DateIn     string // YYYY-MM-DD
	DateEnd    string // YYYY-MM-DD
}

// DormitoryData — данные об общежитии из ЕИС.
type DormitoryData struct {
	EISCode    string // "EIS-BLD-N"
	BuildingID uint
	Name       string
	Address    string
	Floors     []FloorData
}

// FloorData — данные об этаже.
type FloorData struct {
	Number int
	Wings  []string
	Rooms  []RoomData
}

// RoomData — данные о комнате.
type RoomData struct {
	Number   string
	Capacity int
}

// ResidentData — данные о проживающем.
type ResidentData struct {
	PersonID       uint
	FirstName      string
	LastName       string
	MiddleName     string
	Phone          string
	BuildingID     uint
	RoomName       string
	ContractNumber string
	ContractStart  string
	ContractEnd    string
	DebtAmount     float64
	DebtPeriod     string
	DebtDesc       string
}
