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
	scoreService  *service.ScoreService
}

// NewCoreServiceServer creates a new instance of CoreServiceServer
func NewCoreServiceServer(reflexService portin.ReflexUseCase, scoreService *service.ScoreService) *CoreServiceServer {
	return &CoreServiceServer{
		reflexService: reflexService,
		scoreService:  scoreService,
	}
}

// SyncClient handles bidirectional streaming between Client (Dev 2/Vision) and Server
func (s *CoreServiceServer) SyncClient(stream proto.CoreService_SyncClientServer) error {
	log.Println("[CoreService] SyncClient connected")

	// State for Score Calculation
	var consecutiveSleepSec float64 = 0.0
	var currentScore int = 100
	var lastAlertTime time.Time

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

		// Debug Log
		log.Printf("[DEBUG] Heartbeat recv: Keys=%d, Mouse=%d, Click=%d, Ent=%.2f, Drag=%v, App='%s'",
			heartbeat.KeystrokeCount, heartbeat.MouseDistance, heartbeat.ClickCount,
			heartbeat.KeyboardEntropy, heartbeat.IsDragging, heartbeat.ActiveWindowTitle)

		// 2. Aggregate Data
		if heartbeat.IsEyesClosed {
			consecutiveSleepSec += 1.0 // Assuming 1Hz heartbeat
		} else {
			consecutiveSleepSec = 0
		}

		osActivity := int(heartbeat.KeystrokeCount) + int(heartbeat.ClickCount) + int(heartbeat.MouseDistance)
		visionScore := int(heartbeat.ConcentrationScore * 100) // 0.0~1.0 -> 0~100

		// Route Key/Mouse Usage to ReflexService -> Kafka
		if osActivity > 0 {
			activity := domain.NewClientActivity(heartbeat.ClientId, domain.ActivityInputUsage)
			activity.AddMetadata("keystroke_count", fmt.Sprintf("%d", heartbeat.KeystrokeCount))
			activity.AddMetadata("mouse_distance", fmt.Sprintf("%d", heartbeat.MouseDistance))
			activity.AddMetadata("click_count", fmt.Sprintf("%d", heartbeat.ClickCount))
			activity.AddMetadata("entropy", fmt.Sprintf("%.2f", heartbeat.KeyboardEntropy))
			activity.AddMetadata("window_title", heartbeat.ActiveWindowTitle)
			activity.AddMetadata("is_dragging", fmt.Sprintf("%v", heartbeat.IsDragging))
			activity.AddMetadata("avg_dwell_time", fmt.Sprintf("%.2f", heartbeat.AvgDwellTime))

			if _, err := s.reflexService.ProcessActivity(*activity); err != nil {
				log.Printf("[CoreService] Failed to route input activity: %v", err)
			}
		}

		// 3. Calculate Score (JIAA Logic)
		input := service.CalculateInput{
			EyesClosedDurationSec: consecutiveSleepSec,
			HeadPitch:             0, // Not available in heartbeat yet
			URLCategory:           "NEUTRAL",
			OSActivityCount:       osActivity,
			VisionScore:           visionScore,
			CurrentScore:          currentScore,
		}

		result := s.scoreService.CalculateScore(input)
		currentScore = result.FinalScore

		if result.State != "FOCUSING" && result.State != "THINKING" && result.State != "NEUTRAL" {
			log.Printf("[CoreService] State: %s, Score: %d", result.State, currentScore)
		}

		// 4. Command Generation (Feedback) with Cooldown
		now := time.Now()
		if now.Sub(lastAlertTime) > 5*time.Second {
			alertTriggered := false

			if result.State == "SLEEPING" {
				// Trigger TTS
				stream.Send(&proto.ServerCommand{
					Type:    proto.ServerCommand_PLAY_SOUND,
					Payload: "일어나세요! 코딩해야죠!",
				})
				stream.Send(&proto.ServerCommand{
					Type:    proto.ServerCommand_SHAKE_MOUSE,
					Payload: "Wake Up",
				})
				alertTriggered = true
			} else if result.State == "DISTRACTED" {
				stream.Send(&proto.ServerCommand{
					Type:    proto.ServerCommand_PLAY_SOUND,
					Payload: "딴짓 금지! 집중하세요.",
				})
				alertTriggered = true
			} else if result.State == "IDLING" && currentScore < 50 {
				stream.Send(&proto.ServerCommand{
					Type:    proto.ServerCommand_PLAY_SOUND,
					Payload: "멍때리지 마세요!",
				})
				alertTriggered = true
			}

			if alertTriggered {
				lastAlertTime = now
			}
		}
	}
	return nil
}

// ReportAnalysisResult handles reports from AI Service
func (s *CoreServiceServer) ReportAnalysisResult(ctx context.Context, req *proto.AnalysisReport) (*proto.Ack, error) {
	log.Printf("[CoreService] Received Analysis Report: %s - %s", req.Type, req.Content)
	return &proto.Ack{Success: true}, nil
}
