package domain

import "time"

// ExecutionStatus 실행 상태
type ExecutionStatus string

const (
	StatusPending   ExecutionStatus = "PENDING"
	StatusExecuting ExecutionStatus = "EXECUTING"
	StatusSuccess   ExecutionStatus = "SUCCESS"
	StatusFailed    ExecutionStatus = "FAILED"
	StatusPartial   ExecutionStatus = "PARTIAL" // 일부만 성공
)

// ExecutionResult 사보타주 명령 실행 결과
type ExecutionResult struct {
	CommandID      string           // 원본 명령 ID
	ClientID       string           // 대상 클라이언트 ID
	Status         ExecutionStatus  // 실행 상태
	PhysicalResult *ComponentResult // 물리 제어 결과
	ScreenResult   *ComponentResult // 화면 제어 결과
	StartTime      time.Time        // 시작 시간
	EndTime        time.Time        // 종료 시간
	ErrorMessage   string           // 에러 메시지 (실패시)
}

// ComponentResult 개별 컴포넌트 실행 결과
type ComponentResult struct {
	Success   bool   // 성공 여부
	ErrorCode string // 에러 코드
	Message   string // 결과 메시지
	Latency   int64  // 실행 시간 (밀리초)
}

// NewExecutionResult ExecutionResult 생성자
func NewExecutionResult(commandID, clientID string) *ExecutionResult {
	return &ExecutionResult{
		CommandID: commandID,
		ClientID:  clientID,
		Status:    StatusPending,
		StartTime: time.Now(),
	}
}

// SetPhysicalResult 물리 제어 결과 설정
func (r *ExecutionResult) SetPhysicalResult(success bool, errorCode, message string) {
	r.PhysicalResult = &ComponentResult{
		Success:   success,
		ErrorCode: errorCode,
		Message:   message,
	}
	r.updateStatus()
}

// SetScreenResult 화면 제어 결과 설정
func (r *ExecutionResult) SetScreenResult(success bool, errorCode, message string) {
	r.ScreenResult = &ComponentResult{
		Success:   success,
		ErrorCode: errorCode,
		Message:   message,
	}
	r.updateStatus()
}

// Complete 실행 완료 처리
func (r *ExecutionResult) Complete() {
	r.EndTime = time.Now()
	r.updateStatus()
}

// updateStatus 상태 업데이트
func (r *ExecutionResult) updateStatus() {
	physicalOK := r.PhysicalResult == nil || r.PhysicalResult.Success
	screenOK := r.ScreenResult == nil || r.ScreenResult.Success

	if physicalOK && screenOK {
		if r.PhysicalResult != nil || r.ScreenResult != nil {
			r.Status = StatusSuccess
		}
	} else if !physicalOK && !screenOK {
		r.Status = StatusFailed
	} else {
		r.Status = StatusPartial
	}
}

// GetDuration 실행 시간 반환 (밀리초)
func (r *ExecutionResult) GetDuration() int64 {
	if r.EndTime.IsZero() {
		return time.Since(r.StartTime).Milliseconds()
	}
	return r.EndTime.Sub(r.StartTime).Milliseconds()
}
