package handler

import (
	"controller-service/internal/api/request"
	"controller-service/internal/service"
	"encoding/json"
	"fmt"
	"net/http"
)

type ControllerHandler struct {
	Service service.IControllerService
}

// Register Agent godoc
// @Summary Registe agent
// @Description Register a new agent with a name
// @Tags agents
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body request.RegisterAgentRequest true "Agent registration data"
// @Success 200 {object} response.ConfigResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /register [post]
func (h *ControllerHandler) Register(w http.ResponseWriter, r *http.Request) {
	var body request.RegisterAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	agent, err := h.Service.RegisterAgent(r.Context(), body.Name)
	if err != nil {
		http.Error(w, "Failed to register agent", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(agent)
}

// Get Config godoc
// @Summary Get config
// @Description Get config
// @Tags config
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.ConfigResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /config [get]
func (h *ControllerHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	config, version, err := h.Service.GetConfig(r.Context())
	if err != nil {
		http.Error(w, "Failed to get config", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Version", fmt.Sprint(version))
	json.NewEncoder(w).Encode(config)
}

// Update Config godoc
// @Summary Update config
// @Description Update config
// @Tags config
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body request.UpdateConfigRequest true "Config update data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /config [post]
func (h *ControllerHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	var body request.UpdateConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := body.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.Service.UpdateConfig(r.Context(), body); err != nil {
		http.Error(w, "Failed to update config", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
