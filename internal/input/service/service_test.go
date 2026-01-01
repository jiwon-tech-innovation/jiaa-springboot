package service

import (
	"testing"

	"jiaa-server-core/internal/input/domain"
)

// MockBlacklistPort 테스트용 Mock
type MockBlacklistPort struct {
	blacklistedURLs map[string]bool
	blacklistedApps map[string]bool
}

func NewMockBlacklistPort() *MockBlacklistPort {
	return &MockBlacklistPort{
		blacklistedURLs: map[string]bool{
			"youtube.com": true,
			"netflix.com": true,
		},
		blacklistedApps: map[string]bool{
			"Steam":   true,
			"Discord": true,
		},
	}
}

func (m *MockBlacklistPort) IsBlacklisted(url string) bool {
	return m.blacklistedURLs[url]
}

func (m *MockBlacklistPort) IsAppBlacklisted(appName string) bool {
	return m.blacklistedApps[appName]
}

// MockCommandPort 테스트용 Mock
type MockCommandPort struct {
	SentCommands []domain.SabotageAction
}

func (m *MockCommandPort) SendSabotage(cmd domain.SabotageAction) error {
	m.SentCommands = append(m.SentCommands, cmd)
	return nil
}

// MockDataRelayPort 테스트용 Mock
type MockDataRelayPort struct {
	RelayedActivities []domain.ClientActivity
}

func (m *MockDataRelayPort) RelayToAnalyzer(activity domain.ClientActivity) error {
	m.RelayedActivities = append(m.RelayedActivities, activity)
	return nil
}

func TestReflexService_ProcessActivity_BlacklistedURL(t *testing.T) {
	blacklistPort := NewMockBlacklistPort()
	commandPort := &MockCommandPort{}
	dataRelayPort := &MockDataRelayPort{}

	service := NewReflexService(blacklistPort, commandPort, dataRelayPort)

	activity := domain.ClientActivity{
		ClientID:     "client-123",
		URL:          "youtube.com",
		ActivityType: domain.ActivityURLVisit,
	}

	action, err := service.ProcessActivity(activity)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if action == nil {
		t.Error("Expected SabotageAction for blacklisted URL")
	}
	if action.ActionType != domain.ActionBlockURL {
		t.Errorf("Expected ActionType BLOCK_URL, got '%s'", action.ActionType)
	}
	if len(commandPort.SentCommands) != 1 {
		t.Errorf("Expected 1 command sent, got %d", len(commandPort.SentCommands))
	}
	if len(dataRelayPort.RelayedActivities) != 0 {
		t.Error("Should not relay blacklisted URL to analyzer")
	}
}

func TestReflexService_ProcessActivity_NormalURL(t *testing.T) {
	blacklistPort := NewMockBlacklistPort()
	commandPort := &MockCommandPort{}
	dataRelayPort := &MockDataRelayPort{}

	service := NewReflexService(blacklistPort, commandPort, dataRelayPort)

	activity := domain.ClientActivity{
		ClientID:     "client-123",
		URL:          "stackoverflow.com",
		ActivityType: domain.ActivityURLVisit,
	}

	action, err := service.ProcessActivity(activity)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if action != nil {
		t.Error("Expected nil action for normal URL")
	}
	if len(commandPort.SentCommands) != 0 {
		t.Error("Should not send command for normal URL")
	}
	if len(dataRelayPort.RelayedActivities) != 1 {
		t.Errorf("Expected 1 activity relayed, got %d", len(dataRelayPort.RelayedActivities))
	}
}

func TestReflexService_ProcessActivity_BlacklistedApp(t *testing.T) {
	blacklistPort := NewMockBlacklistPort()
	commandPort := &MockCommandPort{}
	dataRelayPort := &MockDataRelayPort{}

	service := NewReflexService(blacklistPort, commandPort, dataRelayPort)

	activity := domain.ClientActivity{
		ClientID:     "client-123",
		AppName:      "Steam",
		ActivityType: domain.ActivityAppOpen,
	}

	action, err := service.ProcessActivity(activity)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if action == nil {
		t.Error("Expected SabotageAction for blacklisted app")
	}
	if action.ActionType != domain.ActionCloseApp {
		t.Errorf("Expected ActionType CLOSE_APP, got '%s'", action.ActionType)
	}
}

// MockPhysicalControlPort 테스트용 Mock
type MockPhysicalControlPort struct {
	SentCommands []domain.SabotageAction
}

func (m *MockPhysicalControlPort) SendToPhysicalController(cmd domain.SabotageAction) error {
	m.SentCommands = append(m.SentCommands, cmd)
	return nil
}

// MockScreenControlPort 테스트용 Mock
type MockScreenControlPort struct {
	SentCommands []domain.SabotageAction
	AIResults    []string
}

func (m *MockScreenControlPort) SendToScreenController(cmd domain.SabotageAction) error {
	m.SentCommands = append(m.SentCommands, cmd)
	return nil
}

func (m *MockScreenControlPort) SendAIResult(clientID string, markdown string) error {
	m.AIResults = append(m.AIResults, markdown)
	return nil
}

func TestCommandRouterService_HandleStateChange_Sleeping(t *testing.T) {
	physicalPort := &MockPhysicalControlPort{}
	screenPort := &MockScreenControlPort{}

	service := NewCommandRouterService(physicalPort, screenPort)

	cmd := domain.NewStateCommand("client-123", domain.StateSleeping)

	err := service.HandleStateChange(*cmd)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(physicalPort.SentCommands) != 1 {
		t.Errorf("Expected 1 physical command, got %d", len(physicalPort.SentCommands))
	}
	if len(screenPort.SentCommands) != 1 {
		t.Errorf("Expected 1 screen command, got %d", len(screenPort.SentCommands))
	}
}

func TestCommandRouterService_HandleStateChange_Thinking(t *testing.T) {
	physicalPort := &MockPhysicalControlPort{}
	screenPort := &MockScreenControlPort{}

	service := NewCommandRouterService(physicalPort, screenPort)

	cmd := domain.NewStateCommand("client-123", domain.StateThinking)

	err := service.HandleStateChange(*cmd)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// THINKING 상태는 아무 조치도 하지 않음
	if len(physicalPort.SentCommands) != 0 {
		t.Error("Should not send physical command for THINKING state")
	}
	if len(screenPort.SentCommands) != 0 {
		t.Error("Should not send screen command for THINKING state")
	}
}

func TestSolutionRouterService_RouteAIResult(t *testing.T) {
	screenPort := &MockScreenControlPort{}

	service := NewSolutionRouterService(screenPort)

	err := service.RouteAIResult("client-123", "# AI 분석 결과\n테스트입니다")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(screenPort.AIResults) != 1 {
		t.Errorf("Expected 1 AI result sent, got %d", len(screenPort.AIResults))
	}
}
