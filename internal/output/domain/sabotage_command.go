package domain

import "time"

// SabotageType 사보타주 명령 유형
type SabotageType string

const (
	SabotageBlockURL     SabotageType = "BLOCK_URL"
	SabotageCloseApp     SabotageType = "CLOSE_APP"
	SabotageMinimizeAll  SabotageType = "MINIMIZE_ALL"
	SabotageMouseLock    SabotageType = "MOUSE_LOCK"
	SabotageScreenGlitch SabotageType = "SCREEN_GLITCH"
	SabotageRedFlash     SabotageType = "RED_FLASH"
	SabotageBlackScreen  SabotageType = "BLACK_SCREEN"
	SabotageWindowShake  SabotageType = "WINDOW_SHAKE"
	SabotageTTS          SabotageType = "TTS"
)

// TargetType 대상 유형
type TargetType string

const (
	TargetPhysical TargetType = "PHYSICAL" // Dev 1 물리 제어
	TargetScreen   TargetType = "SCREEN"   // Dev 3 화면 제어
	TargetBoth     TargetType = "BOTH"     // 둘 다
)

// SabotageCommand 사보타주 명령 도메인 엔티티
type SabotageCommand struct {
	ID           string       // 명령 고유 ID
	ClientID     string       // 대상 클라이언트 ID
	SabotageType SabotageType // 사보타주 유형
	TargetType   TargetType   // 대상 유형
	Intensity    int          // 강도 (1-10)
	DurationMs   int          // 지속 시간 (밀리초)
	Message      string       // 사용자에게 표시할 메시지
	Payload      []byte       // 추가 데이터
	Timestamp    time.Time    // 명령 생성 시간
	Priority     int          // 우선순위 (높을수록 긴급)
}

// NewSabotageCommand SabotageCommand 생성자
func NewSabotageCommand(clientID string, sabotageType SabotageType) *SabotageCommand {
	return &SabotageCommand{
		ClientID:     clientID,
		SabotageType: sabotageType,
		TargetType:   TargetBoth,
		Intensity:    5,
		DurationMs:   3000,
		Timestamp:    time.Now(),
		Priority:     5,
	}
}

// WithIntensity 강도 설정
func (c *SabotageCommand) WithIntensity(intensity int) *SabotageCommand {
	if intensity < 1 {
		intensity = 1
	}
	if intensity > 10 {
		intensity = 10
	}
	c.Intensity = intensity
	return c
}

// WithDuration 지속 시간 설정
func (c *SabotageCommand) WithDuration(durationMs int) *SabotageCommand {
	c.DurationMs = durationMs
	return c
}

// WithMessage 메시지 설정
func (c *SabotageCommand) WithMessage(message string) *SabotageCommand {
	c.Message = message
	return c
}

// WithTarget 대상 유형 설정
func (c *SabotageCommand) WithTarget(target TargetType) *SabotageCommand {
	c.TargetType = target
	return c
}

// WithPriority 우선순위 설정
func (c *SabotageCommand) WithPriority(priority int) *SabotageCommand {
	c.Priority = priority
	return c
}

// RequiresPhysicalControl 물리 제어가 필요한지 확인
func (c *SabotageCommand) RequiresPhysicalControl() bool {
	return c.TargetType == TargetPhysical || c.TargetType == TargetBoth
}

// RequiresScreenControl 화면 제어가 필요한지 확인
func (c *SabotageCommand) RequiresScreenControl() bool {
	return c.TargetType == TargetScreen || c.TargetType == TargetBoth
}
