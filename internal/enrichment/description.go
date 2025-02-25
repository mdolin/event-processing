// Package enrichment This maps events to human-friendly descriptions.
package enrichment

import (
	"fmt"

	"github.com/Bitstarz-eng/event-processing-challenge/internal/casino"
)

// GenerateDescription creates human-readable event descriptions
func GenerateDescription(event casino.Event) string {
	timestamp := event.CreatedAt.Format("January 2, 2006 at 15:04 UTC")
	gameTitle := casino.Games[event.GameID].Title

	switch event.Type {
	case "game_start":
		return fmt.Sprintf("Player #%d started playing a game \"%s\" on %s.",
			event.PlayerID, gameTitle, timestamp)
	case "bet":
		return fmt.Sprintf("Player #%d placed a bet of %d %s (%d EUR) on \"%s\" on %s.",
			event.PlayerID, event.Amount, event.Currency, event.AmountEUR, gameTitle, timestamp)
	case "deposit":
		return fmt.Sprintf("Player #%d made a deposit of %d %s on %s.", event.PlayerID, event.Amount, event.Currency, timestamp)
	default:
		return "Unknown event"
	}
}
