package domain

import "time"

// CommandState Dev 6에서 전송하는 상태 명령
type CommandState string

const (
	StateSleeping   CommandState = "SLEEPING"   // Score < 30 → 수면 모드 (화면 끔)
	StateAwake      CommandState = "AWAKE"      // 깨어남 (정상 상태)
	StateDistracted CommandState = "DISTRACTED" // 집중력 분산 감지
	StateFocused    CommandState = "FOCUSED"    // 집중 상태
	StateThinking   CommandState = "THINKING"   // Score > 80 → 건드리지 마
	StateWarning    CommandState = "WARNING"    // 경고 상태
	StateCritical   CommandState = "CRITICAL"   // 심각한 상태 (즉시 조치 필요)
	StateEmergency  CommandState = "EMERGENCY"  // Audio > 90dB → 구해줘! (비명+에러)
)

// StateCommand Dev 6에서 수신하는 상태 명령
type StateCommand struct {
	ClientID  string       // 대상 클라이언트 ID
	State     CommandState // 명령 상태
	Payload   []byte       // 추가 페이로드 데이터
	Timestamp time.Time    // 명령 발생 시간
	Priority  int          // 우선순위 (1-10, 높을수록 긴급)
}

// NewStateCommand StateCommand 생성자
func NewStateCommand(clientID string, state CommandState) *StateCommand {
	return &StateCommand{
		ClientID:  clientID,
		State:     state,
		Timestamp: time.Now(),
		Priority:  5, // 기본 우선순위
	}
}

// WithPayload 페이로드 설정
func (s *StateCommand) WithPayload(payload []byte) *StateCommand {
	s.Payload = payload
	return s
}

// WithPriority 우선순위 설정
func (s *StateCommand) WithPriority(priority int) *StateCommand {
	if priority < 1 {
		priority = 1
	}
	if priority > 10 {
		priority = 10
	}
	s.Priority = priority
	return s
}

// RequiresImmediateAction 즉각적인 조치가 필요한 상태인지 확인
func (s *StateCommand) RequiresImmediateAction() bool {
	return s.State == StateSleeping || s.State == StateCritical || s.State == StateEmergency || s.Priority >= 8
}

// IsEmergency Emergency 상태인지 확인
func (s *StateCommand) IsEmergency() bool {
	return s.State == StateEmergency
}

// IsThinking Thinking 상태(건드리지 마)인지 확인
func (s *StateCommand) IsThinking() bool {
	return s.State == StateThinking
}

// ToSabotageAction 상태를 SabotageAction으로 변환
func (s *StateCommand) ToSabotageAction() *SabotageAction {
	action := NewSabotageAction(s.ClientID, stateToActionType(s.State))
	action.WithIntensity(s.Priority)
	return action
}

// stateToActionType CommandState를 ActionType으로 매핑
func stateToActionType(state CommandState) ActionType {
	switch state {
	case StateSleeping:
		return ActionSleepScreen
	case StateAwake:
		return ActionWakeScreen
	case StateDistracted, StateWarning, StateCritical:
		return ActionMinimizeAll
	default:
		return ActionWakeScreen
	}
}
