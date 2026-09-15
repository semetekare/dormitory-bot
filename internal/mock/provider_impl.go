package mock

import (
	"context"
)

// MockProvider реализует EISDataProvider на заготовленных данных.
type MockProvider struct {
	data *mockDataset
}

// NewMockProvider создаёт новый провайдер с заготовленным dataset.
func NewMockProvider() *MockProvider {
	return &MockProvider{data: &MockDataset}
}

// VerifyPerson ищет человека по нормализованному телефону.
// Возвращает связанные Person, Employee, Student, Contract. nil для отсутствующих записей.
func (p *MockProvider) VerifyPerson(ctx context.Context, phone string) (*PersonData, *EmployeeData, *StudentData, *ContractData, error) {
	_ = ctx

	// 1. Найти person по нормализованному телефону
	var person *PersonData
	for i := range p.data.Persons {
		if p.data.Persons[i].Phone == phone {
			person = &p.data.Persons[i]
			break
		}
	}
	if person == nil {
		return nil, nil, nil, nil, nil
	}

	// 2. Найти employee по PersonID
	var employee *EmployeeData
	for i := range p.data.Employees {
		if p.data.Employees[i].PersonID == person.PersonID {
			employee = &p.data.Employees[i]
			break
		}
	}

	// 3. Найти student по PersonID
	var student *StudentData
	for i := range p.data.Students {
		if p.data.Students[i].PersonID == person.PersonID {
			student = &p.data.Students[i]
			break
		}
	}

	// 4. Найти contract по PersonID
	var contract *ContractData
	for i := range p.data.Contracts {
		if p.data.Contracts[i].PersonID == person.PersonID {
			contract = &p.data.Contracts[i]
			break
		}
	}

	return person, employee, student, contract, nil
}

// GetDormitories возвращает все общежития.
func (p *MockProvider) GetDormitories(ctx context.Context) ([]DormitoryData, error) {
	_ = ctx
	result := make([]DormitoryData, len(p.data.Dormitories))
	copy(result, p.data.Dormitories)
	return result, nil
}

// GetEmployeesByDormitory возвращает сотрудников конкретного общежития по EIS BuildingID.
func (p *MockProvider) GetEmployeesByDormitory(ctx context.Context, eisBuildingID int) ([]EmployeeData, error) {
	_ = ctx
	var result []EmployeeData
	for _, e := range p.data.Employees {
		if e.BuildingID != nil && int(*e.BuildingID) == eisBuildingID {
			result = append(result, e)
		}
	}
	return result, nil
}

// GetResidentsByBuilding возвращает проживающих в здании по EIS BuildingID.
func (p *MockProvider) GetResidentsByBuilding(ctx context.Context, eisBuildingID int) ([]ResidentData, error) {
	_ = ctx
	var result []ResidentData
	for _, s := range p.data.Students {
		if int(s.BuildingID) != eisBuildingID {
			continue
		}
		// Найти связанного person
		var person *PersonData
		for i := range p.data.Persons {
			if p.data.Persons[i].PersonID == s.PersonID {
				person = &p.data.Persons[i]
				break
			}
		}
		if person == nil {
			continue
		}
		// Найти contract
		var contract *ContractData
		for i := range p.data.Contracts {
			if p.data.Contracts[i].PersonID == s.PersonID {
				contract = &p.data.Contracts[i]
				break
			}
		}

		rd := ResidentData{
			PersonID:   person.PersonID,
			FirstName:  person.FirstName,
			LastName:   person.LastName,
			MiddleName: person.MiddleName,
			Phone:      person.Phone,
			BuildingID: s.BuildingID,
			RoomName:   s.RoomName,
		}
		if contract != nil {
			rd.ContractNumber = contract.Number
			rd.ContractStart = contract.DateIn
			rd.ContractEnd = contract.DateEnd
		}
		if debt, ok := p.data.Debts[s.PersonID]; ok {
			rd.DebtAmount = debt.Amount
			rd.DebtPeriod = debt.Period
			rd.DebtDesc = debt.Desc
		}
		result = append(result, rd)
	}
	return result, nil
}
