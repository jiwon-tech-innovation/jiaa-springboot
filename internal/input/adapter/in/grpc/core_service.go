package grpc

import (
	"context"
	"io"
	"log"

	"fmt"
	"jiaa-server-core/internal/input/domain"
	portin "jiaa-server-core/internal/input/port/in"
	portout "jiaa-server-core/internal/input/port/out"
	"jiaa-server-core/internal/input/service"
	proto "jiaa-server-core/pkg/proto"
)

// CoreServiceServer implements the CoreService gRPC server
type CoreServiceServer struct {
	proto.UnimplementedCoreServiceServer
	reflexService       portin.ReflexUseCase
	scoreService        *service.ScoreService
	intelligenceService portout.IntelligencePort
}

// NewCoreServiceServer creates a new instance of CoreServiceServer
func NewCoreServiceServer(reflexService portin.ReflexUseCase, scoreService *service.ScoreService, intelligenceService portout.IntelligencePort) *CoreServiceServer {
	return &CoreServiceServer{
		reflexService:       reflexService,
		scoreService:        scoreService,
		intelligenceService: intelligenceService,
	}
}

// SyncClient handles bidirectional streaming between Client (Dev 2/Vision) and Server
func (s *CoreServiceServer) SyncClient(stream proto.CoreService_SyncClientServer) error {
	log.Println("[CoreService] SyncClient connected")

	// Wait for first heartbeat to get ClientID
	firstMsg, err := stream.Recv()
	if err != nil {
		log.Printf("[CoreService] Failed to receive first heartbeat: %v", err)
		return err
	}

	clientID := firstMsg.ClientId
	if clientID == "" {
		clientID = "unknown"
	}

	sm := GetStreamManager()
	sm.Register(clientID, stream)
	defer sm.Unregister(clientID)

	// Process first message
	s.processHeartbeat(firstMsg)

	for {
		// 1. Receive Heartbeat from Client
		heartbeat, err := stream.Recv()
		if err == io.EOF {
			log.Println("[CoreService] Client disconnected (EOF)")
			return nil
		}
		if err != nil {
			log.Printf("[CoreService] Error receiving heartbeat: %v", err)
			return err
		}

		s.processHeartbeat(heartbeat)
	}
	return nil
}

func (s *CoreServiceServer) processHeartbeat(heartbeat *proto.ClientHeartbeat) {
	// Debug Log
	// log.Printf("[DEBUG] Heartbeat recv: Keys=%d...", heartbeat.KeystrokeCount)

	// 2. Aggregate Data and Route to ReflexService -> Kafka
	osActivity := int(heartbeat.KeystrokeCount) + int(heartbeat.ClickCount) + int(heartbeat.MouseDistance)

	// 통계 기록 조건: OS 활동이 있거나, 눈이 감겼거나, 비전 점수가 있거나, 시간 추적 데이터가 있으면 기록
	// (생각 모드, 멍 때리기, 일반 상태 등 모든 상황을 포함)
	// 시간 추적 데이터(focus_time, sleep_time, away_time, distraction_time)가 있으면 반드시 기록해야 함
	hasTimeTracking := heartbeat.FocusTime > 0 || heartbeat.SleepTime > 0 || heartbeat.AwayTime > 0 || heartbeat.DistractionTime > 0
	if osActivity > 0 || heartbeat.IsEyesClosed || heartbeat.ConcentrationScore > 0 || hasTimeTracking {
		activity := domain.NewClientActivity(heartbeat.ClientId, domain.ActivityInputUsage)
		activity.AddMetadata("keystroke_count", fmt.Sprintf("%d", heartbeat.KeystrokeCount))
		activity.AddMetadata("mouse_distance", fmt.Sprintf("%d", heartbeat.MouseDistance))
		activity.AddMetadata("click_count", fmt.Sprintf("%d", heartbeat.ClickCount))
		activity.AddMetadata("entropy", fmt.Sprintf("%.2f", heartbeat.KeyboardEntropy))
		activity.AddMetadata("window_title", heartbeat.ActiveWindowTitle)
		activity.AddMetadata("is_dragging", fmt.Sprintf("%v", heartbeat.IsDragging))
		activity.AddMetadata("avg_dwell_time", fmt.Sprintf("%.2f", heartbeat.AvgDwellTime))
		// 눈 감음 상태 및 비전 점수 메타데이터 추가 (통계 기록용)
		activity.AddMetadata("is_eyes_closed", fmt.Sprintf("%v", heartbeat.IsEyesClosed))
		activity.AddMetadata("concentration_score", fmt.Sprintf("%.2f", heartbeat.ConcentrationScore))
		activity.AddMetadata("is_os_idle", fmt.Sprintf("%v", heartbeat.IsOsIdle))
		
		// 시간 추적 데이터 메타데이터 추가 (통계 기록용)
		if heartbeat.FocusTime > 0 {
			activity.AddMetadata("focus_time", fmt.Sprintf("%.2f", heartbeat.FocusTime))
		}
		if heartbeat.SleepTime > 0 {
			activity.AddMetadata("sleep_time", fmt.Sprintf("%.2f", heartbeat.SleepTime))
		}
		if heartbeat.AwayTime > 0 {
			activity.AddMetadata("away_time", fmt.Sprintf("%.2f", heartbeat.AwayTime))
		}
		if heartbeat.DistractionTime > 0 {
			activity.AddMetadata("distraction_time", fmt.Sprintf("%.2f", heartbeat.DistractionTime))
		}

		// [Reflex Check] - Local Fast Path (e.g. Blacklist)
		if _, err := s.reflexService.ProcessActivity(*activity); err != nil {
			log.Printf("[CoreService] Failed to route activity: %v", err)
		}
	}
}

// ReportAnalysisResult handles reports from AI Service
func (s *CoreServiceServer) ReportAnalysisResult(ctx context.Context, req *proto.AnalysisReport) (*proto.Ack, error) {
	log.Printf("[CoreService] Received Analysis Report: %s - %s", req.Type, req.Content)
	return &proto.Ack{Success: true}, nil
}

// SendAppList handles app list updates from client (Forwards to AI)
func (s *CoreServiceServer) SendAppList(ctx context.Context, req *proto.AppListRequest) (*proto.AppListResponse, error) {
	// log.Printf("[CoreService] Forwarding App List to AI (len=%d chars)", len(req.AppsJson))

	msg, cmd, target, err := s.intelligenceService.SendAppList(req.AppsJson)
	if err != nil {
		log.Printf("[CoreService] Failed to forward to AI: %v", err)
		// 에러 발생해도 클라가 크래시나지 않게 성공 처리하되 메시지 전달
		return &proto.AppListResponse{Success: false, Message: fmt.Sprintf("AI Server Error: %v", err)}, nil
	}

	return &proto.AppListResponse{
		Success:   true,
		Message:   msg,
		Command:   cmd,
		TargetApp: target,
	}, nil
}

// TranscribeAudio handles audio stream from client
func (s *CoreServiceServer) TranscribeAudio(stream proto.CoreService_TranscribeAudioServer) error {
	log.Println("[CoreService] Audio stream started")
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			// Finished receiving audio
			log.Println("[CoreService] Audio stream ended")
			return stream.SendAndClose(&proto.AudioResponse{
				Transcript:  "(Go Server) Audio received successfully",
				IsEmergency: false,
			})
		}
		if err != nil {
			log.Printf("[CoreService] Audio stream error: %v", err)
			return err
		}
		// Process audio chunk (req.AudioData)
		if req.IsFinal {
			log.Println("[CoreService] Final audio chunk received")
		}
	}
}
