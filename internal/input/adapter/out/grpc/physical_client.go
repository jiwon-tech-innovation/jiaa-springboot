package grpc

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"jiaa-server-core/internal/input/domain"
	"jiaa-server-core/pkg/proto"
)

// PhysicalControlAdapter Dev 1(물리 제어) gRPC 클라이언트 (Driven Adapter)
type PhysicalControlAdapter struct {
	conn    *grpc.ClientConn
	client  proto.PhysicalControlServiceClient
	address string
}

// NewPhysicalControlAdapter PhysicalControlAdapter 생성자
func NewPhysicalControlAdapter(address string) (*PhysicalControlAdapter, error) {
	conn, err := grpc.NewClient(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return &PhysicalControlAdapter{
		conn:    conn,
		client:  proto.NewPhysicalControlServiceClient(conn),
		address: address,
	}, nil
}

// NewPhysicalControlAdapterLazy 지연 연결 생성자 (연결 없이 생성)
func NewPhysicalControlAdapterLazy(address string) *PhysicalControlAdapter {
	return &PhysicalControlAdapter{
		address: address,
	}
}

// Connect gRPC 연결 수립
func (a *PhysicalControlAdapter) Connect() error {
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
	log.Printf("[PHYSICAL_CONTROL] Connected to Dev 1: %s", a.address)
	return nil
}

// SendToPhysicalController 물리 제어 명령 전송 (gRPC → Dev 1)
func (a *PhysicalControlAdapter) SendToPhysicalController(cmd domain.SabotageAction) error {
	if a.conn == nil {
		if err := a.Connect(); err != nil {
			log.Printf("[PHYSICAL_CONTROL] Failed to connect: %v", err)
			return err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// ActionType → PhysicalActionType 변환
	actionType := mapToPhysicalActionType(cmd.ActionType)

	req := &proto.PhysicalCommandRequest{
		ClientId:   cmd.ClientID,
		ActionType: actionType,
		Intensity:  int32(cmd.Intensity),
		Message:    cmd.Message,
	}

	log.Printf("[PHYSICAL_CONTROL] Sending command to Dev 1: Client: %s, Action: %v",
		cmd.ClientID, actionType)

	resp, err := a.client.ExecuteCommand(ctx, req)
	if err != nil {
		log.Printf("[PHYSICAL_CONTROL] gRPC call failed: %v", err)
		return err
	}

	if !resp.Success {
		log.Printf("[PHYSICAL_CONTROL] Command failed: %s", resp.ErrorCode)
	} else {
		log.Printf("[PHYSICAL_CONTROL] Command executed successfully")
	}

	return nil
}

// Close 연결 종료
func (a *PhysicalControlAdapter) Close() error {
	if a.conn != nil {
		return a.conn.Close()
	}
	return nil
}

// mapToPhysicalActionType ActionType → PhysicalActionType 변환
func mapToPhysicalActionType(action domain.ActionType) proto.PhysicalActionType {
	switch action {
	case domain.ActionCloseApp:
		return proto.PhysicalActionType_CLOSE_APP
	case domain.ActionMinimizeAll:
		return proto.PhysicalActionType_MINIMIZE_ALL
	case domain.ActionSleepScreen:
		return proto.PhysicalActionType_MOUSE_LOCK
	default:
		return proto.PhysicalActionType_PHYSICAL_ACTION_UNKNOWN
	}
}
