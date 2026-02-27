package config

import (
	"log/slog"
	"os"
	"strconv"
)

type Config struct {
	AppPort     string
	DBURL       string
	APIKey      string
	PollSeconds int
}

func Load() Config {
	var (
		pollSeconds int
		err         error
	)

	pollSecondsEnv := os.Getenv("POLL_SECONDS")
	if pollSecondsEnv != "" {
		pollSeconds, err = strconv.Atoi(pollSecondsEnv)
		if err != nil {
			slog.Info("Invalid POLL_SECONDS value, using default of 30 seconds", slog.String("POLL_SECONDS", pollSecondsEnv), slog.Any("error", err))
			pollSeconds = 30 // default value if conversion fails
		}
	}

	appPort := os.Getenv("APP_PORT")
	if appPort == "" {
		appPort = "8080"
	}

	return Config{
		AppPort:     appPort,
		DBURL:       os.Getenv("DB_URL"),
		APIKey:      os.Getenv("API_KEY"),
		PollSeconds: pollSeconds,
	}
}
