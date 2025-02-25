// Package enrichment This fetches player info from the database.
package enrichment

import (
	"database/sql"
	"errors"
	"fmt"
	config "github.com/Bitstarz-eng/event-processing-challenge/internal"
	"github.com/Bitstarz-eng/event-processing-challenge/internal/casino"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/rs/zerolog/log"
)

// PlayerRepository handles database operations related to players.
type PlayerRepository struct {
	db *sql.DB
}

// NewPlayerRepository creates a new PlayerRepository instance.
func NewPlayerRepository() (*PlayerRepository, error) {
	db, err := sql.Open("postgres", config.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &PlayerRepository{db: db}, nil
}

// Close closes the database connection.
func (r *PlayerRepository) Close() error {
	return r.db.Close()
}

// FetchPlayer retrieves player info using an existing database connection.
func (r *PlayerRepository) FetchPlayer(playerID int) (casino.Player, error) {
	var player casino.Player
	query := `SELECT email, last_signed_in_at FROM players WHERE id = $1`
	err := r.db.QueryRow(query, playerID).Scan(&player.Email, &player.LastSignedInAt)

	if errors.Is(err, sql.ErrNoRows) {
		log.Warn().Msgf("Player %d not found in database", playerID)
		return casino.Player{}, nil // Return zero-value player
	} else if err != nil {
		return casino.Player{}, fmt.Errorf("error fetching player: %w", err)
	}

	return player, nil
}
