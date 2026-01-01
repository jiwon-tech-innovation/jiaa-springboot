package kafka

import (
	"encoding/json"
	"log"

	"github.com/confluentinc/confluent-kafka-go/kafka"

	"jiaa-server-core/internal/input/domain"
	portin "jiaa-server-core/internal/input/port/in"
)

// StateConsumer Dev 6 상태 수신을 위한 Kafka Consumer (Driving Adapter)
// command-state 토픽에서 StateCommand를 수신하여 StateReceiverUseCase에 전달
type StateConsumer struct {
	consumer     *kafka.Consumer
	stateUseCase portin.StateReceiverUseCase
	topic        string
	running      bool
}

// StateMessage Kafka 메시지 구조체
type StateMessage struct {
	ClientID  string `json:"client_id"`
	State     string `json:"state"`
	Payload   string `json:"payload,omitempty"`
	Priority  int    `json:"priority"`
	Timestamp int64  `json:"timestamp"`
}

// NewStateConsumer StateConsumer 생성자
func NewStateConsumer(
	brokers string,
	groupID string,
	topic string,
	stateUseCase portin.StateReceiverUseCase,
) (*StateConsumer, error) {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": brokers,
		"group.id":          groupID,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		return nil, err
	}

	return &StateConsumer{
		consumer:     consumer,
		stateUseCase: stateUseCase,
		topic:        topic,
		running:      false,
	}, nil
}

// Start 상태 메시지 소비 시작
func (c *StateConsumer) Start() error {
	if err := c.consumer.Subscribe(c.topic, nil); err != nil {
		return err
	}

	c.running = true
	log.Printf("[STATE_CONSUMER] Started consuming from topic: %s", c.topic)

	go c.consumeLoop()
	return nil
}

// Stop 소비 중지
func (c *StateConsumer) Stop() {
	c.running = false
	c.consumer.Close()
	log.Printf("[STATE_CONSUMER] Stopped")
}

// consumeLoop 메시지 소비 루프
func (c *StateConsumer) consumeLoop() {
	for c.running {
		msg, err := c.consumer.ReadMessage(-1)
		if err != nil {
			log.Printf("[STATE_CONSUMER] Error reading message: %v", err)
			continue
		}

		c.handleMessage(msg)
	}
}

// handleMessage 메시지 처리
func (c *StateConsumer) handleMessage(msg *kafka.Message) {
	var stateMsg StateMessage
	if err := json.Unmarshal(msg.Value, &stateMsg); err != nil {
		log.Printf("[STATE_CONSUMER] Failed to unmarshal message: %v", err)
		return
	}

	log.Printf("[STATE_CONSUMER] Received state: Client: %s, State: %s",
		stateMsg.ClientID, stateMsg.State)

	// Convert to domain entity
	cmd := domain.NewStateCommand(stateMsg.ClientID, domain.CommandState(stateMsg.State)).
		WithPriority(stateMsg.Priority).
		WithPayload([]byte(stateMsg.Payload))

	// Process through StateReceiverUseCase
	if err := c.stateUseCase.HandleStateChange(*cmd); err != nil {
		log.Printf("[STATE_CONSUMER] Failed to handle state change: %v", err)
	}
}
