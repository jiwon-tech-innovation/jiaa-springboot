package out

import "jiaa-server-core/internal/input/domain"

// PhysicalControlPort Dev 1(물리 제어)과 통신하기 위한 Driven Port
type PhysicalControlPort interface {
	// SendToPhysicalController 물리 제어 명령 전송 (gRPC → Dev 1)
	SendToPhysicalController(cmd domain.SabotageAction) error
}
