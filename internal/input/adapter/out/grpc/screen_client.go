package grpc

import (
	"log"

	"jiaa-server-core/internal/input/adapter/in/grpc" // Import for StreamManager
	"jiaa-server-core/internal/input/domain"
	"jiaa-server-core/pkg/proto"
)

// ScreenControlAdapter Dev 3(화면 제어) gRPC 클라이언트 (Driven Adapter)
type ScreenControlAdapter struct {
	// No connection needed, uses StreamManager
}

// NewScreenControlAdapter ScreenControlAdapter 생성자
func NewScreenControlAdapter() *ScreenControlAdapter {
	return &ScreenControlAdapter{}
}

// NewScreenControlAdapterLazy (Deprecated signature kept for compatibility if needed, but ignores address)
func NewScreenControlAdapterLazy(address string) *ScreenControlAdapter {
	return &ScreenControlAdapter{}
}

// Connect (No-op)
func (a *ScreenControlAdapter) Connect() error {
	return nil
}

// SendToScreenController 화면 제어 명령 전송 (Router -> StreamManager -> Client)
func (a *ScreenControlAdapter) SendToScreenController(cmd domain.SabotageAction) error {
	sm := grpc.GetStreamManager()
	
	// Map ActionType to ServerCommandType
	var commandType proto.ServerCommand_CommandType
	var payload string = cmd.Message

	switch cmd.ActionType {
	case domain.ActionBlockURL:
		commandType = proto.ServerCommand_SHOW_MESSAGE
		payload = "BLOCK_URL:" + cmd.TargetURL // Protocol agreement needed
	case domain.ActionCloseApp:
		commandType = proto.ServerCommand_BLOCK_SCREEN
	case domain.ActionSleepScreen:
		commandType = proto.ServerCommand_BLOCK_SCREEN
	case domain.ActionMinimizeAll:
		commandType = proto.ServerCommand_SHAKE_MOUSE // Using closest available
	default:
		commandType = proto.ServerCommand_SHAKE_MOUSE
	}

	serverCmd := &proto.ServerCommand{
		Type:    commandType,
		Payload: payload,
	}

	log.Printf("[SCREEN_CONTROL] Routing command to client via StreamManager: %s", cmd.ClientID)
	
	if err := sm.SendCommand(cmd.ClientID, serverCmd); err != nil {
		log.Printf("[SCREEN_CONTROL] Failed to route via StreamManager: %v", err)
		return err
	}
	
	return nil
}

// SendAIResult AI 결과(Markdown) 전송
func (a *ScreenControlAdapter) SendAIResult(clientID string, markdown string) error {
	sm := grpc.GetStreamManager()

	// Using SHOW_MESSAGE type to send markdown content
	// Protocol agreement: Payload starts with "MARKDOWN:"??? 
	// Or just raw payload if client handles it.
	// For now, let's assume SHOW_MESSAGE displays it.
	
	serverCmd := &proto.ServerCommand{
		Type:    proto.ServerCommand_SHOW_MESSAGE,
		Payload: markdown,
	}

	log.Printf("[SCREEN_CONTROL] Routing AI Result to client via StreamManager: %s", clientID)
	
	if err := sm.SendCommand(clientID, serverCmd); err != nil {
		log.Printf("[SCREEN_CONTROL] Failed to route AI Result: %v", err)
		return err
	}

	return nil
}

// Close (No-op)
func (a *ScreenControlAdapter) Close() error {
	return nil
}

func mapToVisualEffectType(action domain.ActionType) proto.VisualEffectType {
	// Kept for reference or future use if proto creates VisualEffectType
	return proto.VisualEffectType_SCREEN_SHAKE
}

