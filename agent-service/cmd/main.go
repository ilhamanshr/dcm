package main

import (
	"context"
	"log"
	"log/slog"

	"agent-service/internal/config"
	"agent-service/internal/repository/redis"
	"agent-service/internal/service"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		slog.Warn(".env file not found, using system environment variables")
	}

	cfg := config.Load()

	cache := redis.NewRedisHelper(redis.RedisConfig{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	ctx := context.Background()
	if err := cache.Ping(ctx); err != nil {
		log.Fatal("failed to ping redis:", err)
	}

	slog.Info("Redis connection successful")

	agentService := service.NewAgentService(
		cfg.ControllerURL,
		cfg.WorkerURL,
		cfg.APIKey,
		cache,
	)

	if err := agentService.RegisterAgent(ctx); err != nil {
		log.Fatal(err)
	}
}
