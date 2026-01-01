package service

import (
	"log"

	"jiaa-server-core/internal/input/domain"
	portin "jiaa-server-core/internal/input/port/in"
	"jiaa-server-core/internal/input/port/out"
)

// CommandRouterService Dev 6 → Dev 1, Dev 3 명령 라우팅 서비스
// Dev 6에서 받은 상태(SLEEPING, THINKING, EMERGENCY 등)를 수신하면
// 적절한 서비스/컨트롤러에게 실행 명령 전송
type CommandRouterService struct {
	physicalPort     out.PhysicalControlPort
	screenPort       out.ScreenControlPort
	emergencyHandler portin.EmergencyUseCase // Emergency 처리 위임
}

// NewCommandRouterService CommandRouterService 생성자 (DI)
func NewCommandRouterService(
	physicalPort out.PhysicalControlPort,
	screenPort out.ScreenControlPort,
) *CommandRouterService {
	return &CommandRouterService{
		physicalPort: physicalPort,
		screenPort:   screenPort,
	}
}

// SetEmergencyHandler Emergency 핸들러 설정 (순환 의존성 방지용 Setter)
func (s *CommandRouterService) SetEmergencyHandler(handler portin.EmergencyUseCase) {
	s.emergencyHandler = handler
}

// HandleStateChange Dev 6에서 받은 상태 변화를 처리
// 상태에 따라 적절한 액션 수행
func (s *CommandRouterService) HandleStateChange(cmd domain.StateCommand) error {
	log.Printf("[COMMAND_ROUTER] State change received: Client: %s, State: %s, Priority: %d",
		cmd.ClientID, cmd.State, cmd.Priority)

	// 1. THINKING 상태: 건드리지 않음 (Score > 80)
	if cmd.IsThinking() {
		log.Printf("[COMMAND_ROUTER] 🧠 THINKING state - Do not disturb. Client: %s", cmd.ClientID)
		return nil
	}

	// 2. EMERGENCY 상태: 즉시 Dev 5에게 분석 요청 (Audio > 90dB)
	if cmd.IsEmergency() {
		log.Printf("[COMMAND_ROUTER] 🚨 EMERGENCY detected! Delegating to EmergencyService...")
		if s.emergencyHandler != nil {
			// Payload에서 errorLog, screamText 추출
			errorLog := string(cmd.Payload)
			screamText := "Help!" // TODO: 실제로는 Payload 파싱 필요
			return s.emergencyHandler.HandleEmergency(cmd.ClientID, errorLog, screamText)
		}
		log.Printf("[COMMAND_ROUTER] ⚠️ EmergencyHandler not set, falling through to normal handling")
	}

	// 3. 일반 상태: Dev 1, Dev 3에 명령 라우팅
	action := cmd.ToSabotageAction()

	if cmd.RequiresImmediateAction() {
		log.Printf("[COMMAND_ROUTER] ⚡ Immediate action required for state: %s", cmd.State)
	}

	// Dev 1 (물리 제어) 명령 전송
	if err := s.physicalPort.SendToPhysicalController(*action); err != nil {
		log.Printf("[COMMAND_ROUTER] Failed to send to physical controller: %v", err)
		// 물리 제어 실패해도 화면 제어는 시도
	}

	// Dev 3 (화면 제어) 명령 전송
	if err := s.screenPort.SendToScreenController(*action); err != nil {
		log.Printf("[COMMAND_ROUTER] Failed to send to screen controller: %v", err)
		return err
	}

	log.Printf("[COMMAND_ROUTER] ✅ Commands routed successfully for client: %s", cmd.ClientID)
	return nil
}
