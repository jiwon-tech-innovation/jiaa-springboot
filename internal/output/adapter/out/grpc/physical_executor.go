package grpc

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"jiaa-server-core/internal/output/domain"
	"jiaa-server-core/pkg/proto"
)

// PhysicalExecutorAdapter Dev 1(물리 제어) 실행 어댑터 (Driven Adapter)
type PhysicalExecutorAdapter struct {
	conn    *grpc.ClientConn
	client  proto.PhysicalControlServiceClient
	address string
}

// NewPhysicalExecutorAdapter PhysicalExecutorAdapter 생성자
func NewPhysicalExecutorAdapter(address string) (*PhysicalExecutorAdapter, error) {
	conn, err := grpc.NewClient(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return &PhysicalExecutorAdapter{
		conn:    conn,
		client:  proto.NewPhysicalControlServiceClient(conn),
		address: address,
	}, nil
}

// NewPhysicalExecutorAdapterLazy 지연 연결 생성자
func NewPhysicalExecutorAdapterLazy(address string) *PhysicalExecutorAdapter {
	return &PhysicalExecutorAdapter{
		address: address,
	}
}

// Connect gRPC 연결 수립
func (a *PhysicalExecutorAdapter) Connect() error {
	if a.conn != nil {
		return nil
	}

	conn, err := grpc.NewClient(a.address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return err
	}

	a.conn = conn
	a.client = proto.NewPhysicalControlServiceClient(conn)
	log.Printf("[PHYSICAL_EXECUTOR] Connected to Dev 1: %s", a.address)
	return nil
}

// Execute 물리 제어 명령 실행
func (a *PhysicalExecutorAdapter) Execute(cmd domain.SabotageCommand) (*domain.ComponentResult, error) {
	startTime := time.Now()

	if a.conn == nil {
		if err := a.Connect(); err != nil {
			log.Printf("[PHYSICAL_EXECUTOR] Failed to connect: %v", err)
			return &domain.ComponentResult{
				Success:   false,
				ErrorCode: "CONNECTION_ERROR",
				Message:   err.Error(),
			}, err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	actionType := mapSabotageToPhysicalAction(cmd.SabotageType)

	req := &proto.PhysicalCommandRequest{
		ClientId:   cmd.ClientID,
		ActionType: actionType,
		Intensity:  int32(cmd.Intensity),
		Message:    cmd.Message,
	}

	log.Printf("[PHYSICAL_EXECUTOR] Executing: Client: %s, Action: %v", cmd.ClientID, actionType)

	resp, err := a.client.ExecuteCommand(ctx, req)
	latency := time.Since(startTime).Milliseconds()

	if err != nil {
		log.Printf("[PHYSICAL_EXECUTOR] gRPC call failed: %v", err)
		return &domain.ComponentResult{
			Success:   false,
			ErrorCode: "GRPC_ERROR",
			Message:   err.Error(),
			Latency:   latency,
		}, err
	}

	return &domain.ComponentResult{
		Success:   resp.Success,
		ErrorCode: resp.ErrorCode,
		Message:   resp.Message,
		Latency:   latency,
	}, nil
}

// Close 연결 종료
func (a *PhysicalExecutorAdapter) Close() error {
	if a.conn != nil {
		return a.conn.Close()
	}
	return nil
}

// mapSabotageToPhysicalAction SabotageType → PhysicalActionType 변환
func mapSabotageToPhysicalAction(sabotageType domain.SabotageType) proto.PhysicalActionType {
	switch sabotageType {
	case domain.SabotageCloseApp:
		return proto.PhysicalActionType_CLOSE_APP
	case domain.SabotageMinimizeAll:
		return proto.PhysicalActionType_MINIMIZE_ALL
	case domain.SabotageMouseLock:
		return proto.PhysicalActionType_MOUSE_LOCK
	case domain.SabotageWindowShake:
		return proto.PhysicalActionType_WINDOW_SHAKE
	default:
		return proto.PhysicalActionType_PHYSICAL_ACTION_UNKNOWN
	}
}
