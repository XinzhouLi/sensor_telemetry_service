package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"sensor-telemetry-service/backend/internal/telemetry"
)

func TestSummarizeReadings(t *testing.T) {
	average, minimum, maximum := 41.2, 40.9, 41.5
	store := &fakeStore{
		findSensorFound: true,
		summaries: []telemetry.SummaryBucket{
			{
				BucketStart:     time.Date(2026, time.July, 20, 8, 0, 0, 0, time.FixedZone("MDT", -6*60*60)),
				Average:         &average,
				Minimum:         &minimum,
				Maximum:         &maximum,
				ValidCount:      2,
				OutOfRangeCount: 1,
			},
			{
				BucketStart:     time.Date(2026, time.July, 20, 15, 0, 0, 0, time.UTC),
				ValidCount:      0,
				OutOfRangeCount: 2,
			},
		},
	}
	path := "/sensors/nox-analyzer-1/summary?from=2026-07-20T08:15:00-06:00&to=2026-07-20T10:00:00-06:00"
	response := performRequest(t, store, time.Now(), http.MethodGet, path)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", response.Code, response.Body.String())
	}
	var body []summaryResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body) != 2 {
		t.Fatalf("len(response) = %d, want 2", len(body))
	}
	if body[0].BucketStart != "2026-07-20T14:00:00Z" || body[0].Average == nil || *body[0].Average != average {
		t.Fatalf("first bucket = %#v", body[0])
	}
	if body[1].Average != nil || body[1].Minimum != nil || body[1].Maximum != nil {
		t.Fatalf("out-of-range-only bucket statistics = %#v", body[1])
	}
	wantFrom := time.Date(2026, time.July, 20, 14, 15, 0, 0, time.UTC)
	wantTo := time.Date(2026, time.July, 20, 16, 0, 0, 0, time.UTC)
	if store.summarySensorID != "nox-analyzer-1" || !store.summaryFrom.Equal(wantFrom) || !store.summaryTo.Equal(wantTo) {
		t.Fatalf("summary received sensor=%q from=%s to=%s", store.summarySensorID, store.summaryFrom, store.summaryTo)
	}
}

func TestSummarizeReadingsInvalidWindow(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{name: "missing parameter", query: "from=2026-07-20T14:00:00Z"},
		{name: "invalid timestamp", query: "from=bad&to=2026-07-20T15:00:00Z"},
		{name: "invalid order", query: "from=2026-07-20T15:00:00Z&to=2026-07-20T14:00:00Z"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := &fakeStore{findSensorErr: errors.New("must not be called")}
			response := performRequest(t, store, time.Now(), http.MethodGet, "/sensors/unknown/summary?"+test.query)
			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400", response.Code)
			}
		})
	}
}

func TestSummarizeReadingsUnknownSensor(t *testing.T) {
	path := "/sensors/unknown/summary?from=2026-07-20T14:00:00Z&to=2026-07-20T15:00:00Z"
	response := performRequest(t, &fakeStore{}, time.Now(), http.MethodGet, path)
	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", response.Code)
	}
}

func TestSummarizeReadingsEmptyArray(t *testing.T) {
	store := &fakeStore{findSensorFound: true}
	path := "/sensors/nox-analyzer-1/summary?from=2026-07-20T14:00:00Z&to=2026-07-20T15:00:00Z"
	response := performRequest(t, store, time.Now(), http.MethodGet, path)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}
	if got := strings.TrimSpace(response.Body.String()); got != "[]" {
		t.Fatalf("body = %s, want []", got)
	}
}

func TestSummarizeReadingsDatabaseErrors(t *testing.T) {
	path := "/sensors/nox-analyzer-1/summary?from=2026-07-20T14:00:00Z&to=2026-07-20T15:00:00Z"
	tests := []struct {
		name  string
		store *fakeStore
	}{
		{name: "find sensor", store: &fakeStore{findSensorErr: errors.New("database unavailable")}},
		{
			name: "summarize readings",
			store: &fakeStore{
				findSensorFound: true,
				summaryErr:      errors.New("database unavailable"),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := performRequest(t, test.store, time.Now(), http.MethodGet, path)
			if response.Code != http.StatusInternalServerError {
				t.Fatalf("status = %d, want 500", response.Code)
			}
			if got := strings.TrimSpace(response.Body.String()); got != `{"error":"internal server error"}` {
				t.Fatalf("body = %s", got)
			}
		})
	}
}
