package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"sensor-telemetry-service/backend/internal/telemetry"
)

type fakeStore struct {
	pingErr         error
	sensors         []telemetry.Sensor
	listErr         error
	findSensor      telemetry.Sensor
	findSensorFound bool
	findSensorErr   error
	writeResults    []telemetry.WriteResult
	writeErr        error
	writtenSensorID string
	writtenReadings []telemetry.Reading
	writeCalls      int
	readings        []telemetry.Reading
	readErr         error
	readSensorID    string
	readFrom        time.Time
	readTo          time.Time
	summaries       []telemetry.SummaryBucket
	summaryErr      error
	summarySensorID string
	summaryFrom     time.Time
	summaryTo       time.Time
}

func (s *fakeStore) Ping(context.Context) error {
	return s.pingErr
}

func (s *fakeStore) ListSensors(context.Context) ([]telemetry.Sensor, error) {
	return s.sensors, s.listErr
}

func (s *fakeStore) FindSensor(context.Context, string) (telemetry.Sensor, bool, error) {
	return s.findSensor, s.findSensorFound, s.findSensorErr
}

func (s *fakeStore) InsertReadings(
	_ context.Context,
	sensorID string,
	readings []telemetry.Reading,
) ([]telemetry.WriteResult, error) {
	s.writeCalls++
	s.writtenSensorID = sensorID
	s.writtenReadings = append([]telemetry.Reading(nil), readings...)
	return s.writeResults, s.writeErr
}

func (s *fakeStore) ListReadings(
	_ context.Context,
	sensorID string,
	from time.Time,
	to time.Time,
) ([]telemetry.Reading, error) {
	s.readSensorID = sensorID
	s.readFrom = from
	s.readTo = to
	return s.readings, s.readErr
}

func (s *fakeStore) SummarizeReadings(
	_ context.Context,
	sensorID string,
	from time.Time,
	to time.Time,
) ([]telemetry.SummaryBucket, error) {
	s.summarySensorID = sensorID
	s.summaryFrom = from
	s.summaryTo = to
	return s.summaries, s.summaryErr
}

func TestListSensors(t *testing.T) {
	now := time.Date(2026, time.July, 20, 14, 0, 0, 0, time.UTC)
	store := &fakeStore{sensors: []telemetry.Sensor{
		{
			ID:       "nox-analyzer-1",
			Name:     "NOx Analyzer 1",
			Unit:     "ppm",
			ValidMin: 0,
			ValidMax: 250,
			LatestReading: &telemetry.Reading{
				RecordedAt: now.Add(-5 * time.Minute).In(time.FixedZone("MDT", -6*60*60)),
				Value:      41.2,
				Status:     telemetry.ReadingStatusValid,
			},
		},
		{
			ID:       "o2-analyzer-1",
			Name:     "Oxygen Analyzer 1",
			Unit:     "%",
			ValidMin: 0,
			ValidMax: 25,
			LatestReading: &telemetry.Reading{
				RecordedAt: now.Add(-30 * time.Minute),
				Value:      20.8,
				Status:     telemetry.ReadingStatusValid,
			},
		},
		{
			ID:       "stack-temp-1",
			Name:     "Stack Temperature 1",
			Unit:     "°C",
			ValidMin: 0,
			ValidMax: 600,
		},
	}}

	response := performRequest(t, store, now, http.MethodGet, "/sensors")
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if got := response.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}

	var body []sensorResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body) != 3 {
		t.Fatalf("len(response) = %d, want 3", len(body))
	}
	if body[0].Health != telemetry.HealthOK {
		t.Errorf("NOx health = %q, want %q", body[0].Health, telemetry.HealthOK)
	}
	if body[0].LatestReading == nil {
		t.Fatal("NOx latest_reading is nil")
	}
	if body[0].LatestReading.RecordedAt != "2026-07-20T13:55:00Z" {
		t.Errorf("recorded_at = %q, want UTC RFC 3339", body[0].LatestReading.RecordedAt)
	}
	if body[1].Health != telemetry.HealthStale {
		t.Errorf("O2 health = %q, want %q", body[1].Health, telemetry.HealthStale)
	}
	if body[2].Health != telemetry.HealthNeverReported || body[2].LatestReading != nil {
		t.Errorf("never-reported sensor = %#v", body[2])
	}

}

func TestListSensorsEmptyArray(t *testing.T) {
	response := performRequest(t, &fakeStore{}, time.Now(), http.MethodGet, "/sensors")
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if got := strings.TrimSpace(response.Body.String()); got != "[]" {
		t.Fatalf("body = %s, want []", got)
	}
}

func TestListSensorsDatabaseError(t *testing.T) {
	store := &fakeStore{listErr: errors.New("database unavailable")}
	response := performRequest(t, store, time.Now(), http.MethodGet, "/sensors")

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusInternalServerError)
	}
	if got := strings.TrimSpace(response.Body.String()); got != `{"error":"internal server error"}` {
		t.Fatalf("body = %s", got)
	}
}

func TestHealth(t *testing.T) {
	tests := []struct {
		name       string
		pingErr    error
		wantStatus int
		wantBody   string
	}{
		{name: "available", wantStatus: http.StatusOK, wantBody: `{"status":"ok"}`},
		{
			name:       "unavailable",
			pingErr:    errors.New("connection refused"),
			wantStatus: http.StatusServiceUnavailable,
			wantBody:   `{"status":"unavailable"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := performRequest(t, &fakeStore{pingErr: test.pingErr}, time.Now(), http.MethodGet, "/health")
			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, test.wantStatus)
			}
			if got := strings.TrimSpace(response.Body.String()); got != test.wantBody {
				t.Fatalf("body = %s, want %s", got, test.wantBody)
			}
		})
	}
}

func TestMethodNotAllowed(t *testing.T) {
	response := performRequest(t, &fakeStore{}, time.Now(), http.MethodPost, "/sensors")
	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusMethodNotAllowed)
	}
	if got := response.Header().Get("Allow"); got != "GET, HEAD" {
		t.Fatalf("Allow = %q, want GET, HEAD", got)
	}
}

func performRequest(
	t *testing.T,
	store Store,
	now time.Time,
	method string,
	path string,
) *httptest.ResponseRecorder {
	t.Helper()
	handler := newServer(store, func() time.Time { return now })
	request := httptest.NewRequest(method, path, nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}
