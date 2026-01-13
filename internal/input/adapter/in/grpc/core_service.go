package grpc

import (
	"context"
	"io"
	"log"

	"time"

	"fmt"
	"jiaa-server-core/internal/input/domain"
	portin "jiaa-server-core/internal/input/port/in"
	"jiaa-server-core/internal/input/service"
	proto "jiaa-server-core/pkg/proto"
)

	// CoreServiceServer implements the CoreService gRPC server
type CoreServiceServer struct {
	proto.UnimplementedCoreServiceServer
	reflexService portin.ReflexUseCase
	scoreService *service.ScoreService
}

// NewCoreServiceServer creates a new instance of CoreServiceServer
func NewCoreServiceServer(reflexService portin.ReflexUseCase, scoreService *service.ScoreService) *CoreServiceServer {
	return &CoreServiceServer{
		reflexService: reflexService,
		scoreService: scoreService,
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
	
	if osActivity > 0 || heartbeat.IsEyesClosed {
		activity := domain.NewClientActivity(heartbeat.ClientId, domain.ActivityInputUsage)
		activity.AddMetadata("keystroke_count", fmt.Sprintf("%d", heartbeat.KeystrokeCount))
		activity.AddMetadata("mouse_distance", fmt.Sprintf("%d", heartbeat.MouseDistance))
		activity.AddMetadata("click_count", fmt.Sprintf("%d", heartbeat.ClickCount))
		activity.AddMetadata("entropy", fmt.Sprintf("%.2f", heartbeat.KeyboardEntropy))
		activity.AddMetadata("window_title", heartbeat.ActiveWindowTitle)
		activity.AddMetadata("is_dragging", fmt.Sprintf("%v", heartbeat.IsDragging))
		activity.AddMetadata("avg_dwell_time", fmt.Sprintf("%.2f", heartbeat.AvgDwellTime))
		
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

// SendAppList handles app list updates from client
func (s *CoreServiceServer) SendAppList(ctx context.Context, req *proto.AppListRequest) (*proto.AppListResponse, error) {
	log.Printf("[CoreService] Received App List (len=%d chars)", len(req.AppsJson))
	return &proto.AppListResponse{Success: true, Message: "Apps received"}, nil
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
				Transcript: "(Go Server) Audio received successfully",
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
