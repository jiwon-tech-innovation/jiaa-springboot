package grpc

import (
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	portin "jiaa-server-core/internal/input/port/in"
	"jiaa-server-core/internal/input/service"
	"jiaa-server-core/pkg/auth"
	"jiaa-server-core/pkg/proto"
)

// InputGrpcServer manages the gRPC server for incoming client connections (e.g. Vision Service)
type InputGrpcServer struct {
	server        *grpc.Server
	port          string
	coreService   *CoreServiceServer
	reflexService portin.ReflexUseCase
}

// NewInputGrpcServer creates a new gRPC server wrapper
func NewInputGrpcServer(port string, reflexService portin.ReflexUseCase, scoreService *service.ScoreService) *InputGrpcServer {
	return &InputGrpcServer{
		port:          port,
		coreService:   NewCoreServiceServer(reflexService, scoreService),
		reflexService: reflexService,
	}
}

// Start starts the gRPC server
func (s *InputGrpcServer) Start() error {
	lis, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		return err
	}

	// Create JWT Interceptor
	jwtInterceptor := auth.NewJWTInterceptorFromEnv()

	s.server = grpc.NewServer(
		grpc.UnaryInterceptor(jwtInterceptor.UnaryInterceptor()),
		grpc.StreamInterceptor(jwtInterceptor.StreamInterceptor()),
	)

	// Register Services
	proto.RegisterCoreServiceServer(s.server, s.coreService)

	// Register Health Server
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(s.server, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("jiaa.score.ScoreService", grpc_health_v1.HealthCheckResponse_SERVING)

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
