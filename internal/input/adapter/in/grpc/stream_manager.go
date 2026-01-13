package grpc

import (
	"sync"
	"log"
	
	proto "jiaa-server-core/pkg/proto"
)

// StreamManager manages active gRPC streams for clients
type StreamManager struct {
	streams map[string]proto.CoreService_SyncClientServer
	mu      sync.RWMutex
}

var instance *StreamManager
var once sync.Once

// GetStreamManager returns the singleton instance
func GetStreamManager() *StreamManager {
	once.Do(func() {
		instance = &StreamManager{
			streams: make(map[string]proto.CoreService_SyncClientServer),
		}
	})
	return instance
}

// Register registers a stream for a client
func (sm *StreamManager) Register(clientID string, stream proto.CoreService_SyncClientServer) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.streams[clientID] = stream
	log.Printf("[StreamManager] Registered stream for client: %s", clientID)
}

// Unregister removes a stream for a client
func (sm *StreamManager) Unregister(clientID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if _, exists := sm.streams[clientID]; exists {
		delete(sm.streams, clientID)
		log.Printf("[StreamManager] Unregistered stream for client: %s", clientID)
	}
}

// Get returns the stream for a client
func (sm *StreamManager) Get(clientID string) (proto.CoreService_SyncClientServer, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	stream, exists := sm.streams[clientID]
	return stream, exists
}

// SendCommand sends a command to a specific client
func (sm *StreamManager) SendCommand(clientID string, cmd *proto.ServerCommand) error {
	stream, exists := sm.Get(clientID)
	if !exists {
		log.Printf("[StreamManager] Client not found: %s", clientID)
		return nil // Return nil to avoid erroring out caller, just log warning
	}
	return stream.Send(cmd)
}
