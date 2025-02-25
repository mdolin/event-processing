package main

import (
	"context"
	"encoding/json"
	config "github.com/Bitstarz-eng/event-processing-challenge/internal"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/casino"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/enrichment"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/generator"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/materialize"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/pubsub"
	"github.com/rs/zerolog/log"
	"sync"
)

func main() {
	// Load configuration from environment variables
	config.LoadConfig()
	log.Info().Msgf("RabbitMQ URL: %s", config.RabbitMQURL)
	log.Info().Msgf("Exchange Rate API: %s", config.ExchangeRateAPI)
	log.Info().Msgf("Database URL: %s", config.DatabaseURL)

	/// Initialize the database connection pool
	playerRepo, err := enrichment.NewPlayerRepository()
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize player repository")
		return
	}
	defer playerRepo.Close()

	// Create a context to manage event generation
	ctx := context.Background()

	// Set up a wait group for goroutines
	var wg sync.WaitGroup

	// Create an instance of Materializer
	materializer := materialize.NewMaterialize()

	// Start HTTP server in a separate goroutine
	go materializer.StartHTTPServer()

	// Start event generation
	eventCh := generator.Generate(ctx)

	// Publish generated events to RabbitMQ
	wg.Add(1)
	go publishGeneratedEvents(eventCh, playerRepo, &wg)

	// Subscribe to processed events
	wg.Add(1)
	go subscribeToProcessedEvents(materializer, &wg)

	// Wait for all goroutines to finish
	wg.Wait()

	log.Info().Msg("All services stopped. Exiting...")
}

// Publish generated events
func publishGeneratedEvents(eventCh <-chan casino.Event, playerRepo *enrichment.PlayerRepository, wg *sync.WaitGroup) {
	defer wg.Done()
	for event := range eventCh {
		// Convert amount to EUR
		event.AmountEUR = enrichment.ConvertToEUR(event.Amount, event.Currency)

		// Fetch player data using the shared DB connection
		player, err := playerRepo.FetchPlayer(event.PlayerID)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to fetch player %d", event.PlayerID)
		} else {
			event.Player = player
		}

		// Check if player data is missing using IsZero()
		if event.Player.IsZero() {
			log.Warn().Msgf("Player %d not found, leaving player field empty.", event.PlayerID)
		}

		// Generate human-friendly description
		event.Description = enrichment.GenerateDescription(event)

		// Publish event to RabbitMQ
		err = pubsub.PublishEvent(event)
		if err != nil {
			log.Error().Err(err).Msg("Failed to publish event")
		}
	}
}

// Subscribe to processed events
func subscribeToProcessedEvents(materializer *materialize.Materialize, wg *sync.WaitGroup) {
	defer wg.Done()
	err := pubsub.SubscribeEvents(func(event casino.Event) {
		// Update materialized stats
		materializer.UpdateStats(event)
		eventJSON, _ := json.Marshal(event)
		log.Info().Msgf("Processed Event: %s", string(eventJSON))
	})

	if err != nil {
		log.Error().Err(err).Msg("Failed to subscribe to events")
	}
}
