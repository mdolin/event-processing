package materialize

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Bitstarz-eng/event-processing-challenge/internal/casino"
)

func TestUpdateStats(t *testing.T) {
	// Create an instance of Materializer
	materializer := NewMaterialize()

	event := casino.Event{
		PlayerID:  1,
		Type:      "bet",
		Amount:    100,
		AmountEUR: 85,
		CreatedAt: time.Now(),
	}

	materializer.UpdateStats(event)

	if materializer.stats.EventsTotal != 1 {
		t.Errorf("Expected EventsTotal to be 1, got %d", materializer.stats.EventsTotal)
	}

	if materializer.stats.EventsPerMinute != 1 {
		t.Errorf("Expected EventsPerMinute to be 1, got %f", materializer.stats.EventsPerMinute)
	}

	if materializer.stats.EventsPerSecondMovingAvg != 1.0/60.0 {
		t.Errorf("Expected EventsPerSecondMovingAvg to be %f, got %f", 1.0/60.0, materializer.stats.EventsPerSecondMovingAvg)
	}

	if materializer.stats.TopPlayerBets.ID != 1 || materializer.stats.TopPlayerBets.Count != 1 {
		t.Errorf("Expected TopPlayerBets to be {ID: 1, Count: 1}, got %+v", materializer.stats.TopPlayerBets)
	}
}

func TestGetStats(t *testing.T) {
	// Create an instance of Materializer
	materializer := NewMaterialize()

	event := casino.Event{
		PlayerID:  1,
		Type:      "bet",
		Amount:    100,
		AmountEUR: 85,
		CreatedAt: time.Now(),
	}

	materializer.UpdateStats(event)

	req, err := http.NewRequest("GET", "/materialized", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(materializer.GetStats)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response Stats
	err = json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatal(err)
	}

	if response.EventsTotal != 1 {
		t.Errorf("Expected EventsTotal to be 1, got %d", response.EventsTotal)
	}

	if response.TopPlayerBets.ID != 1 || response.TopPlayerBets.Count != 1 {
		t.Errorf("Expected TopPlayerBets to be {ID: 1, Count: 1}, got %+v", response.TopPlayerBets)
	}
}
