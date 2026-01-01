package domain

import (
	"testing"
)

func TestNewSabotageCommand(t *testing.T) {
	cmd := NewSabotageCommand("client-123", SabotageScreenGlitch)

	if cmd.ClientID != "client-123" {
		t.Errorf("Expected ClientID 'client-123', got '%s'", cmd.ClientID)
	}
	if cmd.SabotageType != SabotageScreenGlitch {
		t.Errorf("Expected SabotageType SCREEN_GLITCH, got '%s'", cmd.SabotageType)
	}
	if cmd.TargetType != TargetBoth {
		t.Errorf("Expected TargetType BOTH, got '%s'", cmd.TargetType)
	}
	if cmd.Intensity != 5 {
		t.Errorf("Expected default Intensity 5, got %d", cmd.Intensity)
	}
	if cmd.DurationMs != 3000 {
		t.Errorf("Expected default DurationMs 3000, got %d", cmd.DurationMs)
	}
}

func TestSabotageCommand_RequiresPhysicalControl(t *testing.T) {
	tests := []struct {
		name       string
		targetType TargetType
		expected   bool
	}{
		{"PHYSICAL requires physical", TargetPhysical, true},
		{"BOTH requires physical", TargetBoth, true},
		{"SCREEN does not require physical", TargetScreen, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewSabotageCommand("client", SabotageScreenGlitch).WithTarget(tt.targetType)
			if got := cmd.RequiresPhysicalControl(); got != tt.expected {
				t.Errorf("RequiresPhysicalControl() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSabotageCommand_RequiresScreenControl(t *testing.T) {
	tests := []struct {
		name       string
		targetType TargetType
		expected   bool
	}{
		{"SCREEN requires screen", TargetScreen, true},
		{"BOTH requires screen", TargetBoth, true},
		{"PHYSICAL does not require screen", TargetPhysical, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewSabotageCommand("client", SabotageScreenGlitch).WithTarget(tt.targetType)
			if got := cmd.RequiresScreenControl(); got != tt.expected {
				t.Errorf("RequiresScreenControl() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewExecutionResult(t *testing.T) {
	result := NewExecutionResult("cmd-123", "client-123")

	if result.CommandID != "cmd-123" {
		t.Errorf("Expected CommandID 'cmd-123', got '%s'", result.CommandID)
	}
	if result.ClientID != "client-123" {
		t.Errorf("Expected ClientID 'client-123', got '%s'", result.ClientID)
	}
	if result.Status != StatusPending {
		t.Errorf("Expected Status PENDING, got '%s'", result.Status)
	}
}

func TestExecutionResult_SetResults(t *testing.T) {
	result := NewExecutionResult("cmd-123", "client-123")

	// 물리 제어 성공
	result.SetPhysicalResult(true, "", "OK")
	if result.PhysicalResult == nil {
		t.Error("Expected PhysicalResult to be set")
	}
	if !result.PhysicalResult.Success {
		t.Error("Expected PhysicalResult.Success to be true")
	}

	// 화면 제어 성공
	result.SetScreenResult(true, "", "OK")
	if result.ScreenResult == nil {
		t.Error("Expected ScreenResult to be set")
	}
	if result.Status != StatusSuccess {
		t.Errorf("Expected Status SUCCESS, got '%s'", result.Status)
	}
}

func TestExecutionResult_PartialSuccess(t *testing.T) {
	result := NewExecutionResult("cmd-123", "client-123")

	result.SetPhysicalResult(true, "", "OK")
	result.SetScreenResult(false, "ERROR", "Failed")

	if result.Status != StatusPartial {
		t.Errorf("Expected Status PARTIAL, got '%s'", result.Status)
	}
}

func TestExecutionResult_AllFailed(t *testing.T) {
	result := NewExecutionResult("cmd-123", "client-123")

	result.SetPhysicalResult(false, "ERROR1", "Failed")
	result.SetScreenResult(false, "ERROR2", "Failed")

	if result.Status != StatusFailed {
		t.Errorf("Expected Status FAILED, got '%s'", result.Status)
	}
}

func TestExecutionResult_GetDuration(t *testing.T) {
	result := NewExecutionResult("cmd-123", "client-123")

	// Complete 호출 전에는 현재 시간까지의 duration
	duration := result.GetDuration()
	if duration < 0 {
		t.Error("Expected non-negative duration")
	}

	result.Complete()

	// Complete 후에는 고정된 duration
	duration2 := result.GetDuration()
	if duration2 < 0 {
		t.Error("Expected non-negative duration after complete")
	}
}
