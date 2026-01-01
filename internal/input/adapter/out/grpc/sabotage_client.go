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

// SabotageCommandAdapter 기존 SabotageCommandService gRPC 클라이언트 (Driven Adapter)
// CommandPort 인터페이스 구현
type SabotageCommandAdapter struct {
	conn    *grpc.ClientConn
	client  proto.SabotageCommandServiceClient
	address string
}

// NewSabotageCommandAdapter SabotageCommandAdapter 생성자
func NewSabotageCommandAdapter(address string) (*SabotageCommandAdapter, error) {
	conn, err := grpc.NewClient(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return &SabotageCommandAdapter{
		conn:    conn,
		client:  proto.NewSabotageCommandServiceClient(conn),
		address: address,
	}, nil
}

// NewSabotageCommandAdapterLazy 지연 연결 생성자
func NewSabotageCommandAdapterLazy(address string) *SabotageCommandAdapter {
	return &SabotageCommandAdapter{
		address: address,
	}
}

// Connect gRPC 연결 수립
func (a *SabotageCommandAdapter) Connect() error {
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
	a.client = proto.NewSabotageCommandServiceClient(conn)
	log.Printf("[SABOTAGE_CMD] Connected to: %s", a.address)
	return nil
}

// SendSabotage 사보타주 명령 전송 (CommandPort 구현)
func (a *SabotageCommandAdapter) SendSabotage(cmd domain.SabotageAction) error {
	if a.conn == nil {
		if err := a.Connect(); err != nil {
			log.Printf("[SABOTAGE_CMD] Failed to connect: %v", err)
			return err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &proto.SabotageRequest{
		ClientId:   cmd.ClientID,
		ActionType: string(cmd.ActionType),
		Intensity:  int32(cmd.Intensity),
		Message:    cmd.Message,
	}

	log.Printf("[SABOTAGE_CMD] Sending sabotage command: Client: %s, Action: %s, Intensity: %d",
		cmd.ClientID, cmd.ActionType, cmd.Intensity)

	resp, err := a.client.ExecuteSabotage(ctx, req)
	if err != nil {
		log.Printf("[SABOTAGE_CMD] gRPC call failed: %v", err)
		return err
	}

	if !resp.Success {
		log.Printf("[SABOTAGE_CMD] Sabotage command failed: %s", resp.ErrorCode)
	} else {
		log.Printf("[SABOTAGE_CMD] Sabotage command executed successfully")
	}

	return nil
}

// Close 연결 종료
func (a *SabotageCommandAdapter) Close() error {
	if a.conn != nil {
		return a.conn.Close()
	}
	return nil
}
