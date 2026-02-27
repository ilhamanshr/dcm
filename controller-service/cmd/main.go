// @title Distributed Config Controller API
// @version 1.0
// @description Central configuration management service
// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
package main

import (
	"controller-service/internal/api/handler"
	"controller-service/internal/api/middleware"
	"controller-service/internal/config"
	"controller-service/internal/database"
	queries "controller-service/internal/repository/sqlc"
	"controller-service/internal/service"
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	_ "controller-service/docs" // swagger docs

	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // PostgreSQL driver
	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	if err := godotenv.Load(); err != nil {
		slog.Warn(".env file not found, using system environment variables")
	}

	cfg := config.Load()

	dbConn, errDB := sql.Open("postgres", cfg.DBURL)
	dbConn.SetMaxIdleConns(5)
	dbConn.SetConnMaxIdleTime(10 * time.Second)
	dbConn.SetMaxOpenConns(90)
	if errDB != nil {
		slog.Error("Failed to open database connection", slog.Any("error", errDB))
		panic(errDB)
	}

	if err := dbConn.Ping(); err != nil {
		slog.Error("Failed to ping database", slog.Any("error", err))
		panic(err)
	}
	slog.Info("Database connection successful")

	if err := database.MigrateAll(dbConn); err != nil {
		slog.Error("Failed to migrate database", slog.Any("error", err))
		panic(err.Error())
	}
	slog.Info("Database migration successful")

	queries := queries.New(dbConn)

	svc := &service.ControllerService{
		DB:   dbConn,
		Repo: queries,
	}

	h := &handler.ControllerHandler{Service: svc}

	mux := http.NewServeMux()
	auth := middleware.APIKeyAuth(cfg.APIKey)

	mux.Handle("POST /register", auth(http.HandlerFunc(h.Register)))
	mux.Handle("GET /config", auth(http.HandlerFunc(h.GetConfig)))
	mux.Handle("POST /config", auth(http.HandlerFunc(h.UpdateConfig)))

	mux.Handle("/docs/", httpSwagger.WrapHandler)

	slog.Info("Starting HTTPS server at :" + cfg.AppPort)
	if err := http.ListenAndServeTLS(":"+cfg.AppPort, cfg.TLSCertFile, cfg.TLSKeyFile, mux); err != nil {
		slog.Error("ListenAndServeTLS: ", slog.Any("error", err))
		panic(err)
	}
}
