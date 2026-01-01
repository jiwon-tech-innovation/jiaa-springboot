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

// ScreenControlAdapter Dev 3(화면 제어) gRPC 클라이언트 (Driven Adapter)
type ScreenControlAdapter struct {
	conn    *grpc.ClientConn
	client  proto.ScreenControlServiceClient
	address string
}

// NewScreenControlAdapter ScreenControlAdapter 생성자
func NewScreenControlAdapter(address string) (*ScreenControlAdapter, error) {
	conn, err := grpc.NewClient(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return &ScreenControlAdapter{
		conn:    conn,
		client:  proto.NewScreenControlServiceClient(conn),
		address: address,
	}, nil
}

// NewScreenControlAdapterLazy 지연 연결 생성자 (연결 없이 생성)
func NewScreenControlAdapterLazy(address string) *ScreenControlAdapter {
	return &ScreenControlAdapter{
		address: address,
	}
}

// Connect gRPC 연결 수립
func (a *ScreenControlAdapter) Connect() error {
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
	a.client = proto.NewScreenControlServiceClient(conn)
	log.Printf("[SCREEN_CONTROL] Connected to Dev 3: %s", a.address)
	return nil
}

// SendToScreenController 화면 제어 명령 전송 (gRPC → Dev 3)
func (a *ScreenControlAdapter) SendToScreenController(cmd domain.SabotageAction) error {
	if a.conn == nil {
		if err := a.Connect(); err != nil {
			log.Printf("[SCREEN_CONTROL] Failed to connect: %v", err)
			return err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// ActionType → VisualEffectType 변환
	effectType := mapToVisualEffectType(cmd.ActionType)

	req := &proto.VisualCommandRequest{
		ClientId:   cmd.ClientID,
		EffectType: effectType,
		Intensity:  int32(cmd.Intensity),
		DurationMs: 3000, // 3초 지속
		Message:    cmd.Message,
	}

	log.Printf("[SCREEN_CONTROL] Sending visual command to Dev 3: Client: %s, Effect: %v",
		cmd.ClientID, effectType)

	resp, err := a.client.ExecuteVisualCommand(ctx, req)
	if err != nil {
		log.Printf("[SCREEN_CONTROL] gRPC call failed: %v", err)
		return err
	}

	if !resp.Success {
		log.Printf("[SCREEN_CONTROL] Visual command failed: %s", resp.ErrorCode)
	} else {
		log.Printf("[SCREEN_CONTROL] Visual command executed successfully")
	}

	return nil
}

// SendAIResult AI 결과(Markdown) 전송 (Solution Router → Dev 3)
func (a *ScreenControlAdapter) SendAIResult(clientID string, markdown string) error {
	if a.conn == nil {
		if err := a.Connect(); err != nil {
			log.Printf("[SCREEN_CONTROL] Failed to connect: %v", err)
			return err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := &proto.AIResultRequest{
		ClientId:   clientID,
		Markdown:   markdown,
		Title:      "AI Analysis Result",
		ResultType: proto.AIResultType_ERROR_SOLUTION,
	}

	log.Printf("[SCREEN_CONTROL] Sending AI result to Dev 3: Client: %s, Length: %d",
		clientID, len(markdown))

	resp, err := a.client.DisplayAIResult(ctx, req)
	if err != nil {
		log.Printf("[SCREEN_CONTROL] gRPC call failed: %v", err)
		return err
	}

	if !resp.Success {
		log.Printf("[SCREEN_CONTROL] DisplayAIResult failed: %s", resp.ErrorCode)
	} else {
		log.Printf("[SCREEN_CONTROL] AI result displayed successfully")
	}

	return nil
}

// Close 연결 종료
func (a *ScreenControlAdapter) Close() error {
	if a.conn != nil {
		return a.conn.Close()
	}
	return nil
}

// mapToVisualEffectType ActionType → VisualEffectType 변환
func mapToVisualEffectType(action domain.ActionType) proto.VisualEffectType {
	switch action {
	case domain.ActionBlockURL:
		return proto.VisualEffectType_RED_FLASH
	case domain.ActionCloseApp:
		return proto.VisualEffectType_SCREEN_GLITCH
	case domain.ActionSleepScreen:
		return proto.VisualEffectType_BLACK_SCREEN
	case domain.ActionMinimizeAll:
		return proto.VisualEffectType_BLUR_OVERLAY
	default:
		return proto.VisualEffectType_SCREEN_SHAKE
	}
}
