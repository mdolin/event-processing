// Package pubsub This connects to RabbitMQ and consume events from the casino_events queue
// Make sure RabbitMQ is running (docker-compose up -d) before running this program
package pubsub

import (
	"encoding/json"
	config "github.com/Bitstarz-eng/event-processing-challenge/internal"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/casino"
	"github.com/rs/zerolog/log"
	"github.com/streadway/amqp"
)

// SubscribeEvents listens for incoming events from RabbitMQ
func SubscribeEvents(processEvent func(event casino.Event)) error {
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

	// Declare queue before consuming
	q, err := ch.QueueDeclare(
		"casino_events",
		true, false, false, false, nil,
	)
	if err != nil {
		return err
	}

	// Consume messages from queue
	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		return err
	}

	// Process each message
	go func() {
		for d := range msgs {
			var event casino.Event
			err := json.Unmarshal(d.Body, &event)
			if err != nil {
				log.Error().Err(err).Msg("Failed to unmarshal event")
				continue
			}
			processEvent(event)
		}
	}()

	log.Info().Msg("Subscribed to events")
	select {} // Keep the function running
}
