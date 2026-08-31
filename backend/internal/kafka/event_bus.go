package kafka

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"razorpay-recovery-intelligence/backend/internal/domain"

	"github.com/segmentio/kafka-go"
)

type EventHandler func(ctx context.Context, envelope domain.EventEnvelope) error

type EventBus struct {
	brokers        []string
	enabled        bool
	groupID        string
	writers        map[domain.EventType]*kafka.Writer
	handlers       map[domain.EventType][]EventHandler
	mu             sync.RWMutex
	inMemChan      chan domain.EventEnvelope
	stopChan       chan struct{}
	consumerTopics map[domain.EventType]struct{}
}

func NewEventBus(brokers string, enabled bool, groupID string) *EventBus {
	var brokerList []string
	if brokers != "" {
		brokerList = []string{brokers}
	}

	bus := &EventBus{
		brokers:        brokerList,
		enabled:        enabled,
		groupID:        groupID,
		writers:        make(map[domain.EventType]*kafka.Writer),
		handlers:       make(map[domain.EventType][]EventHandler),
		inMemChan:      make(chan domain.EventEnvelope, 1000),
		stopChan:       make(chan struct{}),
		consumerTopics: make(map[domain.EventType]struct{}),
	}

	if enabled {
		log.Printf("[Kafka] Initializing Kafka EventBus with brokers: %s", brokers)
	} else {
		log.Printf("[Kafka] Kafka is disabled or in local-standalone mode. Using asynchronous In-Memory EventBus.")
	}

	// Start in-memory event dispatcher worker
	go bus.runInMemDispatcher()

	return bus
}

func (b *EventBus) RegisterHandler(eventType domain.EventType, handler EventHandler) {
	b.mu.Lock()
	b.handlers[eventType] = append(b.handlers[eventType], handler)
	startConsumer := b.enabled
	if _, started := b.consumerTopics[eventType]; started {
		startConsumer = false
	} else if startConsumer {
		b.consumerTopics[eventType] = struct{}{}
	}
	b.mu.Unlock()
	log.Printf("[EventBus] Registered consumer handler for event: %s", eventType)
	if startConsumer {
		go b.runKafkaConsumer(eventType)
	}
}

// runKafkaConsumer delivers persisted Kafka events to local handlers. The
// recovery pipeline itself is idempotent, so a local synchronous response and
// a later broker delivery safely resolve to the same decision.
func (b *EventBus) runKafkaConsumer(eventType domain.EventType) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  b.brokers,
		GroupID:  b.groupID,
		Topic:    string(eventType),
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	defer reader.Close()

	for {
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			select {
			case <-b.stopChan:
				cancel()
			case <-ctx.Done():
			}
		}()
		message, err := reader.FetchMessage(ctx)
		cancel()
		if err != nil {
			select {
			case <-b.stopChan:
				return
			default:
				log.Printf("[Kafka] Consumer error on topic %s: %v", eventType, err)
				time.Sleep(time.Second)
				continue
			}
		}

		var envelope domain.EventEnvelope
		if err := json.Unmarshal(message.Value, &envelope); err != nil {
			log.Printf("[Kafka] Ignoring invalid message on topic %s: %v", eventType, err)
			_ = reader.CommitMessages(context.Background(), message)
			continue
		}
		b.dispatch(context.Background(), envelope)
		if err := reader.CommitMessages(context.Background(), message); err != nil {
			log.Printf("[Kafka] Commit error on topic %s: %v", eventType, err)
		}
	}
}

func (b *EventBus) Publish(ctx context.Context, envelope domain.EventEnvelope) error {
	data, err := json.Marshal(envelope)
	if err != nil {
		return err
	}

	log.Printf("[EventBus] Publishing event [%s] correlation_id=%s", envelope.EventType, envelope.CorrelationID)

	if b.enabled && len(b.brokers) > 0 {
		writer := b.getWriter(envelope.EventType)
		err := writer.WriteMessages(ctx, kafka.Message{
			Key:   []byte(envelope.CorrelationID.String()),
			Value: data,
			Time:  time.Now(),
		})
		if err != nil {
			log.Printf("[Kafka] Write error on topic %s: %v. Emitting to internal dispatcher.", envelope.EventType, err)
			b.inMemChan <- envelope
			return nil
		}
		return nil
	}

	// Internal asynchronous dispatcher
	select {
	case b.inMemChan <- envelope:
	default:
		log.Printf("[EventBus] Internal channel full, dispatching synchronously for %s", envelope.EventType)
		go b.dispatch(ctx, envelope)
	}

	return nil
}

func (b *EventBus) getWriter(eventType domain.EventType) *kafka.Writer {
	b.mu.Lock()
	defer b.mu.Unlock()

	writer, exists := b.writers[eventType]
	if !exists {
		writer = &kafka.Writer{
			Addr:                   kafka.TCP(b.brokers...),
			Topic:                  string(eventType),
			Balancer:               &kafka.LeastBytes{},
			AllowAutoTopicCreation: true,
			Async:                  true,
		}
		b.writers[eventType] = writer
	}
	return writer
}

func (b *EventBus) runInMemDispatcher() {
	for {
		select {
		case <-b.stopChan:
			return
		case env := <-b.inMemChan:
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			b.dispatch(ctx, env)
			cancel()
		}
	}
}

func (b *EventBus) dispatch(ctx context.Context, envelope domain.EventEnvelope) {
	b.mu.RLock()
	handlers := b.handlers[envelope.EventType]
	b.mu.RUnlock()

	for _, h := range handlers {
		if err := h(ctx, envelope); err != nil {
			log.Printf("[EventBus] Handler execution error for %s: %v", envelope.EventType, err)
		}
	}
}

func (b *EventBus) Close() {
	close(b.stopChan)
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, w := range b.writers {
		_ = w.Close()
	}
}
