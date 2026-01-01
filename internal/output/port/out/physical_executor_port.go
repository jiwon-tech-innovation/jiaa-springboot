package out

import "jiaa-server-core/internal/output/domain"

// PhysicalExecutorPort Dev 1(물리 제어) 실행을 위한 Driven Port
type PhysicalExecutorPort interface {
	// Execute 물리 제어 명령 실행
	// 마우스 감도 저하, 창 흔들기, 앱 종료 등
	Execute(cmd domain.SabotageCommand) (*domain.ComponentResult, error)
}
