package service

import (
	"agent-service/internal/repository"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

type AgentService struct {
	controllerURL   string
	workerURL       string
	apiKey          string
	cache           repository.ICache
	agentID         string
	poolingInterval int
}

type configResponse struct {
	AgentID      string `json:"agent_id"`
	PollURL      string `json:"poll_url"`
	PollInterval int    `json:"poll_interval"`
	Version      int    `json:"version"`
}

type workerConfig struct {
	URL string `json:"url"`
}

func NewAgentService(controllerURL, workerURL, apiKey string, cache repository.ICache) IAgentService {
	return &AgentService{
		controllerURL: controllerURL,
		workerURL:     workerURL,
		apiKey:        apiKey,
		cache:         cache,
	}
}

//go:generate mockgen -destination=mocks/agent.go -source=agent.go IAgentService
type IAgentService interface {
	RegisterAgent(ctx context.Context) error
}

func (p *AgentService) RegisterAgent(ctx context.Context) error {
	agentName := fmt.Sprintf("agent-%s", randomString(6))
	body, err := json.Marshal(map[string]string{"name": agentName})
	if err != nil {
		slog.Error("RegisterAgent failed to marshal registration data:", slog.Any("error", err))
		return err
	}

	req, err := http.NewRequest("POST", p.controllerURL+"/register", bytes.NewBuffer(body))
	if err != nil {
		slog.Error("RegisterAgent failed to create registration request:", slog.Any("error", err))
		return err
	}

	req.Header.Set("X-API-Key", p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("RegisterAgent failed to do registration request:", slog.Any("error", err))
		return err
	}
	defer resp.Body.Close()

	var regResp configResponse
	if err := json.NewDecoder(resp.Body).Decode(&regResp); err != nil {
		slog.Error("RegisterAgent failed to decode registration response:", slog.Any("error", err))
		return err
	}

	p.agentID = regResp.AgentID
	p.poolingInterval = regResp.PollInterval
	if p.poolingInterval == 0 {
		p.poolingInterval = 5 // default pooling interval
	}

	regRespJSON, err := json.Marshal(regResp)
	if err != nil {
		slog.Error("RegisterAgent failed to marshal config response", slog.Any("error", err))
		return err
	}

	if err := p.cache.SetKey(ctx, fmt.Sprintf("config_agent:%s", regResp.AgentID), string(regRespJSON)); err != nil {
		slog.Error("RegisterAgent failed to set key", slog.Any("error", err))
		return err
	}

	slog.Info("Registered with controller, starting poller")

	p.pooling(ctx)

	return nil
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func (p *AgentService) pooling(ctx context.Context) {
	backoff := time.Second

	// start polling with backoff
	for {
		err := p.configCheck(ctx)
		if err != nil {
			slog.Error("pooling failed to check config", slog.Any("error", err))
			time.Sleep(backoff)
			if backoff < 30*time.Second {
				backoff *= 2
			}
			continue
		}

		backoff = time.Second
		time.Sleep(time.Duration(p.poolingInterval) * time.Second)
	}
}

func (p *AgentService) configCheck(ctx context.Context) error {
	var cachedConfig configResponse
	cachedConfigString, err := p.cache.GetKey(ctx, fmt.Sprintf("config_agent:%s", p.agentID))
	if err != nil {
		slog.Error("configCheck failed to get old config from cache", slog.Any("error", err))
		return err
	}

	if err := json.Unmarshal([]byte(cachedConfigString), &cachedConfig); err != nil {
		slog.Error("configCheck failed to unmarshal old config", slog.Any("error", err))
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.controllerURL+"/config", nil)
	if err != nil {
		slog.Error("configCheck failed to create request", slog.Any("error", err))
		return err
	}

	req.Header.Set("X-API-Key", p.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("configCheck failed to do request", slog.Any("error", err))
		return err
	}
	defer resp.Body.Close()

	var newConfig configResponse
	if err := json.NewDecoder(resp.Body).Decode(&newConfig); err != nil {
		slog.Error("configCheck failed to decode config", slog.Any("error", err))
		return err
	}

	if resp.StatusCode != http.StatusOK {
		slog.Error("configCheck failed to get config", slog.Any("status", resp.StatusCode))
		return errors.New("configCheck failed to get config")
	}

	versionString := resp.Header.Get("Version")
	if versionString == "" {
		slog.Error("configCheck failed to get version from header")
		return errors.New("version not found in header")
	}

	newVersion, err := strconv.Atoi(versionString)
	if err != nil {
		slog.Error("configCheck failed to convert version to int", slog.Any("error", err))
		return err
	}

	// check version
	if newVersion == cachedConfig.Version {
		slog.Info("configCheck config is up to date", slog.Any("version", newVersion))
		return nil
	}

	slog.Info("configCheck config is out of date, sending new config", slog.Any("version", newVersion))

	// update cached config
	newConfig.AgentID = p.agentID
	newConfig.Version = newVersion

	cfgJSON, err := json.Marshal(newConfig)
	if err != nil {
		slog.Error("configCheck failed to marshal config", slog.Any("error", err))
		return err
	}

	// send config to worker
	if err := p.sendConfig(workerConfig{
		URL: newConfig.PollURL,
	}); err != nil {
		slog.Error("configCheck failed to send config", slog.Any("error", err))
		return err
	}

	// update cached config
	if err := p.cache.SetKey(ctx, fmt.Sprintf("config_agent:%s", p.agentID), string(cfgJSON)); err != nil {
		slog.Error("configCheck failed to set key", slog.Any("error", err))
		return err
	}

	// update pooling interval
	p.poolingInterval = newConfig.PollInterval

	return nil
}

func (c *AgentService) sendConfig(cfg workerConfig) error {
	body, err := json.Marshal(cfg)
	if err != nil {
		slog.Error("sendConfig Failed to marshal config", slog.Any("error", err))
		return err
	}

	req, err := http.NewRequest(http.MethodPost, c.workerURL+"/config", bytes.NewBuffer(body))
	if err != nil {
		slog.Error("sendConfig Failed to create request", slog.Any("error", err))
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("sendConfig Failed to send config", slog.Any("error", err))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("sendConfig failed to send config", slog.Any("status", resp.StatusCode))
		return errors.New("sendConfig failed to send config")
	}

	slog.Info("sendConfig update config to worker success")
	return nil
}
