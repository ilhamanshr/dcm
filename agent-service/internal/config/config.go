package config

import (
	"log/slog"
	"os"
	"strconv"
)

type Config struct {
	ControllerURL string
	APIKey        string
	WorkerURL     string

	RedisAddr     string
	RedisPassword string
	RedisDB       int
}

func Load() Config {
	var (
		err     error
		redisDB int
	)

	redisDBEnv := os.Getenv("REDIS_DB")
	if redisDBEnv != "" {
		redisDB, err = strconv.Atoi(redisDBEnv)
		if err != nil {
			slog.Info("Invalid REDIS_DB value, using default of 0", slog.String("REDIS_DB", redisDBEnv), slog.Any("error", err))
			redisDB = 0 // default value if conversion fails
		}
	}

	return Config{
		ControllerURL: os.Getenv("CONTROLLER_URL"),
		APIKey:        os.Getenv("API_KEY"),
		WorkerURL:     os.Getenv("WORKER_URL"),
		RedisAddr:     os.Getenv("REDIS_ADDR"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		RedisDB:       redisDB,
	}
}
