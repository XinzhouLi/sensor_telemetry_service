package api

import (
	"net/http"
	"time"
)

type summaryResponse struct {
	BucketStart     string   `json:"bucket_start"`
	Average         *float64 `json:"average"`
	Minimum         *float64 `json:"minimum"`
	Maximum         *float64 `json:"maximum"`
	ValidCount      int64    `json:"valid_count"`
	OutOfRangeCount int64    `json:"out_of_range_count"`
}

func (s *Server) summarizeReadings(w http.ResponseWriter, r *http.Request) {
	from, to, err := parseTimeWindow(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	sensorID := r.PathValue("id")
	_, found, err := s.store.FindSensor(r.Context(), sensorID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}
	if !found {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "sensor not found"})
		return
	}

	buckets, err := s.store.SummarizeReadings(r.Context(), sensorID, from, to)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	response := make([]summaryResponse, len(buckets))
	for index, bucket := range buckets {
		response[index] = summaryResponse{
			BucketStart:     bucket.BucketStart.UTC().Format(time.RFC3339Nano),
			Average:         bucket.Average,
			Minimum:         bucket.Minimum,
			Maximum:         bucket.Maximum,
			ValidCount:      bucket.ValidCount,
			OutOfRangeCount: bucket.OutOfRangeCount,
		}
	}
	writeJSON(w, http.StatusOK, response)
}
