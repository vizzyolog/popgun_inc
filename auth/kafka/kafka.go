package kafka

import (
	_ "embed"
	"fmt"
	"log"
	"time"

	"github.com/dangkaka/go-kafka-avro"

	"popug_auth/model"
)

type CudUser struct {
	producer *kafka.AvroProducer
	schema   string
	topic    string
}

var kafkaServers = []string{"localhost:9092"}
var schemaRegistryServers = []string{"http://localhost:8081"}
var topic = "users"

func ProducerUserUPD() *CudUser {
	schemaEmbeded := `{
		"type": "record",
		"name": "Example",
		"fields": [
		  {"name": "Id", "type": "string"},
		  {"name": "Type", "type": "string"},
		  {"name": "Data", "type": "string"}
		]
	  }`

	producerAvro, err := kafka.NewAvroProducer(kafkaServers, schemaRegistryServers)
	if err != nil {
		fmt.Printf("Could not create avro producer: %s", err)
	}

	return &CudUser{
		producer: producerAvro,
		schema:   schemaEmbeded,
		topic:    topic,
	}
}

func (c *CudUser) AddMsg(userdata model.User) {

	value := `{
		"Id": "1",
		"Type": "example_type",
		"Data": "example_data"
	}`

	key := time.Now().String()
	err := c.producer.Add(c.topic, c.schema, []byte(key), []byte(value))
	if err != nil {
		log.Fatal("err produce msg", err)
	}
}
