package grpc

import (
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	portin "jiaa-server-core/internal/input/port/in"
	"jiaa-server-core/pkg/proto"
)

// InputGrpcServer manages the gRPC server for incoming client connections (e.g. Vision Service)
type InputGrpcServer struct {
	server      *grpc.Server
	port        string
	coreService *CoreServiceServer
	reflexService portin.ReflexUseCase
}

// NewInputGrpcServer creates a new gRPC server wrapper
func NewInputGrpcServer(port string, reflexService portin.ReflexUseCase) *InputGrpcServer {
	return &InputGrpcServer{
		port:        port,
		coreService: NewCoreServiceServer(reflexService),
		reflexService: reflexService,
	}
}

// Start starts the gRPC server
func (s *InputGrpcServer) Start() error {
	lis, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		return err
	}

	s.server = grpc.NewServer()
	
	// Register Services
	proto.RegisterCoreServiceServer(s.server, s.coreService)

	// Enable Reflection
	reflection.Register(s.server)

	log.Printf("[INPUT_GRPC] Starting gRPC server on port %s", s.port)

	go func() {
		if err := s.server.Serve(lis); err != nil {
			log.Printf("[INPUT_GRPC] Server error: %v", err)
		}
	}()

	return nil
}

// Stop stops the gRPC server
func (s *InputGrpcServer) Stop() {
	if s.server != nil {
		s.server.GracefulStop()
		log.Printf("[INPUT_GRPC] Server stopped")
	}
}
