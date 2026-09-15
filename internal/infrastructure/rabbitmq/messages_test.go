package rabbitmq

import (
	"encoding/json"
	"testing"
)

func TestVerifyRequestSerialization(t *testing.T) {
	req := VerifyRequest{
		Version:       ContractVersion,
		CorrelationID: "test-correlation-id",
		Phone:         "79025643215",
		MaxUserID:     143554557,
		Platform:      "max",
		Timestamp:     "2026-07-22T10:00:00Z",
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	var got VerifyRequest
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Phone != "79025643215" {
		t.Errorf("Phone = %s, want 79025643215", got.Phone)
	}
	if got.CorrelationID != "test-correlation-id" {
		t.Errorf("CorrelationID = %s, want test-correlation-id", got.CorrelationID)
	}
	if got.MaxUserID != 143554557 {
		t.Errorf("MaxUserID = %d, want 143554557", got.MaxUserID)
	}
}

func TestVerifyResponseVersionValidation(t *testing.T) {
	resp := VerifyResponse{Version: 1}
	if !resp.ValidateVersion() {
		t.Error("version 1 should be valid")
	}
	resp.Version = 0
	if resp.ValidateVersion() {
		t.Error("version 0 should be invalid")
	}
	resp.Version = 2
	if resp.ValidateVersion() {
		t.Error("version 2 should be invalid")
	}
}

func TestVerifyResponseSerialization(t *testing.T) {
	resp := VerifyResponse{
		Version:       ContractVersion,
		CorrelationID: "abc123",
		Found:         true,
		Person: &PersonInfo{
			PersonID:   42,
			LastName:   "Ivanov",
			FirstName:  "Ivan",
			MiddleName: "Ivanovich",
			Phone:      "79025643215",
		},
		Student: &StudentInfo{
			BuildingID:   1,
			BuildingName: "Test Building",
			RoomName:     "101",
			RoomCapacity: 3,
			FloorID:      1,
			FloorName:    "1 этаж",
			StatusID:     3,
			StatusName:   "active",
		},
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	var got VerifyResponse
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if !got.Found {
		t.Error("should be found")
	}
	if got.Person == nil {
		t.Fatal("person should not be nil")
	}
	if got.Person.LastName != "Ivanov" {
		t.Errorf("LastName = %s, want Ivanov", got.Person.LastName)
	}
	if got.Student == nil {
		t.Fatal("student should not be nil")
	}
	if got.Student.RoomName != "101" {
		t.Errorf("RoomName = %s, want 101", got.Student.RoomName)
	}
}

func TestSyncResidentsPushVersionValidation(t *testing.T) {
	msg := SyncResidentsPush{Version: 1}
	if !msg.ValidateVersion() {
		t.Error("version 1 should be valid")
	}
	msg.Version = 2
	if msg.ValidateVersion() {
		t.Error("version 2 should be invalid")
	}
}

func TestSyncEmployeesPushVersionValidation(t *testing.T) {
	msg := SyncEmployeesPush{Version: 1}
	if !msg.ValidateVersion() {
		t.Error("version 1 should be valid")
	}
	msg.Version = 0
	if msg.ValidateVersion() {
		t.Error("version 0 should be invalid")
	}
}

func TestSyncDormitoriesPushVersionValidation(t *testing.T) {
	msg := SyncDormitoriesPush{Version: 1}
	if !msg.ValidateVersion() {
		t.Error("version 1 should be valid")
	}
	msg.Version = 3
	if msg.ValidateVersion() {
		t.Error("version 3 should be invalid")
	}
}
