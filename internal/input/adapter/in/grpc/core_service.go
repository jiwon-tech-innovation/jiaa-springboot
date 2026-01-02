package grpc

import (
	"context"
	"io"
	"log"

	"fmt"
	"jiaa-server-core/internal/input/domain"
	portin "jiaa-server-core/internal/input/port/in"
	proto "jiaa-server-core/pkg/proto"
)

// CoreServiceServer implements the CoreService gRPC server
type CoreServiceServer struct {
	proto.UnimplementedCoreServiceServer
	reflexService portin.ReflexUseCase
}

// NewCoreServiceServer creates a new instance of CoreServiceServer
func NewCoreServiceServer(reflexService portin.ReflexUseCase) *CoreServiceServer {
	return &CoreServiceServer{
		reflexService: reflexService,
	}
}

// SyncClient handles bidirectional streaming between Client (Dev 2/Vision) and Server
func (s *CoreServiceServer) SyncClient(stream proto.CoreService_SyncClientServer) error {
	log.Println("[CoreService] SyncClient connected")

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
		log.Printf("[DEBUG] Heartbeat recv: Keys=%d, Mouse=%d", heartbeat.KeystrokeCount, heartbeat.MouseDistance)

		// Log minimal info to avoid spam
		if heartbeat.IsEyesClosed {
			log.Printf("[CoreService] [%s] Eyes Closed! Score: %.2f", heartbeat.ClientId, heartbeat.ConcentrationScore)
		}

		// Route Key/Mouse Usage to ReflexService -> Kafka -> DataService
		if heartbeat.KeystrokeCount > 0 || heartbeat.MouseDistance > 0 || heartbeat.ClickCount > 0 {
			activity := domain.NewClientActivity(heartbeat.ClientId, domain.ActivityInputUsage)
			activity.AddMetadata("keystroke_count", fmt.Sprintf("%d", heartbeat.KeystrokeCount))
			activity.AddMetadata("mouse_distance", fmt.Sprintf("%d", heartbeat.MouseDistance))
			activity.AddMetadata("click_count", fmt.Sprintf("%d", heartbeat.ClickCount))
			
			if _, err := s.reflexService.ProcessActivity(*activity); err != nil {
				log.Printf("[CoreService] Failed to route input activity: %v", err)
			} else {
				// log.Printf("[CoreService] Routed input activity for %s", heartbeat.ClientId)
			}
		}

		// 2. Logic: Reflex / Command Generation
		// Simple Mock Logic for now: If Eyes Closed -> SHAKE_MOUSE
		if heartbeat.IsEyesClosed {
			cmd := &proto.ServerCommand{
				Type:    proto.ServerCommand_SHAKE_MOUSE,
				Payload: "Wake up! Eyes detected closed.",
			}
			if err := stream.Send(cmd); err != nil {
				log.Printf("[CoreService] Failed to send command: %v", err)
				return err
			}
			log.Println("[CoreService] Sent SHAKE_MOUSE command")
		} else if heartbeat.ConcentrationScore > 0 && heartbeat.ConcentrationScore < 0.3 {
			// Low concentration -> Message
			cmd := &proto.ServerCommand{
				Type:    proto.ServerCommand_SHOW_MESSAGE,
				Payload: "Focus logic triggered (Low Score)",
			}
			if err := stream.Send(cmd); err != nil {
				log.Printf("[CoreService] Failed to send command: %v", err)
				return err
			}
		}
	}
}

// ReportAnalysisResult handles reports from AI Service
func (s *CoreServiceServer) ReportAnalysisResult(ctx context.Context, req *proto.AnalysisReport) (*proto.Ack, error) {
	log.Printf("[CoreService] Received Analysis Report: %s - %s", req.Type, req.Content)
	return &proto.Ack{Success: true}, nil
}
