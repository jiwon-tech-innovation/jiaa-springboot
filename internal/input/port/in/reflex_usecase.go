package in

import "jiaa-server-core/internal/input/domain"

// ReflexUseCase Reflex 반응 처리를 위한 Driving Port
// 속도가 생명인 즉각 반응 처리 (Blacklist URL → 즉시 차단)
type ReflexUseCase interface {
	// ProcessActivity 클라이언트 활동을 처리하고 필요시 즉각 반응
	// Blacklist URL인 경우 즉시 SabotageAction 반환
	// 일반 트래픽은 Kafka로 릴레이 후 nil 반환
	ProcessActivity(activity domain.ClientActivity) (*domain.SabotageAction, error)
}
