package materialize

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/Bitstarz-eng/event-processing-challenge/internal/casino"
	"github.com/rs/zerolog/log"
)

// Stats represents the materialized data.
type Stats struct {
	EventsTotal              int         `json:"events_total"`
	EventsPerMinute          float64     `json:"events_per_minute"`
	EventsPerSecondMovingAvg float64     `json:"events_per_second_moving_average"`
	TopPlayerBets            PlayerStats `json:"top_player_bets"`
	TopPlayerWins            PlayerStats `json:"top_player_wins"`
	TopPlayerDeposits        PlayerStats `json:"top_player_deposits"`
}

// PlayerStats represents the statistics for a player.
type PlayerStats struct {
	ID    int `json:"id"`
	Count int `json:"count"`
}

// Materialize holds the state and methods for materializing events.
type Materialize struct {
	stats           Stats
	mu              sync.Mutex
	eventTimestamps []time.Time
	playerBets      map[int]int
	playerWins      map[int]int
	playerDeposits  map[int]int
}

// NewMaterializer creates a new Materializer instance.
func NewMaterialize() *Materialize {
	return &Materialize{
		playerBets:     make(map[int]int),
		playerWins:     make(map[int]int),
		playerDeposits: make(map[int]int),
	}
}

// UpdateStats updates the materialized data with the given event.
// It increments the total event count, updates the moving average
// of events per second over the last 60 seconds, and tracks the
// top players by number of bets, wins, and sum of deposits in EUR.
func (m *Materialize) UpdateStats(event casino.Event) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.stats.EventsTotal++

	// Track event timestamps for moving average calculation
	m.eventTimestamps = append(m.eventTimestamps, time.Now())

	// Remove old timestamps (keep last 60 seconds)
	oneMinuteAgo := time.Now().Add(-time.Minute)
	for len(m.eventTimestamps) > 0 && m.eventTimestamps[0].Before(oneMinuteAgo) {
		m.eventTimestamps = m.eventTimestamps[1:]
	}

	// Calculate events per second (moving average over last 60 seconds)
	m.stats.EventsPerSecondMovingAvg = float64(len(m.eventTimestamps)) / 60.0

	// Calculate events per minute (scaled from total events in last 60s)
	m.stats.EventsPerMinute = float64(len(m.eventTimestamps))

	// Process event types
	switch event.Type {
	case "bet":
		m.playerBets[event.PlayerID]++
	case "game_stop":
		if event.HasWon {
			m.playerWins[event.PlayerID]++
		}
	case "deposit":
		m.playerDeposits[event.PlayerID] += event.AmountEUR // Track in EUR
	}

	// Update top players
	m.stats.TopPlayerBets = m.getTopPlayer(m.playerBets)
	m.stats.TopPlayerWins = m.getTopPlayer(m.playerWins)
	m.stats.TopPlayerDeposits = m.getTopPlayer(m.playerDeposits)
}

// GetStats handles HTTP requests to retrieve the materialized data.
func (m *Materialize) GetStats(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(m.stats)
	if err != nil {
		return
	}
}

// StartHTTPServer starts the HTTP server to serve the materialized data.
func (m *Materialize) StartHTTPServer() {
	http.HandleFunc("/materialized", m.GetStats)
	log.Info().Msg("Materialized data available at http://localhost:8080/materialized")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		return
	}
}

// getTopPlayer finds the player with the highest count in a given map.
func (m *Materialize) getTopPlayer(playerMap map[int]int) PlayerStats {
	var topPlayer PlayerStats
	for id, count := range playerMap {
		if count > topPlayer.Count {
			topPlayer.ID = id
			topPlayer.Count = count
		}
	}
	return topPlayer
}
