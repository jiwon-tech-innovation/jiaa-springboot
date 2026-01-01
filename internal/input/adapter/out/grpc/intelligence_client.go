package grpc

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"jiaa-server-core/pkg/proto"
)

// IntelligenceAdapter Dev 5(Intelligence Worker) gRPC 클라이언트 (Driven Adapter)
// Emergency 상황에서 AI 분석 요청
type IntelligenceAdapter struct {
	conn    *grpc.ClientConn
	client  proto.IntelligenceServiceClient
	address string
	timeout time.Duration
}

// NewIntelligenceAdapter IntelligenceAdapter 생성자
func NewIntelligenceAdapter(address string) (*IntelligenceAdapter, error) {
	conn, err := grpc.NewClient(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return &IntelligenceAdapter{
		conn:    conn,
		client:  proto.NewIntelligenceServiceClient(conn),
		address: address,
		timeout: 30 * time.Second, // AI 분석은 시간이 걸릴 수 있음
	}, nil
}

// NewIntelligenceAdapterLazy 지연 연결 생성자
func NewIntelligenceAdapterLazy(address string) *IntelligenceAdapter {
	return &IntelligenceAdapter{
		address: address,
		timeout: 30 * time.Second,
	}
}

// Connect gRPC 연결 수립
func (a *IntelligenceAdapter) Connect() error {
	if a.conn != nil {
		return nil
	}

	conn, err := grpc.NewClient(a.address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return err
	}

	a.conn = conn
	a.client = proto.NewIntelligenceServiceClient(conn)
	log.Printf("[INTELLIGENCE] Connected to Dev 5: %s", a.address)
	return nil
}

// RequestLogAnalysis 에러 로그 분석 요청 (Emergency Protocol)
// Dev 6가 EMERGENCY 상태 전송 시 호출
func (a *IntelligenceAdapter) RequestLogAnalysis(clientID string, errorLog string, screamText string) (string, error) {
	if a.conn == nil {
		if err := a.Connect(); err != nil {
			log.Printf("[INTELLIGENCE] Failed to connect: %v", err)
			return "", err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), a.timeout)
	defer cancel()

	req := &proto.LogAnalysisRequest{
		ClientId:   clientID,
		ErrorLog:   errorLog,
		ScreamText: screamText,
	}

	log.Printf("[INTELLIGENCE] Requesting log analysis from Dev 5: Client: %s", clientID)
	log.Printf("[INTELLIGENCE] ErrorLog length: %d, ScreamText: %s", len(errorLog), screamText)

	resp, err := a.client.AnalyzeLog(ctx, req)
	if err != nil {
		log.Printf("[INTELLIGENCE] gRPC call failed: %v", err)
		return "", err
	}

	if !resp.Success {
		log.Printf("[INTELLIGENCE] Log analysis failed")
		return "", nil
	}

	log.Printf("[INTELLIGENCE] Log analysis received. Confidence: %.2f, ErrorType: %s",
		resp.Confidence, resp.ErrorType)

	return resp.Markdown, nil
}

// RequestURLClassification URL/Title 분석 요청
func (a *IntelligenceAdapter) RequestURLClassification(clientID string, url string, title string) (string, error) {
	if a.conn == nil {
		if err := a.Connect(); err != nil {
			log.Printf("[INTELLIGENCE] Failed to connect: %v", err)
			return "", err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := &proto.URLClassifyRequest{
		ClientId: clientID,
		Url:      url,
		Title:    title,
	}

	log.Printf("[INTELLIGENCE] Requesting URL classification: URL: %s, Title: %s", url, title)

	resp, err := a.client.ClassifyURL(ctx, req)
	if err != nil {
		log.Printf("[INTELLIGENCE] gRPC call failed: %v", err)
		return "", err
	}

	classification := "UNKNOWN"
	switch resp.Classification {
	case proto.URLClassification_STUDY:
		classification = "STUDY"
	case proto.URLClassification_PLAY:
		classification = "PLAY"
	case proto.URLClassification_NEUTRAL:
		classification = "NEUTRAL"
	case proto.URLClassification_WORK:
		classification = "WORK"
	}

	log.Printf("[INTELLIGENCE] URL classification result: %s (confidence: %.2f)",
		classification, resp.Confidence)

	return classification, nil
}

// Close 연결 종료
func (a *IntelligenceAdapter) Close() error {
	if a.conn != nil {
		return a.conn.Close()
	}
	return nil
}
