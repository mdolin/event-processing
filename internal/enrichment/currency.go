// Package enrichment This fetches exchange rates, caches them, and converts amounts to EUR.
package enrichment

import (
	"encoding/json"
	config "github.com/Bitstarz-eng/event-processing-challenge/internal"
	"github.com/rs/zerolog/log"
	"net/http"
	"sync"
	"time"
)

// Hardcoded exchange rates as a fallback (different currencies to EUR)
var fallbackRates = map[string]float64{
	"EUR": 1.0,
	"USD": 0.85,
	"GBP": 1.15,
	"NZD": 0.60,
	"BTC": 30000.0, // Example rate
}

// Cache to store exchange rates (valid for 1 minute)
var cache struct {
	sync.Mutex
	rates     map[string]float64
	timestamp time.Time
}

// FetchExchangeRates retrieves conversion rates from API
func FetchExchangeRates() (map[string]float64, error) {
	cache.Lock()
	defer cache.Unlock()

	// Use cached data if less than 1 min old
	if time.Since(cache.timestamp) < time.Minute {
		return cache.rates, nil
	}

	resp, err := http.Get(config.ExchangeRateAPI)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Rates   map[string]float64 `json:"rates"`
		Success bool               `json:"success"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Success {
		log.Info().Msg("Failed to fetch exchange rates, using fallback rates")
		cache.rates = fallbackRates
		cache.timestamp = time.Now()

		return result.Rates, nil
	}

	// Update cache
	cache.rates = result.Rates
	cache.timestamp = time.Now()

	return result.Rates, nil
}

// ConvertToEUR converts given amount to EUR
func ConvertToEUR(amount int, currency string) int {
	rates, err := FetchExchangeRates()
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch exchange rates, returning original amount")
		return amount
	}

	rate, exists := rates[currency]
	if !exists {
		log.Warn().Str("currency", currency).Msg("Unknown currency")
		return amount
	}

	return int(float64(amount) / rate)
}
