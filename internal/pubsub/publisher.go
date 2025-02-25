// Package pubsub This connects to RabbitMQ and publishes events to the casino_events queue
// Make sure RabbitMQ is running (docker-compose up -d) before running this program
package pubsub

import (
	"encoding/json"
	config "github.com/Bitstarz-eng/event-processing-challenge/internal"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/casino"
	"github.com/rs/zerolog/log"

	"github.com/streadway/amqp"
)

// PublishEvent sends generated events to RabbitMQ
func PublishEvent(event casino.Event) error {
	// Connect to RabbitMQ
	conn, err := amqp.Dial(config.RabbitMQURL)
	if err != nil {
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	// Declare a queue for event publishing
	q, err := ch.QueueDeclare(
		"casino_events", // Queue name
		true,            // Durable
		false,           // Auto-delete
		false,           // Exclusive
		false,           // No-wait
		nil,             // Arguments
	)
	if err != nil {
		return err
	}

	// Serialize the event to JSON
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// Publish event to RabbitMQ
	err = ch.Publish(
		"",     // Exchange
		q.Name, // Routing key (queue name)
		false,  // Mandatory
		false,  // Immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        eventJSON,
		},
	)

	if err != nil {
		return err
	}

	log.Info().Msgf("Published event: %s", string(eventJSON))
	return nil
}
