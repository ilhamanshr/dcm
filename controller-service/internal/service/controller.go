package service

import (
	"context"
	"controller-service/internal/api/request"
	"controller-service/internal/api/response"
	"controller-service/internal/repository"
	queries "controller-service/internal/repository/sqlc"
	"database/sql"
	"encoding/json"
	"log/slog"
)

type ControllerService struct {
	DB   *sql.DB
	Repo repository.IRepository
}

type globalConfig struct {
	URL          string `json:"url"`
	PollInterval int    `json:"poll_interval"`
}

func NewControllerService(db *sql.DB, repo repository.IRepository) IControllerService {
	return &ControllerService{
		DB:   db,
		Repo: repo,
	}
}

//go:generate mockgen -destination=mocks/controller_usecase.go -source=controller.go IControllerService
type IControllerService interface {
	RegisterAgent(ctx context.Context, name string) (*response.ConfigResponse, error)
	GetConfig(ctx context.Context) (*response.ConfigResponse, int, error)
	UpdateConfig(ctx context.Context, payload request.UpdateConfigRequest) error
}

func (s *ControllerService) RegisterAgent(ctx context.Context, name string) (*response.ConfigResponse, error) {
	latestGlobalConfig, err := s.Repo.GetLatestVersionGlobalConfig(ctx)
	if err != nil {
		slog.Error("RegisterAgent Failed to fetch global config", slog.Any("error", err))
		return nil, err
	}

	var globalConfig globalConfig
	err = json.Unmarshal(latestGlobalConfig.Config, &globalConfig)
	if err != nil {
		slog.Error("RegisterAgent Failed to unmarshal global config", slog.Any("error", err))
		return nil, err
	}

	agentID, err := s.Repo.CreateAgent(ctx, name)
	if err != nil {
		slog.Error("RegisterAgent Failed to create agent", slog.Any("error", err))
		return nil, err
	}

	return &response.ConfigResponse{
		AgentID:      agentID.String(),
		PollURL:      globalConfig.URL,
		PollInterval: globalConfig.PollInterval,
	}, nil
}

func (s *ControllerService) GetConfig(ctx context.Context) (*response.ConfigResponse, int, error) {
	latestGlobalConfig, err := s.Repo.GetLatestVersionGlobalConfig(ctx)
	if err != nil {
		slog.Error("GetConfig Failed to fetch global config", slog.Any("error", err))
		return nil, 0, err
	}

	var globalConfig globalConfig
	err = json.Unmarshal(latestGlobalConfig.Config, &globalConfig)
	if err != nil {
		slog.Error("GetConfig Failed to unmarshal global config", slog.Any("error", err))
		return nil, 0, err
	}

	return &response.ConfigResponse{
		PollURL:      globalConfig.URL,
		PollInterval: globalConfig.PollInterval,
	}, int(latestGlobalConfig.Version), nil
}

func (s *ControllerService) UpdateConfig(ctx context.Context, payload request.UpdateConfigRequest) error {
	tx, err := s.DB.Begin()
	if err != nil {
		slog.Error("UpdateConfig Failed to begin transaction", slog.Any("error", err))
		return err
	}
	defer tx.Rollback()

	queryTx := s.Repo.WithTx(tx)

	latestGlobalConfig, err := queryTx.GetLatestVersionGlobalConfig(ctx)
	if err != nil {
		slog.Error("UpdateConfig Failed to fetch global config", slog.Any("error", err))
		return err
	}

	var latestConfig globalConfig
	err = json.Unmarshal(latestGlobalConfig.Config, &latestConfig)
	if err != nil {
		slog.Error("UpdateConfig Failed to unmarshal global config", slog.Any("error", err))
		return err
	}

	if latestConfig.URL == payload.URL && latestConfig.PollInterval == payload.PollInterval {
		slog.Info("UpdateConfig config is already up to date", slog.Any("url", payload.URL), slog.Any("poll_interval", payload.PollInterval))
		return nil
	}

	globalConfig := globalConfig{
		URL:          payload.URL,
		PollInterval: payload.PollInterval,
	}

	configBytes, err := json.Marshal(globalConfig)
	if err != nil {
		slog.Error("UpdateConfig Failed to marshal global config", slog.Any("error", err))
		return err
	}

	if _, err = queryTx.CreateGlobalConfig(ctx, queries.CreateGlobalConfigParams{
		Config:  configBytes,
		Version: latestGlobalConfig.Version + 1,
	}); err != nil {
		slog.Error("UpdateConfig Failed to create global config", slog.Any("error", err))
		return err
	}

	tx.Commit()

	return nil
}
