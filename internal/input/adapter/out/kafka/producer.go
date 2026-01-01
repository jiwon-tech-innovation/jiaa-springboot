package kafka

import (
	"encoding/json"
	"log"
	"sync"
	"sync/atomic"

	"github.com/confluentinc/confluent-kafka-go/kafka"

	"jiaa-server-core/internal/input/domain"
)

// DataRelayAdapter Dev 6으로 데이터 릴레이를 위한 Kafka Producer (Driven Adapter)
// 비동기 방식으로 client-activity 토픽에 ClientActivity를 전송
// ⚡ 핵심: 절대 blocking 하지 않음 → 전체 시스템 렉 방지
type DataRelayAdapter struct {
	producer     *kafka.Producer
	topic        string
	pendingCount int64 // 대기 중인 메시지 수
	wg           sync.WaitGroup
	stopChan     chan struct{}
}

// ActivityMessage Kafka 메시지 구조체
type ActivityMessage struct {
	ClientID     string            `json:"client_id"`
	URL          string            `json:"url,omitempty"`
	AppName      string            `json:"app_name,omitempty"`
	ActivityType string            `json:"activity_type"`
	Timestamp    int64             `json:"timestamp"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// NewDataRelayAdapter DataRelayAdapter 생성자
func NewDataRelayAdapter(brokers string, topic string) (*DataRelayAdapter, error) {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers":            brokers,
		"go.batch.producer":            true,  // 배치 전송 활성화
		"linger.ms":                    5,     // 5ms 동안 메시지 모아서 배치 전송
		"batch.size":                   16384, // 16KB 배치 크기
		"acks":                         "1",   // Leader만 확인 (속도 우선)
		"queue.buffering.max.messages": 100000,
	})
	if err != nil {
		return nil, err
	}

	adapter := &DataRelayAdapter{
		producer: producer,
		topic:    topic,
		stopChan: make(chan struct{}),
	}

	// Background delivery report handler 시작
	adapter.wg.Add(1)
	go adapter.deliveryReportHandler()

	return adapter, nil
}

// deliveryReportHandler 백그라운드에서 delivery report 처리
func (a *DataRelayAdapter) deliveryReportHandler() {
	defer a.wg.Done()

	for {
		select {
		case <-a.stopChan:
			return
		case e := <-a.producer.Events():
			switch ev := e.(type) {
			case *kafka.Message:
				atomic.AddInt64(&a.pendingCount, -1)
				if ev.TopicPartition.Error != nil {
					log.Printf("[DATA_RELAY] Async delivery failed: %v", ev.TopicPartition.Error)
				} else {
					log.Printf("[DATA_RELAY] Async delivered to %s [%d] @ %v",
						*ev.TopicPartition.Topic, ev.TopicPartition.Partition, ev.TopicPartition.Offset)
				}
			case kafka.Error:
				log.Printf("[DATA_RELAY] Kafka error: %v", ev)
			}
		}
	}
}

// RelayToAnalyzer 클라이언트 활동을 Dev 6(분석기)로 비동기 릴레이
// ⚡ Non-blocking: 즉시 반환, 백그라운드에서 전송
func (a *DataRelayAdapter) RelayToAnalyzer(activity domain.ClientActivity) error {
	msg := ActivityMessage{
		ClientID:     activity.ClientID,
		URL:          activity.URL,
		AppName:      activity.AppName,
		ActivityType: string(activity.ActivityType),
		Timestamp:    activity.Timestamp.UnixMilli(),
		Metadata:     activity.Metadata,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// ⚡ 비동기 전송 (deliveryChan = nil → non-blocking)
	err = a.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &a.topic, Partition: kafka.PartitionAny},
		Value:          data,
		Key:            []byte(activity.ClientID),
	}, nil) // nil = async mode

	if err != nil {
		log.Printf("[DATA_RELAY] Failed to enqueue message: %v", err)
		return err
	}

	atomic.AddInt64(&a.pendingCount, 1)
	log.Printf("[DATA_RELAY] Message enqueued (pending: %d)", atomic.LoadInt64(&a.pendingCount))

	return nil
}

// GetPendingCount 대기 중인 메시지 수 반환
func (a *DataRelayAdapter) GetPendingCount() int64 {
	return atomic.LoadInt64(&a.pendingCount)
}

// Flush 모든 대기 메시지 강제 전송
func (a *DataRelayAdapter) Flush(timeoutMs int) int {
	return a.producer.Flush(timeoutMs)
}

// Close Producer 종료 (graceful)
func (a *DataRelayAdapter) Close() {
	log.Printf("[DATA_RELAY] Closing... (pending: %d)", atomic.LoadInt64(&a.pendingCount))

	// 대기 메시지 모두 전송
	remaining := a.producer.Flush(15 * 1000)
	if remaining > 0 {
		log.Printf("[DATA_RELAY] Warning: %d messages not delivered", remaining)
	}

	// Background handler 종료
	close(a.stopChan)
	a.wg.Wait()

	a.producer.Close()
	log.Printf("[DATA_RELAY] Closed")
}
