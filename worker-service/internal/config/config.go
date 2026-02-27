package config

import "os"

type Config struct {
	AppPort string
	APIKey  string
}

func Load() Config {
	appPort := os.Getenv("APP_PORT")
	if appPort == "" {
		appPort = "8081"
	}

	return Config{
		AppPort: appPort,
		APIKey:  os.Getenv("API_KEY"),
	}
}
