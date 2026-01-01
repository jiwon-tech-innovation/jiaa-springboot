package domain

// ActionType 사보타주 명령 유형
type ActionType string

const (
	ActionBlockURL    ActionType = "BLOCK_URL"
	ActionCloseApp    ActionType = "CLOSE_APP"
	ActionSleepScreen ActionType = "SLEEP_SCREEN"
	ActionWakeScreen  ActionType = "WAKE_SCREEN"
	ActionMinimizeAll ActionType = "MINIMIZE_ALL"
)

// SabotageAction 사보타주 명령을 나타내는 도메인 엔티티
// Dev 1(물리 제어) 및 Dev 3(화면 제어)에게 전송되는 명령
type SabotageAction struct {
	ClientID   string     // 대상 클라이언트 ID
	ActionType ActionType // 수행할 액션 유형
	Intensity  int        // 명령 강도 (1-10)
	Message    string     // 사용자에게 표시할 메시지
	TargetURL  string     // 차단 대상 URL (BLOCK_URL인 경우)
	TargetApp  string     // 종료 대상 앱 (CLOSE_APP인 경우)
}

// NewSabotageAction SabotageAction 생성자
func NewSabotageAction(clientID string, actionType ActionType) *SabotageAction {
	return &SabotageAction{
		ClientID:   clientID,
		ActionType: actionType,
		Intensity:  5, // 기본 강도
	}
}

// WithIntensity 강도 설정 (Builder 패턴)
func (s *SabotageAction) WithIntensity(intensity int) *SabotageAction {
	if intensity < 1 {
		intensity = 1
	}
	if intensity > 10 {
		intensity = 10
	}
	s.Intensity = intensity
	return s
}

// WithMessage 메시지 설정
func (s *SabotageAction) WithMessage(message string) *SabotageAction {
	s.Message = message
	return s
}

// WithTargetURL 대상 URL 설정
func (s *SabotageAction) WithTargetURL(url string) *SabotageAction {
	s.TargetURL = url
	return s
}

// WithTargetApp 대상 앱 설정
func (s *SabotageAction) WithTargetApp(app string) *SabotageAction {
	s.TargetApp = app
	return s
}
