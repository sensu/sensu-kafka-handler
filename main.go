package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	kafka "github.com/segmentio/kafka-go"
	"github.com/sensu-community/sensu-plugin-sdk/sensu"
	"github.com/sensu/sensu-go/types"
)

// Config represents the handler plugin config.
type Config struct {
	sensu.PluginConfig
	KafkaURL   string
	KafkaTopic string
	Verbose    bool
	Dryrun     bool
}

var (
	plugin = Config{
		PluginConfig: sensu.PluginConfig{
			Name:     "sensu-kafa-handler",
			Short:    "Sensu handler to convert Sensu events to Kafka messages ",
			Keyspace: "sensu.io/plugins/sensu-kafa-handler/config",
		},
	}

	options = []*sensu.PluginConfigOption{
		&sensu.PluginConfigOption{
			Env:       "KAFKA_URL",
			Argument:  "kafka-url",
			Shorthand: "u",
			Default:   "localhost:9092",
			Usage:     "The Kafka broker url",
			Value:     &plugin.KafkaURL,
		},
		&sensu.PluginConfigOption{
			Env:       "KAFKA_TOPIC",
			Argument:  "topic",
			Shorthand: "t",
			Default:   "sensu-event",
			Usage:     "Kafka topic to post to",
			Value:     &plugin.KafkaTopic,
		},
		&sensu.PluginConfigOption{
			Argument:  "verbose",
			Shorthand: "v",
			Default:   false,
			Usage:     "Verbose output to stdout, useful for testing",
			Value:     &plugin.Verbose,
		},
		&sensu.PluginConfigOption{
			Argument:  "dryrun",
			Shorthand: "n",
			Default:   false,
			Usage:     "Dryrun, do not connect to Kafka broker",
			Value:     &plugin.Dryrun,
		},
	}
)

func newKafkaWriter(kafkaURL, topic string) *kafka.Writer {
	return kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{kafkaURL},
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	})
}

func main() {
	handler := sensu.NewGoHandler(&plugin.PluginConfig, options, checkArgs, executeHandler)
	handler.Execute()
}

func checkArgs(_ *types.Event) error {
	if len(plugin.KafkaTopic) == 0 {
		return fmt.Errorf("--topic or KAFKA_TOPIC environment variable is required")
	}
	if len(plugin.KafkaURL) == 0 {
		return fmt.Errorf("--kafka-url or KAFKA_URL environment variable is required")
	}
	return nil
}

func executeHandler(event *types.Event) error {

	eventBytes, err := json.Marshal(event)
	if err != nil {
		fmt.Println(err)
		return err
	}
	uid, err := uuid.FromBytes(event.ID)
	if err != nil {
		fmt.Println(err)
		return err
	}
	idBytes, err := json.Marshal(uid.String())
	if err != nil {
		fmt.Println(err)
		return err
	}
	writer := newKafkaWriter(plugin.KafkaURL, plugin.KafkaTopic)
	defer writer.Close()

	msg := kafka.Message{
		Key:   idBytes,
		Value: eventBytes,
	}
	if plugin.Dryrun {
		fmt.Printf("Dryrun enabled, reporting configured settings\n")
		fmt.Printf("  Kafka Broker Url: %v\n", plugin.KafkaURL)
		fmt.Printf("  Kafka Topic: %v\n", plugin.KafkaTopic)
		fmt.Printf("  Kafka Message Key: %s\n", string(idBytes))
		fmt.Printf("  Kafka Message Value:\n %v\n", string(eventBytes))
	} else {
		err := writer.WriteMessages(context.Background(), msg)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	return nil
}
