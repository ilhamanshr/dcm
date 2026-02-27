package handler

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"sync"
)

type WorkerHandler struct {
	mu     sync.RWMutex
	config WorkerConfig
}

// WorkerConfig holds the worker's runtime configuration.
type WorkerConfig struct {
	URL string `json:"url"`
}

func New() *WorkerHandler {
	return &WorkerHandler{
		config: WorkerConfig{},
	}
}

// UpdateConfig godoc
// @Summary Update worker config
// @Description Set the URL the worker should hit
// @Tags config
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body WorkerConfig true "Worker config"
// @Success 200
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /config [post]
func (s *WorkerHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var cfg WorkerConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if cfg.URL == "" {
		slog.Error("worker config update failed: url is empty")
		http.Error(w, "url is empty", 400)
		return
	}

	s.config.URL = cfg.URL

	slog.Info("worker config updated:", slog.Any("config", s.config))

	w.WriteHeader(http.StatusOK)
}

// Hit godoc
// @Summary Hit the configured URL
// @Description Makes a GET request to the configured URL and returns the response body
// @Tags hit
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {string} string
// @Failure 400 {string} string
// @Failure 500 {string} string
// @Router /hit [get]
func (s *WorkerHandler) Hit(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	url := s.config.URL
	if url == "" {
		slog.Error("worker hit failed: url is empty")
		http.Error(w, "url is empty", 400)
		return
	}

	resp, err := http.Get(url)
	if err != nil {
		slog.Error("worker hit failed to get url", slog.Any("error", err))
		http.Error(w, err.Error(), 500)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("worker hit failed to read body", slog.Any("error", err))
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write(body)
}
