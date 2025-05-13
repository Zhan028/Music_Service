package kafka

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"log"
	"time"
)

func PublishMessage(ctx context.Context, writer *kafka.Writer, topic string, key string, payload interface{}) {
	msgBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("failed to marshal payload for topic %s: %v", topic, err)
		return
	}

	err = writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: msgBytes,
		Time:  time.Now(),
	})
	if err != nil {
		log.Printf("failed to publish to kafka topic %s: %v", topic, err)
	}

}
