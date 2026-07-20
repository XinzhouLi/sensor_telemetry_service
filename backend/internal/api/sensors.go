package api

import (
	"net/http"
	"time"

	"sensor-telemetry-service/backend/internal/telemetry"
)

type sensorResponse struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Unit          string                 `json:"unit"`
	ValidMin      float64                `json:"valid_min"`
	ValidMax      float64                `json:"valid_max"`
	LatestReading *latestReadingResponse `json:"latest_reading"`
	Health        telemetry.Health       `json:"health"`
}

type latestReadingResponse struct {
	RecordedAt string                  `json:"recorded_at"`
	Value      float64                 `json:"value"`
	Status     telemetry.ReadingStatus `json:"status"`
}

func (s *Server) listSensors(w http.ResponseWriter, r *http.Request) {
	sensors, err := s.store.ListSensors(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	now := s.now()
	response := make([]sensorResponse, 0, len(sensors))
	for _, sensor := range sensors {
		item := sensorResponse{
			ID:       sensor.ID,
			Name:     sensor.Name,
			Unit:     sensor.Unit,
			ValidMin: sensor.ValidMin,
			ValidMax: sensor.ValidMax,
			Health:   telemetry.SensorHealth(sensor.LatestReading, now),
		}
		if sensor.LatestReading != nil {
			item.LatestReading = &latestReadingResponse{
				RecordedAt: sensor.LatestReading.RecordedAt.UTC().Format(time.RFC3339Nano),
				Value:      sensor.LatestReading.Value,
				Status:     sensor.LatestReading.Status,
			}
		}
		response = append(response, item)
	}

	writeJSON(w, http.StatusOK, response)
}
