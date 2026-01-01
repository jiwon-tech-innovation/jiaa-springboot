package domain

import (
	"testing"
	"time"
)

func TestNewSabotageAction(t *testing.T) {
	action := NewSabotageAction("client-123", ActionBlockURL)

	if action.ClientID != "client-123" {
		t.Errorf("Expected ClientID 'client-123', got '%s'", action.ClientID)
	}
	if action.ActionType != ActionBlockURL {
		t.Errorf("Expected ActionType BLOCK_URL, got '%s'", action.ActionType)
	}
	if action.Intensity != 5 {
		t.Errorf("Expected default Intensity 5, got %d", action.Intensity)
	}
}

func TestSabotageAction_WithIntensity(t *testing.T) {
	action := NewSabotageAction("client-123", ActionBlockURL)

	// 정상 범위
	action.WithIntensity(8)
	if action.Intensity != 8 {
		t.Errorf("Expected Intensity 8, got %d", action.Intensity)
	}

	// 최소값 미만
	action.WithIntensity(0)
	if action.Intensity != 1 {
		t.Errorf("Expected Intensity 1 (min), got %d", action.Intensity)
	}

	// 최대값 초과
	action.WithIntensity(15)
	if action.Intensity != 10 {
		t.Errorf("Expected Intensity 10 (max), got %d", action.Intensity)
	}
}

func TestSabotageAction_Builder(t *testing.T) {
	action := NewSabotageAction("client-123", ActionCloseApp).
		WithIntensity(7).
		WithMessage("테스트 메시지").
		WithTargetApp("Steam")

	if action.Intensity != 7 {
		t.Errorf("Expected Intensity 7, got %d", action.Intensity)
	}
	if action.Message != "테스트 메시지" {
		t.Errorf("Expected Message '테스트 메시지', got '%s'", action.Message)
	}
	if action.TargetApp != "Steam" {
		t.Errorf("Expected TargetApp 'Steam', got '%s'", action.TargetApp)
	}
}

func TestNewClientActivity(t *testing.T) {
	activity := NewClientActivity("client-123", ActivityURLVisit)

	if activity.ClientID != "client-123" {
		t.Errorf("Expected ClientID 'client-123', got '%s'", activity.ClientID)
	}
	if activity.ActivityType != ActivityURLVisit {
		t.Errorf("Expected ActivityType URL_VISIT, got '%s'", activity.ActivityType)
	}
	if activity.Metadata == nil {
		t.Error("Expected Metadata to be initialized")
	}
}

func TestClientActivity_IsURLActivity(t *testing.T) {
	tests := []struct {
		name     string
		activity ClientActivity
		expected bool
	}{
		{
			name: "URL Visit with URL",
			activity: ClientActivity{
				ActivityType: ActivityURLVisit,
				URL:          "https://youtube.com",
			},
			expected: true,
		},
		{
			name: "URL Visit without URL",
			activity: ClientActivity{
				ActivityType: ActivityURLVisit,
				URL:          "",
			},
			expected: false,
		},
		{
			name: "App Open is not URL activity",
			activity: ClientActivity{
				ActivityType: ActivityAppOpen,
				URL:          "https://youtube.com",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.activity.IsURLActivity(); got != tt.expected {
				t.Errorf("IsURLActivity() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewStateCommand(t *testing.T) {
	cmd := NewStateCommand("client-123", StateSleeping)

	if cmd.ClientID != "client-123" {
		t.Errorf("Expected ClientID 'client-123', got '%s'", cmd.ClientID)
	}
	if cmd.State != StateSleeping {
		t.Errorf("Expected State SLEEPING, got '%s'", cmd.State)
	}
	if cmd.Priority != 5 {
		t.Errorf("Expected default Priority 5, got %d", cmd.Priority)
	}
	if cmd.Timestamp.IsZero() {
		t.Error("Expected Timestamp to be set")
	}
}

func TestStateCommand_RequiresImmediateAction(t *testing.T) {
	tests := []struct {
		name     string
		state    CommandState
		priority int
		expected bool
	}{
		{"SLEEPING requires immediate", StateSleeping, 5, true},
		{"CRITICAL requires immediate", StateCritical, 5, true},
		{"EMERGENCY requires immediate", StateEmergency, 5, true},
		{"FOCUSED does not require immediate", StateFocused, 5, false},
		{"High priority requires immediate", StateFocused, 8, true},
		{"Low priority AWAKE does not require", StateAwake, 3, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewStateCommand("client", tt.state).WithPriority(tt.priority)
			if got := cmd.RequiresImmediateAction(); got != tt.expected {
				t.Errorf("RequiresImmediateAction() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestStateCommand_IsEmergency(t *testing.T) {
	emergencyCmd := NewStateCommand("client", StateEmergency)
	if !emergencyCmd.IsEmergency() {
		t.Error("Expected IsEmergency() to be true for EMERGENCY state")
	}

	normalCmd := NewStateCommand("client", StateAwake)
	if normalCmd.IsEmergency() {
		t.Error("Expected IsEmergency() to be false for AWAKE state")
	}
}

func TestStateCommand_ToSabotageAction(t *testing.T) {
	cmd := NewStateCommand("client-123", StateSleeping).WithPriority(7)
	action := cmd.ToSabotageAction()

	if action.ClientID != "client-123" {
		t.Errorf("Expected ClientID 'client-123', got '%s'", action.ClientID)
	}
	if action.ActionType != ActionSleepScreen {
		t.Errorf("Expected ActionType SLEEP_SCREEN, got '%s'", action.ActionType)
	}
	if action.Intensity != 7 {
		t.Errorf("Expected Intensity 7, got %d", action.Intensity)
	}
}

func TestClientActivity_WithTimestamp(t *testing.T) {
	activity := NewClientActivity("client", ActivityURLVisit)
	testTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	activity.WithTimestamp(testTime)

	if !activity.Timestamp.Equal(testTime) {
		t.Errorf("Expected Timestamp %v, got %v", testTime, activity.Timestamp)
	}
}
