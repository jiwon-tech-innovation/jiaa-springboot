package service

import (
	"log"
	"sync"

	"jiaa-server-core/internal/output/domain"
	"jiaa-server-core/internal/output/port/out"
)

// SabotageExecutorService 사보타주 명령 실행 서비스
// Dev 1(물리 제어)과 Dev 3(화면 제어)에 병렬로 명령 전달
type SabotageExecutorService struct {
	physicalExecutor out.PhysicalExecutorPort
	screenExecutor   out.ScreenExecutorPort
}

// NewSabotageExecutorService SabotageExecutorService 생성자 (DI)
func NewSabotageExecutorService(
	physicalExecutor out.PhysicalExecutorPort,
	screenExecutor out.ScreenExecutorPort,
) *SabotageExecutorService {
	return &SabotageExecutorService{
		physicalExecutor: physicalExecutor,
		screenExecutor:   screenExecutor,
	}
}

// ExecuteSabotage 사보타주 명령 실행
func (s *SabotageExecutorService) ExecuteSabotage(cmd domain.SabotageCommand) (*domain.ExecutionResult, error) {
	log.Printf("[SABOTAGE_EXECUTOR] Executing sabotage: Client: %s, Type: %s, Intensity: %d",
		cmd.ClientID, cmd.SabotageType, cmd.Intensity)

	result := domain.NewExecutionResult(cmd.ID, cmd.ClientID)
	result.Status = domain.StatusExecuting

	var wg sync.WaitGroup
	var physicalResult, screenResult *domain.ComponentResult
	var physicalErr, screenErr error

	// 물리 제어 실행 (Dev 1)
	if cmd.RequiresPhysicalControl() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			log.Printf("[SABOTAGE_EXECUTOR] Executing physical control...")
			physicalResult, physicalErr = s.physicalExecutor.Execute(cmd)
			if physicalErr != nil {
				log.Printf("[SABOTAGE_EXECUTOR] Physical control failed: %v", physicalErr)
			}
		}()
	}

	// 화면 제어 실행 (Dev 3)
	if cmd.RequiresScreenControl() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			log.Printf("[SABOTAGE_EXECUTOR] Executing screen control...")
			screenResult, screenErr = s.screenExecutor.Execute(cmd)
			if screenErr != nil {
				log.Printf("[SABOTAGE_EXECUTOR] Screen control failed: %v", screenErr)
			}
		}()
	}

	// 모든 실행 완료 대기
	wg.Wait()

	// 결과 설정
	if physicalResult != nil {
		result.SetPhysicalResult(physicalResult.Success, physicalResult.ErrorCode, physicalResult.Message)
	}
	if screenResult != nil {
		result.SetScreenResult(screenResult.Success, screenResult.ErrorCode, screenResult.Message)
	}

	result.Complete()

	log.Printf("[SABOTAGE_EXECUTOR] Execution completed: Status: %s, Duration: %dms",
		result.Status, result.GetDuration())

	return result, nil
}
