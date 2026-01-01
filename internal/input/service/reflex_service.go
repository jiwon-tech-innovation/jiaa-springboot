package service

import (
	"log"

	"jiaa-server-core/internal/input/domain"
	"jiaa-server-core/internal/input/port/out"
)

// ReflexService 속도가 생명인 즉각 반응 처리 서비스
// URL=Blacklist → 즉시 Sabotage 명령 (점수 계산 기다릴 시간 없음)
// 일반 트래픽 → Kafka로 Dev 6에 릴레이
type ReflexService struct {
	blacklistPort out.BlacklistPort
	commandPort   out.CommandPort
	dataRelayPort out.DataRelayPort
}

// NewReflexService ReflexService 생성자 (DI)
func NewReflexService(
	blacklistPort out.BlacklistPort,
	commandPort out.CommandPort,
	dataRelayPort out.DataRelayPort,
) *ReflexService {
	return &ReflexService{
		blacklistPort: blacklistPort,
		commandPort:   commandPort,
		dataRelayPort: dataRelayPort,
	}
}

// ProcessActivity 클라이언트 활동을 처리하고 필요시 즉각 반응
// Blacklist URL/App인 경우 즉시 SabotageAction 반환
// 일반 트래픽은 Kafka로 릴레이 후 nil 반환
func (s *ReflexService) ProcessActivity(activity domain.ClientActivity) (*domain.SabotageAction, error) {
	// 1. URL 블랙리스트 체크 (즉각 차단)
	if activity.IsURLActivity() && s.blacklistPort.IsBlacklisted(activity.URL) {
		log.Printf("[REFLEX] Blacklisted URL detected: %s, Client: %s", activity.URL, activity.ClientID)

		action := domain.NewSabotageAction(activity.ClientID, domain.ActionBlockURL).
			WithTargetURL(activity.URL).
			WithIntensity(10). // 최고 강도
			WithMessage("차단된 URL에 접근하였습니다.")

		// 즉시 차단 명령 전송
		if err := s.commandPort.SendSabotage(*action); err != nil {
			log.Printf("[REFLEX] Failed to send sabotage command: %v", err)
			return nil, err
		}

		return action, nil
	}

	// 2. App 블랙리스트 체크 (즉각 차단)
	if activity.IsAppActivity() && s.blacklistPort.IsAppBlacklisted(activity.AppName) {
		log.Printf("[REFLEX] Blacklisted App detected: %s, Client: %s", activity.AppName, activity.ClientID)

		action := domain.NewSabotageAction(activity.ClientID, domain.ActionCloseApp).
			WithTargetApp(activity.AppName).
			WithIntensity(10).
			WithMessage("차단된 앱을 실행하였습니다.")

		if err := s.commandPort.SendSabotage(*action); err != nil {
			log.Printf("[REFLEX] Failed to send sabotage command: %v", err)
			return nil, err
		}

		return action, nil
	}

	// 3. 일반 트래픽 → Dev 6으로 릴레이 (분석용)
	log.Printf("[REFLEX] Normal activity, relaying to analyzer: Client: %s, Type: %s",
		activity.ClientID, activity.ActivityType)

	if err := s.dataRelayPort.RelayToAnalyzer(activity); err != nil {
		log.Printf("[REFLEX] Failed to relay to analyzer: %v", err)
		return nil, err
	}

	return nil, nil
}
