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
	host    string
	topic   string
	verbose bool
	dryrun  bool
}

var (
	plugin = Config{
		PluginConfig: sensu.PluginConfig{
			Name:     "sensu-kafka-handler",
			Short:    "Sensu handler to convert Sensu events to Kafka messages ",
			Keyspace: "sensu.io/plugins/sensu-kafka-handler/config",
		},
	}

	options = []*sensu.PluginConfigOption{
		&sensu.PluginConfigOption{
			Path:      "host",
			Env:       "KAFKA_HOST",
			Argument:  "host",
			Shorthand: "H",
			Default:   "localhost:9092",
			Usage:     "The Kafka broker host, defaults to value of KAFKA_HOST env variable",
			Value:     &plugin.host,
		},
		&sensu.PluginConfigOption{
			Path:      "topic",
			Env:       "KAFKA_TOPIC",
			Argument:  "topic",
			Shorthand: "t",
			Default:   "sensu-event",
			Usage:     "Kafka topic to post to, defaults to value of KAFKA_TOPIC env variable",
			Value:     &plugin.topic,
		},
		&sensu.PluginConfigOption{
			Argument:  "verbose",
			Shorthand: "v",
			Default:   false,
			Usage:     "Verbose output to stdout, useful for testing",
			Value:     &plugin.verbose,
		},
		&sensu.PluginConfigOption{
			Argument:  "dryrun",
			Shorthand: "n",
			Default:   false,
			Usage:     "Dryrun, do not connect to Kafka broker",
			Value:     &plugin.dryrun,
		},
	}
)

func newKafkaWriter(host, topic string) *kafka.Writer {
	return kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{host},
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	})
}

func main() {
	handler := sensu.NewGoHandler(&plugin.PluginConfig, options, checkArgs, executeHandler)
	handler.Execute()
}

func checkArgs(_ *types.Event) error {
	if len(plugin.topic) == 0 {
		return fmt.Errorf("--topic or KAFKA_TOPIC environment variable is required")
	}
	if len(plugin.host) == 0 {
		return fmt.Errorf("--host or KAFKA_HOST environment variable is required")
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
	writer := newKafkaWriter(plugin.host, plugin.topic)
	defer writer.Close()

	msg := kafka.Message{
		Key:   idBytes,
		Value: eventBytes,
	}
	if plugin.dryrun {
		fmt.Printf("Dryrun enabled, reporting configured settings\n")
		fmt.Printf("  Kafka Broker Host: %v\n", plugin.host)
		fmt.Printf("  Kafka Topic: %v\n", plugin.topic)
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
