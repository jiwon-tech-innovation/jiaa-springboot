package grpc

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"

	"jiaa-server-core/internal/output/domain"
	portin "jiaa-server-core/internal/output/port/in"
	"jiaa-server-core/pkg/proto"
)

// SabotageServer gRPC 서버 (Driving Adapter)
// Dev 4(Core Decision Service)로부터 사보타주 명령 수신
type SabotageServer struct {
	proto.UnimplementedSabotageCommandServiceServer
	sabotageExecutor portin.SabotageExecutorUseCase
	server           *grpc.Server
	port             string
}

// NewSabotageServer SabotageServer 생성자
func NewSabotageServer(port string, executor portin.SabotageExecutorUseCase) *SabotageServer {
	return &SabotageServer{
		sabotageExecutor: executor,
		port:             port,
	}
}

// Start gRPC 서버 시작
func (s *SabotageServer) Start() error {
	lis, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		return err
	}

	s.server = grpc.NewServer()
	proto.RegisterSabotageCommandServiceServer(s.server, s)

	log.Printf("[SABOTAGE_SERVER] Starting gRPC server on port %s", s.port)

	go func() {
		if err := s.server.Serve(lis); err != nil {
			log.Printf("[SABOTAGE_SERVER] Server error: %v", err)
		}
	}()

	return nil
}

// Stop gRPC 서버 종료
func (s *SabotageServer) Stop() {
	if s.server != nil {
		s.server.GracefulStop()
		log.Printf("[SABOTAGE_SERVER] Server stopped")
	}
}

// ExecuteSabotage gRPC 메서드 구현
func (s *SabotageServer) ExecuteSabotage(ctx context.Context, req *proto.SabotageRequest) (*proto.SabotageResponse, error) {
	log.Printf("[SABOTAGE_SERVER] Received sabotage request: Client: %s, Action: %s",
		req.ClientId, req.ActionType)

	// Convert to domain entity
	cmd := domain.SabotageCommand{
		ClientID:     req.ClientId,
		SabotageType: mapToSabotageType(req.ActionType),
		Intensity:    int(req.Intensity),
		Message:      req.Message,
	}

	// Execute
	result, err := s.sabotageExecutor.ExecuteSabotage(cmd)
	if err != nil {
		log.Printf("[SABOTAGE_SERVER] Execution error: %v", err)
		return &proto.SabotageResponse{
			Success:   false,
			ErrorCode: "EXECUTION_ERROR",
		}, nil
	}

	success := result.Status == domain.StatusSuccess || result.Status == domain.StatusPartial

	log.Printf("[SABOTAGE_SERVER] Execution result: %s", result.Status)

	return &proto.SabotageResponse{
		Success:   success,
		ErrorCode: "",
	}, nil
}

// mapToSabotageType ActionType 문자열을 SabotageType으로 변환
func mapToSabotageType(actionType string) domain.SabotageType {
	switch actionType {
	case "BLOCK_URL":
		return domain.SabotageBlockURL
	case "CLOSE_APP":
		return domain.SabotageCloseApp
	case "MINIMIZE_ALL":
		return domain.SabotageMinimizeAll
	case "SCREEN_GLITCH":
		return domain.SabotageScreenGlitch
	case "RED_FLASH":
		return domain.SabotageRedFlash
	case "BLACK_SCREEN":
		return domain.SabotageBlackScreen
	case "TTS":
		return domain.SabotageTTS
	default:
		return domain.SabotageScreenGlitch
	}
}
