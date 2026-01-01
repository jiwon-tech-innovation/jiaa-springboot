package out

import "jiaa-server-core/internal/output/domain"

// ScreenExecutorPort Dev 3(화면 제어) 실행을 위한 Driven Port
type ScreenExecutorPort interface {
	// Execute 화면 제어 명령 실행
	// 글리치, 붉은 점멸, 화면 흔들림, TTS 등
	Execute(cmd domain.SabotageCommand) (*domain.ComponentResult, error)
}
