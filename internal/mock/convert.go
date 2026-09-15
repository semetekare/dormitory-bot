package mock

import "github.com/dormitory-bot/internal/infrastructure/rabbitmq"

// ── Conversion helpers: mock types → rabbitmq types ──

func (p *PersonData) ToRabbitMQ() *rabbitmq.PersonInfo {
	if p == nil {
		return nil
	}
	return &rabbitmq.PersonInfo{
		PersonID:   p.PersonID,
		LastName:   p.LastName,
		FirstName:  p.FirstName,
		MiddleName: p.MiddleName,
		Phone:      p.Phone,
	}
}

func (e *EmployeeData) ToRabbitMQ() *rabbitmq.EmployeeInfo {
	if e == nil {
		return nil
	}
	return &rabbitmq.EmployeeInfo{
		EmployeeID:   e.EmployeeID,
		PositionName: e.PositionName,
		Department:   e.Department,
		BuildingID:   e.BuildingID,
	}
}

func (s *StudentData) ToRabbitMQ() *rabbitmq.StudentInfo {
	if s == nil {
		return nil
	}
	return &rabbitmq.StudentInfo{
		BuildingID:   s.BuildingID,
		BuildingName: s.BuildingName,
		RoomName:     s.RoomName,
		RoomCapacity: uint(s.RoomCapacity),
	}
}

func (c *ContractData) ToRabbitMQ() *rabbitmq.ContractInfo {
	if c == nil {
		return nil
	}
	return &rabbitmq.ContractInfo{
		Number:  c.Number,
		DateIn:  c.DateIn,
		DateEnd: c.DateEnd,
	}
}
