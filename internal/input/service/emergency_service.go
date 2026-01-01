package service

import (
	"log"

	"jiaa-server-core/internal/input/port/out"
)

// EmergencyService Emergency 상황 처리 서비스
// Dev 6가 EMERGENCY(비명+에러) 상태를 선언했을 때:
// 1. 하던 일을 멈추고
// 2. Dev 5(AI)에게 로그 분석 요청
// 3. 결과를 Dev 3에게 전달
type EmergencyService struct {
	intelligencePort out.IntelligencePort
	screenPort       out.ScreenControlPort
}

// NewEmergencyService EmergencyService 생성자 (DI)
func NewEmergencyService(
	intelligencePort out.IntelligencePort,
	screenPort out.ScreenControlPort,
) *EmergencyService {
	return &EmergencyService{
		intelligencePort: intelligencePort,
		screenPort:       screenPort,
	}
}

// HandleEmergency Emergency 상황 처리
func (s *EmergencyService) HandleEmergency(clientID string, errorLog string, screamText string) error {
	log.Printf("[EMERGENCY] 🚨 Emergency triggered! Client: %s", clientID)
	log.Printf("[EMERGENCY] ErrorLog length: %d, ScreamText: %s", len(errorLog), screamText)

	// 1. Dev 5 (Intelligence Worker)에게 즉시 로그 분석 요청
	log.Printf("[EMERGENCY] Requesting AI analysis from Dev 5...")
	markdown, err := s.intelligencePort.RequestLogAnalysis(clientID, errorLog, screamText)
	if err != nil {
		log.Printf("[EMERGENCY] ❌ Failed to get AI analysis: %v", err)
		// 실패해도 기본 응급 메시지는 보냄
		markdown = generateFallbackEmergencyMessage(errorLog, screamText)
	}

	log.Printf("[EMERGENCY] AI analysis received, length: %d", len(markdown))

	// 2. Dev 3 (Screen Controller)에게 결과 전달
	log.Printf("[EMERGENCY] Sending AI result to Dev 3...")
	if err := s.screenPort.SendAIResult(clientID, markdown); err != nil {
		log.Printf("[EMERGENCY] ❌ Failed to send to screen controller: %v", err)
		return err
	}

	log.Printf("[EMERGENCY] ✅ Emergency handled successfully for client: %s", clientID)
	return nil
}

// generateFallbackEmergencyMessage AI 분석 실패 시 기본 응급 메시지 생성
func generateFallbackEmergencyMessage(errorLog string, screamText string) string {
	return `# 🚨 응급 상황 감지

## 상황
에러가 감지되었습니다. AI 분석을 수행할 수 없습니다.

## 에러 로그
` + "```\n" + truncateString(errorLog, 500) + "\n```" + `

## 권장 조치
1. 에러 메시지를 확인하세요
2. 최근 변경사항을 되돌려보세요
3. 필요시 동료에게 도움을 요청하세요
`
}

// truncateString 문자열을 최대 길이로 자르기
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "... (truncated)"
}
