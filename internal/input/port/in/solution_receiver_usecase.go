package in

// SolutionReceiverUseCase Dev 5 AI 결과 수신을 위한 Driving Port
// Dev 5에서 RAG 결과(Markdown)를 받아 Dev 3에게 전달
type SolutionReceiverUseCase interface {
	// RouteAIResult Dev 5의 AI 결과를 Dev 3(화면 제어)에 라우팅
	RouteAIResult(clientID string, markdown string) error
}
