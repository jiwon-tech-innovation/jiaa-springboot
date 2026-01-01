package in

// EmergencyUseCase Emergency 상황 처리를 위한 Driving Port
// Dev 6가 EMERGENCY 상태를 전송하면 호출됨
type EmergencyUseCase interface {
	// HandleEmergency Emergency 상황 처리
	// 1. 하던 일 중단
	// 2. Dev 5에게 로그 분석 요청
	// 3. 결과를 Dev 3에게 전달
	HandleEmergency(clientID string, errorLog string, screamText string) error
}
