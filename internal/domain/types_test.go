package domain

import "testing"

func TestThreadStateValidation(t *testing.T) {
	valid := []ThreadState{
		ThreadStateActive,
		ThreadStateWaiting,
		ThreadStateEscalated,
		ThreadStateResolved,
		ThreadStateDormant,
	}

	for _, state := range valid {
		if !state.IsValid() {
			t.Fatalf("expected state %q to be valid", state)
		}
	}

	if ThreadState("BROKEN").IsValid() {
		t.Fatal("expected unknown thread state to be invalid")
	}
}

func TestMessageDirectionValidation(t *testing.T) {
	valid := []MessageDirection{
		MessageDirectionInbound,
		MessageDirectionOutbound,
	}

	for _, direction := range valid {
		if !direction.IsValid() {
			t.Fatalf("expected direction %q to be valid", direction)
		}
	}

	if MessageDirection("sideways").IsValid() {
		t.Fatal("expected unknown message direction to be invalid")
	}
}

func TestAgentStatusValidation(t *testing.T) {
	valid := []AgentStatus{
		AgentStatusActive,
		AgentStatusPaused,
		AgentStatusArchived,
	}

	for _, status := range valid {
		if !status.IsValid() {
			t.Fatalf("expected status %q to be valid", status)
		}
	}

	if AgentStatus("deleted").IsValid() {
		t.Fatal("expected unknown agent status to be invalid")
	}
}
