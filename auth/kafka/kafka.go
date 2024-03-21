package kafka

import (
	"context"
	"log"

	"github.com/dangkaka/go-kafka-avro"
)

var (
	kafkaBrokers                 = []string{"localhost:9092"}
	schemaRegistryServers        = []string{"http://localhost:8081"}
	kafkaTopic            string = "users"
)

func NewProducer() {
	producer, err := kafka.NewAvroProducer(kafkaBrokers, schemaRegistryServers)
	if err != nil {
		log.Fatal(err)
	}
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
func (k *Kafkaapp) ProduceKafkaEvent(message []byte) error {
	msg := kafka.Message{
		Topic: kafkaTopic,
		Value: message,
	}
	return k.w.WriteMessages(context.Background(), msg)
}

// Функция для консьюминга событий из Kafka
func (k *Kafkaapp) ConsumeKafkaEvent() {
	for {
		msg, err := k.r.ReadMessage(context.Background())
		if err != nil {
			log.Fatal("err to read from Kafka", err)
			continue
		}
		log.Printf("kafka msg: %s", msg.Value)
	}
}
