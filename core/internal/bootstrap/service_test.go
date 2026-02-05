package bootstrap

import (
	"encoding/json"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// Mock NATS for testing?
// Ideally we use a real NATS server or a mock interface.
// For this environment, we might skip NATS test if no server is running,
// OR we can refactor Service to take an interface.
// For now, let's test the logic we can control, or assume extensive integration tests later.
// Actually, let's verify SQL interactions.

func TestService_HandleAnnouncement(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// We can't easily trigger the NATS callback without mocking NATS or refactoring.
	// We will refactor Service to have a public 'ProcessAnnouncement' method
	// to make it testable without NATS.

	s := &Service{db: db}

	payload := []byte(`{"id": "node-1", "type": "gpu-worker", "specs": {"ram": "16gb"}}`)

	// Expectation
	mock.ExpectExec("INSERT INTO nodes").
		WithArgs("node-1", "gpu-worker", json.RawMessage(`{"ram": "16gb"}`)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := s.processAnnouncement(payload); err != nil {
		t.Errorf("error processing announcement: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
