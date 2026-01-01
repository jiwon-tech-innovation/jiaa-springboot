package service

import (
	"log"

	"jiaa-server-core/internal/input/port/out"
)

// SolutionRouterService Dev 5 AI 결과 → Dev 3 전달 서비스
// Dev 5에서 받은 RAG 결과(Markdown)를 Dev 3(화면 제어)에게 라우팅
type SolutionRouterService struct {
	screenPort out.ScreenControlPort
}

// NewSolutionRouterService SolutionRouterService 생성자 (DI)
func NewSolutionRouterService(screenPort out.ScreenControlPort) *SolutionRouterService {
	return &SolutionRouterService{
		screenPort: screenPort,
	}
}

// RouteAIResult Dev 5의 AI 결과를 Dev 3(화면 제어)에 라우팅
func (s *SolutionRouterService) RouteAIResult(clientID string, markdown string) error {
	log.Printf("[SOLUTION_ROUTER] Routing AI result to screen controller: Client: %s, Content length: %d",
		clientID, len(markdown))

	if err := s.screenPort.SendAIResult(clientID, markdown); err != nil {
		log.Printf("[SOLUTION_ROUTER] Failed to send AI result: %v", err)
		return err
	}

	log.Printf("[SOLUTION_ROUTER] AI result routed successfully for client: %s", clientID)
	return nil
}
