package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/rs/zerolog/log"
	
	"github.com/panoramablock/thegraph-data-extraction/internal/domain/entity"
)

// Publisher is an adapter for Kafka that implements the ports.EventPublisher interface
type Publisher struct {
	writers       map[string]*kafka.Writer
	brokers       []string
	topicPrefix   string
	producer      string
	flushInterval time.Duration
	batchSize     int
	async         bool
}

// PublisherConfig holds the configuration for the Kafka publisher
type PublisherConfig struct {
	Brokers       []string
	TopicPrefix   string
	Producer      string
	FlushInterval time.Duration
	BatchSize     int
	Async         bool
}

// NewPublisher creates a new Kafka publisher
func NewPublisher(config PublisherConfig) *Publisher {
	// Set default producer name if not provided
	if config.Producer == "" {
		config.Producer = "thegraph-extraction"
	}
	
	// Set default batch size if not provided
	if config.BatchSize <= 0 {
		config.BatchSize = 100
	}
	
	// Set default flush interval if not provided
	if config.FlushInterval <= 0 {
		config.FlushInterval = 1 * time.Second
	}
	
	return &Publisher{
		writers:       make(map[string]*kafka.Writer),
		brokers:       config.Brokers,
		topicPrefix:   config.TopicPrefix,
		producer:      config.Producer,
		flushInterval: config.FlushInterval,
		batchSize:     config.BatchSize,
		async:         config.Async,
	}
}

// getOrCreateWriter gets an existing writer for a topic or creates a new one
func (p *Publisher) getOrCreateWriter(topic string) *kafka.Writer {
	// Check if we already have a writer for this topic
	if writer, exists := p.writers[topic]; exists {
		return writer
	}
	
	// Format the full topic name with prefix if needed
	fullTopic := topic
	if p.topicPrefix != "" {
		fullTopic = fmt.Sprintf("%s.%s", p.topicPrefix, topic)
	}
	
	// Create a new writer
	writer := &kafka.Writer{
		Addr:         kafka.TCP(p.brokers...),
		Topic:        fullTopic,
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    p.batchSize,
		BatchTimeout: p.flushInterval,
		Async:        p.async,
	}
	
	// Store the writer for reuse
	p.writers[topic] = writer
	
	log.Info().
		Str("topic", fullTopic).
		Msg("Created new Kafka writer")
	
	return writer
}

// PublishEntity publishes an entity to the message bus
func (p *Publisher) PublishEntity(ctx context.Context, entity *entity.Entity, topic string) error {
	// Marshal the entity to JSON
	data, err := entity.MarshalForEvent()
	if err != nil {
		return fmt.Errorf("error marshaling entity: %w", err)
	}
	
	// Use the entity ID as the key
	return p.PublishRaw(ctx, entity.ID, data, topic)
}

// PublishRaw publishes raw data to the message bus
func (p *Publisher) PublishRaw(ctx context.Context, key string, data []byte, topic string) error {
	// Get or create a writer for this topic
	writer := p.getOrCreateWriter(topic)
	
	// Create a Kafka message
	msg := kafka.Message{
		Key:   []byte(key),
		Value: data,
		Time:  time.Now(),
		Headers: []kafka.Header{
			{Key: "producer", Value: []byte(p.producer)},
			{Key: "timestamp", Value: []byte(fmt.Sprintf("%d", time.Now().UnixMilli()))},
		},
	}
	
	// Write the message
	err := writer.WriteMessages(ctx, msg)
	if err != nil {
		log.Error().
			Str("topic", topic).
			Str("key", key).
			Err(err).
			Msg("Failed to publish message to Kafka")
		return fmt.Errorf("failed to write message to %s: %w", topic, err)
	}
	
	log.Debug().
		Str("topic", topic).
		Str("key", key).
		Int("dataSize", len(data)).
		Msg("Published message to Kafka")
	
	return nil
}

// Close closes the publisher connection
func (p *Publisher) Close() error {
	var errors []error
	
	// Close all writers
	for topic, writer := range p.writers {
		if err := writer.Close(); err != nil {
			log.Error().
				Str("topic", topic).
				Err(err).
				Msg("Error closing Kafka writer")
			errors = append(errors, fmt.Errorf("error closing writer for %s: %w", topic, err))
		}
	}
	
	// Clear the writers map
	p.writers = make(map[string]*kafka.Writer)
	
	// Return an error if any writers failed to close
	if len(errors) > 0 {
		return fmt.Errorf("failed to close %d writers", len(errors))
	}
	
	return nil
} 