package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"sensor-telemetry-service/backend/internal/telemetry"
)

type Store interface {
	Ping(context.Context) error
	ListSensors(context.Context) ([]telemetry.Sensor, error)
	FindSensor(context.Context, string) (telemetry.Sensor, bool, error)
	InsertReadings(context.Context, string, []telemetry.Reading) ([]telemetry.WriteResult, error)
	ListReadings(context.Context, string, time.Time, time.Time) ([]telemetry.Reading, error)
	SummarizeReadings(context.Context, string, time.Time, time.Time) ([]telemetry.SummaryBucket, error)
}

type Server struct {
	store Store
	now   func() time.Time
}

func NewServer(store Store) http.Handler {
	return newServer(store, time.Now)
}

func newServer(store Store, now func() time.Time) http.Handler {
	server := &Server{store: store, now: now}
	router := http.NewServeMux()
	router.HandleFunc("GET /health", server.health)
	router.HandleFunc("GET /sensors", server.listSensors)
	router.HandleFunc("POST /sensors/{id}/readings", server.ingestReadings)
	router.HandleFunc("GET /sensors/{id}/readings", server.listReadings)
	router.HandleFunc("GET /sensors/{id}/summary", server.summarizeReadings)
	return router
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	if err := s.store.Ping(r.Context()); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
