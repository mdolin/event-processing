// Package config This holds application settings
package config

import (
	"os"

	"github.com/rs/zerolog/log"
)

// Config holds all application settings

var RabbitMQURL string
var ExchangeRateAPI string
var DatabaseURL string

// LoadConfig reads environment variables and sets up config
func LoadConfig() {
	RabbitMQURL = getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	ExchangeRateAPI = getEnv("EXCHANGE_RATE_API", "https://api.exchangerate.host/latest")
	DatabaseURL = getEnv("DATABASE_URL", "postgres://casino:casino@localhost:5432/casino?sslmode=disable")

	log.Info().Msg("Configuration loaded successfully")
}

// getEnv fetches an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
