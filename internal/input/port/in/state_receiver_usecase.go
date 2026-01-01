package in

import "jiaa-server-core/internal/input/domain"

// StateReceiverUseCase Dev 6 상태 수신을 위한 Driving Port
// Kafka Consumer가 호출하는 유스케이스
type StateReceiverUseCase interface {
	// HandleStateChange Dev 6에서 받은 상태 변화를 처리
	// 상태에 따라 Dev 1(물리 제어), Dev 3(화면 제어)에 명령 전송
	HandleStateChange(cmd domain.StateCommand) error
}
