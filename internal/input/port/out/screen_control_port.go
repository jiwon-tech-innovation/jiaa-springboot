package out

import "jiaa-server-core/internal/input/domain"

// ScreenControlPort Dev 3(화면 제어)와 통신하기 위한 Driven Port
type ScreenControlPort interface {
	// SendToScreenController 화면 제어 명령 전송 (gRPC → Dev 3)
	SendToScreenController(cmd domain.SabotageAction) error

	// SendAIResult AI 결과(Markdown) 전송 (Solution Router → Dev 3)
	SendAIResult(clientID string, markdown string) error
}
