// @title Worker Service API
// @version 1.0
// @description Worker service for hitting configured URLs
// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
package main

import (
	"log/slog"
	"net/http"
	"worker-service/internal/api/handler"
	"worker-service/internal/api/middleware"
	"worker-service/internal/config"

	_ "worker-service/docs"

	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	if err := godotenv.Load(); err != nil {
		slog.Warn(".env file not found, using system environment variables")
	}

	cfg := config.Load()
	srv := handler.New()

	mux := http.NewServeMux()
	auth := middleware.APIKeyAuth(cfg.APIKey)

	mux.Handle("POST /config", auth(http.HandlerFunc(srv.UpdateConfig)))
	mux.Handle("GET /hit", auth(http.HandlerFunc(srv.Hit)))

	mux.Handle("/docs/", httpSwagger.WrapHandler)

	slog.Info("Starting server at :" + cfg.AppPort)
	if err := http.ListenAndServe(":"+cfg.AppPort, mux); err != nil {
		slog.Error("ListenAndServe: ", slog.Any("error", err))
		panic(err)
	}
}
