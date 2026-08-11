package service_test

import (
	"testing"
	"time"

	"suporter-backend/internal/domain"
	"suporter-backend/internal/service"
)



func TestSSEBroker_ProjectIsolation(t *testing.T) {
	broker := service.NewSSEBroker()

	proj1Chan := broker.Subscribe("prj_111")
	proj2Chan := broker.Subscribe("prj_222")

	defer broker.Unsubscribe("prj_111", proj1Chan)
	defer broker.Unsubscribe("prj_222", proj2Chan)

	alert1 := domain.Alert{
		ID:      1,
		Name:    "Alice",
		Message: "Hello Project 1!",
	}

	broker.Broadcast("prj_111", alert1)

	select {
	case received := <-proj1Chan:
		if received.Name != "Alice" {
			t.Errorf("Expected alert for Alice, got %s", received.Name)
		}
	case <-time.After(100 * time.Millisecond):
		t.Errorf("Timed out waiting for alert on project 1")
	}

	select {
	case unexpected := <-proj2Chan:
		t.Errorf("Project 2 received unexpected alert meant for Project 1: %v", unexpected)
	case <-time.After(50 * time.Millisecond):
		// Success! Project 2 isolated from Project 1.
	}
}
