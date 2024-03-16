package kafkaapp

import (
	"context"
	"fmt"

	"github.com/segmentio/kafka-go"
)

var (
	kafkaBrokers        = []string{"localhost:9092"}
	kafkaTopic   string = "users"
)

type Kafkaapp struct {
	w *kafka.Writer
	r *kafka.Reader
}

func NewKafka() *Kafkaapp {
	return &Kafkaapp{
		w: kafka.NewWriter(kafka.WriterConfig{
			Brokers: kafkaBrokers,
		}),
		r: kafka.NewReader(kafka.ReaderConfig{
			Brokers: kafkaBrokers,
			Topic:   kafkaTopic,
		}),
	}
}

// Функция для продюсирования событий в Kafka
func (k *Kafkaapp) ProduceKafkaEvent(message string) error {
	msg := kafka.Message{
		Topic: kafkaTopic,
		Value: []byte(message),
	}
	return k.w.WriteMessages(context.Background(), msg)
}

// Функция для консьюминга событий из Kafka
func (k *Kafkaapp) ConsumeKafkaEvent(msgChan chan string) {
	for {
		msg, err := k.r.ReadMessage(context.Background())
		if err != nil {
			fmt.Printf("Ошибка чтения сообщения: %s\n", err)
			continue
		}
		msgChan <- string(msg.Value)
	}
}
