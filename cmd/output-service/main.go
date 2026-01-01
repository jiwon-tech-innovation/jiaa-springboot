package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	// Adapters - In
	grpcIn "jiaa-server-core/internal/output/adapter/in/grpc"

	// Adapters - Out
	grpcOut "jiaa-server-core/internal/output/adapter/out/grpc"

	// Services
	"jiaa-server-core/internal/output/service"
)

// Config 서버 설정
type Config struct {
	GRPCPort            string // gRPC 서버 포트
	PhysicalControlAddr string // Dev 1 gRPC 주소
	ScreenControlAddr   string // Dev 3 gRPC 주소
}

func main() {
	// 1. Configuration
	config := loadConfig()
	log.Printf("[MAIN] Starting Output Service (Sabotage Executor)")
	log.Printf("[MAIN] Config: gRPC Port=%s, Dev1=%s, Dev3=%s",
		config.GRPCPort, config.PhysicalControlAddr, config.ScreenControlAddr)

	// 2. Initialize Adapters (Driven - Out)
	// Physical Executor (→ Dev 1)
	physicalExecutor := grpcOut.NewPhysicalExecutorAdapterLazy(config.PhysicalControlAddr)
	log.Printf("[MAIN] Physical executor adapter initialized (lazy)")

	// Screen Executor (→ Dev 3)
	screenExecutor := grpcOut.NewScreenExecutorAdapterLazy(config.ScreenControlAddr)
	log.Printf("[MAIN] Screen executor adapter initialized (lazy)")

	// 3. Initialize Services
	sabotageExecutorService := service.NewSabotageExecutorService(physicalExecutor, screenExecutor)
	log.Printf("[MAIN] SabotageExecutorService initialized")

	// 4. Initialize gRPC Server (Driving - In)
	grpcServer := grpcIn.NewSabotageServer(config.GRPCPort, sabotageExecutorService)

	// 5. Start gRPC Server
	if err := grpcServer.Start(); err != nil {
		log.Fatalf("[MAIN] Failed to start gRPC server: %v", err)
	}
	log.Printf("[MAIN] gRPC server started on port %s", config.GRPCPort)

	// 6. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Printf("[MAIN] Shutting down...")

	// Cleanup
	grpcServer.Stop()
	physicalExecutor.Close()
	screenExecutor.Close()

	log.Printf("[MAIN] Shutdown complete")
}

// loadConfig 환경 변수에서 설정 로드
func loadConfig() Config {
	return Config{
		GRPCPort:            getEnv("GRPC_PORT", "50053"),
		PhysicalControlAddr: getEnv("PHYSICAL_CONTROL_ADDR", "localhost:50051"),
		ScreenControlAddr:   getEnv("SCREEN_CONTROL_ADDR", "localhost:50052"),
	}
}

// getEnv 환경 변수 조회 (기본값 지원)
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
