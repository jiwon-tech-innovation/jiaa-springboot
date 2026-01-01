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

// ScreenExecutorAdapter Dev 3(화면 제어) 실행 어댑터 (Driven Adapter)
type ScreenExecutorAdapter struct {
	conn    *grpc.ClientConn
	client  proto.ScreenControlServiceClient
	address string
}

// NewScreenExecutorAdapter ScreenExecutorAdapter 생성자
func NewScreenExecutorAdapter(address string) (*ScreenExecutorAdapter, error) {
	conn, err := grpc.NewClient(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return &ScreenExecutorAdapter{
		conn:    conn,
		client:  proto.NewScreenControlServiceClient(conn),
		address: address,
	}, nil
}

// NewScreenExecutorAdapterLazy 지연 연결 생성자
func NewScreenExecutorAdapterLazy(address string) *ScreenExecutorAdapter {
	return &ScreenExecutorAdapter{
		address: address,
	}
}

// Connect gRPC 연결 수립
func (a *ScreenExecutorAdapter) Connect() error {
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
	log.Printf("[SCREEN_EXECUTOR] Connected to Dev 3: %s", a.address)
	return nil
}

// Execute 화면 제어 명령 실행
func (a *ScreenExecutorAdapter) Execute(cmd domain.SabotageCommand) (*domain.ComponentResult, error) {
	startTime := time.Now()

	if a.conn == nil {
		if err := a.Connect(); err != nil {
			log.Printf("[SCREEN_EXECUTOR] Failed to connect: %v", err)
			return &domain.ComponentResult{
				Success:   false,
				ErrorCode: "CONNECTION_ERROR",
				Message:   err.Error(),
			}, err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// TTS는 별도 처리
	if cmd.SabotageType == domain.SabotageTTS {
		return a.executeTTS(ctx, cmd, startTime)
	}

	effectType := mapSabotageToVisualEffect(cmd.SabotageType)

	req := &proto.VisualCommandRequest{
		ClientId:   cmd.ClientID,
		EffectType: effectType,
		Intensity:  int32(cmd.Intensity),
		DurationMs: int32(cmd.DurationMs),
		Message:    cmd.Message,
	}

	log.Printf("[SCREEN_EXECUTOR] Executing: Client: %s, Effect: %v", cmd.ClientID, effectType)

	resp, err := a.client.ExecuteVisualCommand(ctx, req)
	latency := time.Since(startTime).Milliseconds()

	if err != nil {
		log.Printf("[SCREEN_EXECUTOR] gRPC call failed: %v", err)
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
		Latency:   latency,
	}, nil
}

// executeTTS TTS 실행
func (a *ScreenExecutorAdapter) executeTTS(ctx context.Context, cmd domain.SabotageCommand, startTime time.Time) (*domain.ComponentResult, error) {
	req := &proto.TTSRequest{
		ClientId: cmd.ClientID,
		Text:     cmd.Message,
		Speed:    1.0,
	}

	log.Printf("[SCREEN_EXECUTOR] Executing TTS: Client: %s, Text: %s", cmd.ClientID, cmd.Message)

	resp, err := a.client.PlayTTS(ctx, req)
	latency := time.Since(startTime).Milliseconds()

	if err != nil {
		return &domain.ComponentResult{
			Success:   false,
			ErrorCode: "TTS_ERROR",
			Message:   err.Error(),
			Latency:   latency,
		}, err
	}

	return &domain.ComponentResult{
		Success:   resp.Success,
		ErrorCode: resp.ErrorCode,
		Latency:   latency,
	}, nil
}

// Close 연결 종료
func (a *ScreenExecutorAdapter) Close() error {
	if a.conn != nil {
		return a.conn.Close()
	}
	return nil
}

// mapSabotageToVisualEffect SabotageType → VisualEffectType 변환
func mapSabotageToVisualEffect(sabotageType domain.SabotageType) proto.VisualEffectType {
	switch sabotageType {
	case domain.SabotageScreenGlitch:
		return proto.VisualEffectType_SCREEN_GLITCH
	case domain.SabotageRedFlash:
		return proto.VisualEffectType_RED_FLASH
	case domain.SabotageBlackScreen:
		return proto.VisualEffectType_BLACK_SCREEN
	case domain.SabotageBlockURL:
		return proto.VisualEffectType_RED_FLASH
	default:
		return proto.VisualEffectType_SCREEN_SHAKE
	}
}
