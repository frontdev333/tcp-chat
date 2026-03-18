package server

import (
	"encoding/json"
	"frontdev333/tcp-chat/internal/hub"
	"log/slog"
	"net/http"
)

type StatsProvider interface {
	GetStats() hub.ServerStats
}

func StartHTTPMonitoring(hub StatsProvider, port string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealthEndpoint(hub))
	mux.HandleFunc("/stats", handleStatsEndpoint(hub))
	return http.ListenAndServe(":"+port, mux)
}

func handleHealthEndpoint(h StatsProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats := h.GetStats()
		dto := hub.HealthCheck{
			Status:            "healthy",
			ActiveConnections: stats.ActiveConnections,
			UptimeSeconds:     stats.UptimeSeconds,
		}

		jsonBytes, err := json.MarshalIndent(dto, "", "	")
		if err != nil {
			slog.Error(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err = w.Write(jsonBytes); err != nil {
			slog.Error(err.Error())
			return
		}
	}

}

func handleStatsEndpoint(h StatsProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats := h.GetStats()
		jsonBytes, err := json.MarshalIndent(stats, "", "	")
		if err != nil {
			slog.Error(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err = w.Write(jsonBytes); err != nil {
			slog.Error(err.Error())
			return
		}
	}
}
