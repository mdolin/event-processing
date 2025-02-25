package enrichment

import (
	"testing"
	"time"

	"github.com/Bitstarz-eng/event-processing-challenge/internal/casino"
)

func TestGenerateDescription(t *testing.T) {
	// Define test cases
	tests := []struct {
		name     string
		event    casino.Event
		expected string
	}{
		{
			name: "game_start event",
			event: casino.Event{
				PlayerID:  15,
				GameID:    1,
				Type:      "game_start",
				CreatedAt: time.Date(2025, 2, 19, 20, 50, 0, 0, time.UTC),
			},
			expected: "Player #15 started playing a game \"Western Gold 2\" on February 19, 2025 at 20:50 UTC.",
		},
		{
			name: "bet event",
			event: casino.Event{
				PlayerID:  15,
				GameID:    1,
				Type:      "bet",
				Amount:    1392,
				Currency:  "NZD",
				AmountEUR: 2320,
				CreatedAt: time.Date(2025, 2, 19, 20, 50, 0, 0, time.UTC),
			},
			expected: "Player #15 placed a bet of 1392 NZD (2320 EUR) on \"Western Gold 2\" on February 19, 2025 at 20:50 UTC.",
		},
		{
			name: "deposit event",
			event: casino.Event{
				PlayerID:  15,
				Type:      "deposit",
				Amount:    500,
				Currency:  "USD",
				CreatedAt: time.Date(2025, 2, 19, 20, 50, 0, 0, time.UTC),
			},
			expected: "Player #15 made a deposit of 500 USD on February 19, 2025 at 20:50 UTC.",
		},
		{
			name: "unknown event",
			event: casino.Event{
				Type: "unknown",
			},
			expected: "Unknown event",
		},
	}

	// Mock game titles
	casino.Games = map[int]casino.Game{
		1: {Title: "Western Gold 2"},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateDescription(tt.event)
			if got != tt.expected {
				t.Errorf("GenerateDescription() = %v, want %v", got, tt.expected)
			}
		})
	}
}
