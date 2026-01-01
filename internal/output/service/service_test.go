package service

import (
	"testing"

	"jiaa-server-core/internal/output/domain"
)

// MockPhysicalExecutorPort 테스트용 Mock
type MockPhysicalExecutorPort struct {
	ExecutedCommands []domain.SabotageCommand
	ShouldFail       bool
}

func (m *MockPhysicalExecutorPort) Execute(cmd domain.SabotageCommand) (*domain.ComponentResult, error) {
	m.ExecutedCommands = append(m.ExecutedCommands, cmd)
	return &domain.ComponentResult{
		Success:   !m.ShouldFail,
		ErrorCode: "",
		Message:   "Executed",
		Latency:   10,
	}, nil
}

// MockScreenExecutorPort 테스트용 Mock
type MockScreenExecutorPort struct {
	ExecutedCommands []domain.SabotageCommand
	ShouldFail       bool
}

func (m *MockScreenExecutorPort) Execute(cmd domain.SabotageCommand) (*domain.ComponentResult, error) {
	m.ExecutedCommands = append(m.ExecutedCommands, cmd)
	return &domain.ComponentResult{
		Success:   !m.ShouldFail,
		ErrorCode: "",
		Message:   "Executed",
		Latency:   15,
	}, nil
}

func TestSabotageExecutorService_ExecuteSabotage_Both(t *testing.T) {
	physicalExecutor := &MockPhysicalExecutorPort{}
	screenExecutor := &MockScreenExecutorPort{}

	service := NewSabotageExecutorService(physicalExecutor, screenExecutor)

	cmd := domain.NewSabotageCommand("client-123", domain.SabotageScreenGlitch).
		WithTarget(domain.TargetBoth).
		WithIntensity(7)

	result, err := service.ExecuteSabotage(*cmd)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if result.Status != domain.StatusSuccess {
		t.Errorf("Expected Status SUCCESS, got '%s'", result.Status)
	}
	if len(physicalExecutor.ExecutedCommands) != 1 {
		t.Errorf("Expected 1 physical execution, got %d", len(physicalExecutor.ExecutedCommands))
	}
	if len(screenExecutor.ExecutedCommands) != 1 {
		t.Errorf("Expected 1 screen execution, got %d", len(screenExecutor.ExecutedCommands))
	}
}

func TestSabotageExecutorService_ExecuteSabotage_PhysicalOnly(t *testing.T) {
	physicalExecutor := &MockPhysicalExecutorPort{}
	screenExecutor := &MockScreenExecutorPort{}

	service := NewSabotageExecutorService(physicalExecutor, screenExecutor)

	cmd := domain.NewSabotageCommand("client-123", domain.SabotageMinimizeAll).
		WithTarget(domain.TargetPhysical)

	result, err := service.ExecuteSabotage(*cmd)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result.Status != domain.StatusSuccess {
		t.Errorf("Expected Status SUCCESS, got '%s'", result.Status)
	}
	if len(physicalExecutor.ExecutedCommands) != 1 {
		t.Errorf("Expected 1 physical execution, got %d", len(physicalExecutor.ExecutedCommands))
	}
	if len(screenExecutor.ExecutedCommands) != 0 {
		t.Errorf("Expected 0 screen executions, got %d", len(screenExecutor.ExecutedCommands))
	}
}

func TestSabotageExecutorService_ExecuteSabotage_ScreenOnly(t *testing.T) {
	physicalExecutor := &MockPhysicalExecutorPort{}
	screenExecutor := &MockScreenExecutorPort{}

	service := NewSabotageExecutorService(physicalExecutor, screenExecutor)

	cmd := domain.NewSabotageCommand("client-123", domain.SabotageRedFlash).
		WithTarget(domain.TargetScreen)

	result, err := service.ExecuteSabotage(*cmd)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result.Status != domain.StatusSuccess {
		t.Errorf("Expected Status SUCCESS, got '%s'", result.Status)
	}
	if len(physicalExecutor.ExecutedCommands) != 0 {
		t.Errorf("Expected 0 physical executions, got %d", len(physicalExecutor.ExecutedCommands))
	}
	if len(screenExecutor.ExecutedCommands) != 1 {
		t.Errorf("Expected 1 screen execution, got %d", len(screenExecutor.ExecutedCommands))
	}
}

func TestSabotageExecutorService_ExecuteSabotage_PartialFailure(t *testing.T) {
	physicalExecutor := &MockPhysicalExecutorPort{ShouldFail: true}
	screenExecutor := &MockScreenExecutorPort{ShouldFail: false}

	service := NewSabotageExecutorService(physicalExecutor, screenExecutor)

	cmd := domain.NewSabotageCommand("client-123", domain.SabotageScreenGlitch).
		WithTarget(domain.TargetBoth)

	result, err := service.ExecuteSabotage(*cmd)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result.Status != domain.StatusPartial {
		t.Errorf("Expected Status PARTIAL, got '%s'", result.Status)
	}
}

func TestSabotageExecutorService_ExecuteSabotage_AllFailed(t *testing.T) {
	physicalExecutor := &MockPhysicalExecutorPort{ShouldFail: true}
	screenExecutor := &MockScreenExecutorPort{ShouldFail: true}

	service := NewSabotageExecutorService(physicalExecutor, screenExecutor)

	cmd := domain.NewSabotageCommand("client-123", domain.SabotageScreenGlitch).
		WithTarget(domain.TargetBoth)

	result, err := service.ExecuteSabotage(*cmd)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result.Status != domain.StatusFailed {
		t.Errorf("Expected Status FAILED, got '%s'", result.Status)
	}
}
