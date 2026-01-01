package in

import "jiaa-server-core/internal/output/domain"

// SabotageExecutorUseCase 사보타주 실행을 위한 Driving Port
// Dev 4(Core Decision Service)로부터 명령을 받아 실행
type SabotageExecutorUseCase interface {
	// ExecuteSabotage 사보타주 명령 실행
	// Dev 1(물리 제어), Dev 3(화면 제어)에 명령 전달
	ExecuteSabotage(cmd domain.SabotageCommand) (*domain.ExecutionResult, error)
}
