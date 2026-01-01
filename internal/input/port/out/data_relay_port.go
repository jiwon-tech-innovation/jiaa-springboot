package out

import "jiaa-server-core/internal/input/domain"

// DataRelayPort 클라이언트 데이터를 Kafka로 릴레이하기 위한 Driven Port
// Dev 6(분석 서비스)에게 데이터 전송
type DataRelayPort interface {
	// RelayToAnalyzer 클라이언트 활동을 Dev 6(분석기)로 릴레이
	RelayToAnalyzer(activity domain.ClientActivity) error
}
