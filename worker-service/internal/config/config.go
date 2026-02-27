package config

import "os"

type Config struct {
	AppPort     string
	APIKey      string
	TLSCertFile string
	TLSKeyFile  string
}

func Load() Config {
	appPort := os.Getenv("APP_PORT")
	if appPort == "" {
		appPort = "8081"
	}

	return Config{
		AppPort:     appPort,
		APIKey:      os.Getenv("API_KEY"),
		TLSCertFile: os.Getenv("TLS_CERT_FILE"),
		TLSKeyFile:  os.Getenv("TLS_KEY_FILE"),
	}
}
