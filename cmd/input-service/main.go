package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	// Adapters - In
	grpcIn "jiaa-server-core/internal/input/adapter/in/grpc"
	httpAdapter "jiaa-server-core/internal/input/adapter/in/http"
	kafkaIn "jiaa-server-core/internal/input/adapter/in/kafka"

	// Adapters - Out
	grpcOut "jiaa-server-core/internal/input/adapter/out/grpc"
	kafkaOut "jiaa-server-core/internal/input/adapter/out/kafka"
	"jiaa-server-core/internal/input/adapter/out/memory"

	// Services
	"jiaa-server-core/internal/input/service"
)

// Config 서버 설정
type Config struct {
	HTTPPort            string
	KafkaBrokers        string
	ActivityTopic       string // client-activity topic (→ Dev 6)
	StateTopic          string // command-state topic (← Dev 6)
	PhysicalControlAddr string // Dev 1 gRPC 주소
	ScreenControlAddr   string // Dev 3 gRPC 주소
	SabotageCommandAddr string // SabotageCommand gRPC 주소
	IntelligenceAddr    string // Dev 5 (AI) gRPC 주소
}

func main() {
	// 1. Configuration
	config := loadConfig()
	log.Printf("[MAIN] Starting Core Decision Service (Input Service)")
	log.Printf("[MAIN] Config: HTTP Port=%s, Kafka=%s", config.HTTPPort, config.KafkaBrokers)

	// 2. Initialize Adapters (Driven - Out)
	// Blacklist Adapter (인메모리 - 기본 블랙리스트 포함)
	blacklistAdapter := memory.NewBlacklistAdapterWithDefaults()
	log.Printf("[MAIN] Blacklist adapter initialized with defaults")

	// [LOCAL HOTFIX] Fetch Blacklist from Data Service (localhost:8083)
	// We do this in a goroutine to not block startup, but with a small delay to ensure adapter is ready
	go func() {
		time.Sleep(2 * time.Second) // Wait for server to start
		url := "http://localhost:8083/api/v1/blacklist"
		log.Printf("[MAIN] Fetching external blacklist from %s...", url)

		resp, err := http.Get(url)
		if err != nil {
			log.Printf("[MAIN] Failed to fetch blacklist: %v", err)
			return
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("[MAIN] Failed to read blacklist body: %v", err)
			return
		}

		var result struct {
			Success bool `json:"success"`
			Data    []struct {
				AppName string `json:"appName"`
			} `json:"data"`
		}

		if err := json.Unmarshal(body, &result); err != nil {
			log.Printf("[MAIN] Failed to parse blacklist JSON: %v", err)
			return
		}

		count := 0
		for _, item := range result.Data {
			blacklistAdapter.AddAppToBlacklist(item.AppName)
			count++
		}
		log.Printf("[MAIN] Successfully synced %d apps from Data Service.", count)
	}()

	// Kafka Producer (→ Dev 6)
	dataRelayAdapter, err := kafkaOut.NewDataRelayAdapter(config.KafkaBrokers, config.ActivityTopic)
	if err != nil {
		log.Printf("[MAIN] Warning: Failed to initialize Kafka producer: %v", err)
		// 개발 환경에서는 Kafka 없이도 실행 가능하도록 nil 허용
	}

	// gRPC Clients (지연 연결)
	sabotageAdapter := grpcOut.NewSabotageCommandAdapterLazy(config.SabotageCommandAddr)
	physicalAdapter := grpcOut.NewPhysicalControlAdapterLazy(config.PhysicalControlAddr)
	screenAdapter := grpcOut.NewScreenControlAdapterLazy(config.ScreenControlAddr)
	intelligenceAdapter := grpcOut.NewIntelligenceAdapterLazy(config.IntelligenceAddr)
	log.Printf("[MAIN] gRPC adapters initialized (lazy connection)")

	// 3. Initialize Services
	// ReflexService - 즉각 반응 처리
	reflexService := service.NewReflexService(blacklistAdapter, sabotageAdapter, dataRelayAdapter)
	log.Printf("[MAIN] ReflexService initialized")

	// CommandRouterService - Dev 6 → Dev 1/3 라우팅
	commandRouterService := service.NewCommandRouterService(physicalAdapter, screenAdapter)
	log.Printf("[MAIN] CommandRouterService initialized")

	// SolutionRouterService - Dev 5 → Dev 3 라우팅
	solutionRouterService := service.NewSolutionRouterService(screenAdapter)
	log.Printf("[MAIN] SolutionRouterService initialized")

	// 4. Initialize Adapters (Driving - In)
	// HTTP Handler
	activityHandler := httpAdapter.NewActivityHandler(reflexService)

	// Kafka Consumer (← Dev 6)
	var stateConsumer *kafkaIn.StateConsumer
	if config.KafkaBrokers != "" {
		stateConsumer, err = kafkaIn.NewStateConsumer(
			config.KafkaBrokers,
			"core-decision-service",
			config.StateTopic,
			commandRouterService,
		)
		if err != nil {
			log.Printf("[MAIN] Warning: Failed to initialize Kafka consumer: %v", err)
		} else {
			if err := stateConsumer.Start(); err != nil {
				log.Printf("[MAIN] Warning: Failed to start Kafka consumer: %v", err)
			}
		}
	}

	// 5. Setup Echo HTTP Server
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Register routes
	activityHandler.RegisterRoutes(e)

	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	// Solution Router endpoint (Dev 5 → Dev 3)
	e.POST("/api/v1/solution", func(c echo.Context) error {
		var req struct {
			ClientID string `json:"client_id"`
			Markdown string `json:"markdown"`
		}
		if err := c.Bind(&req); err != nil {
			return c.JSON(400, map[string]string{"error": "Invalid request"})
		}
		if err := solutionRouterService.RouteAIResult(req.ClientID, req.Markdown); err != nil {
			return c.JSON(500, map[string]string{"error": err.Error()})
		}
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	// 6. Start Servers
	// HTTP Server
	go func() {
		log.Printf("[MAIN] Starting HTTP server on port %s", config.HTTPPort)
		if err := e.Start(":" + config.HTTPPort); err != nil {
			log.Printf("[MAIN] HTTP server stopped: %v", err)
		}
	}()

	// ScoreService - 점수 산정 (Algorithmic Logic)
	scoreService := service.NewScoreService()
	log.Printf("[MAIN] ScoreService initialized")

	// gRPC Server (Vision Service Input) on Port 50052
	// gRPC Server (Vision Service Input) on Port 50052
	inputGrpcServer := grpcIn.NewInputGrpcServer("50052", reflexService, scoreService, intelligenceAdapter)
	if err := inputGrpcServer.Start(); err != nil {
		log.Printf("[MAIN] Failed to start Input gRPC server: %v", err)
	}

	// 7. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Printf("[MAIN] Shutting down...")

	// Cleanup
	inputGrpcServer.Stop()
	if stateConsumer != nil {
		stateConsumer.Stop()
	}
	if dataRelayAdapter != nil {
		dataRelayAdapter.Close()
	}
	sabotageAdapter.Close()
	physicalAdapter.Close()
	screenAdapter.Close()

	log.Printf("[MAIN] Shutdown complete")
}

// loadConfig 환경 변수에서 설정 로드
func loadConfig() Config {
	return Config{
		HTTPPort:            getEnv("HTTP_PORT", "8080"),
		KafkaBrokers:        getEnv("KAFKA_BOOTSTRAP_SERVERS", "localhost:9092"),
		ActivityTopic:       getEnv("ACTIVITY_TOPIC", "client-activity"),
		StateTopic:          getEnv("STATE_TOPIC", "command-state"),
		PhysicalControlAddr: getEnv("PHYSICAL_CONTROL_ADDR", "localhost:50051"),
		ScreenControlAddr:   getEnv("SCREEN_CONTROL_ADDR", "localhost:50052"),
		SabotageCommandAddr: getEnv("SABOTAGE_CMD_ADDR", "localhost:50053"),
		IntelligenceAddr:    getEnv("INTELLIGENCE_ADDR", "localhost:50051"),
	}
}

// getEnv 환경 변수 조회 (기본값 지원)
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
