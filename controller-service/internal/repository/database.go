package repository

import (
	"context"
	queries "controller-service/internal/repository/sqlc"
	"database/sql"

	"github.com/google/uuid"
)

//go:generate mockgen -destination=mocks/mock_database.go -source=database.go IRepository
type IRepository interface {
	// Transaction
	WithTx(tx *sql.Tx) *queries.Queries

	// Global Config
	CreateGlobalConfig(ctx context.Context, arg queries.CreateGlobalConfigParams) (queries.GlobalConfig, error)
	GetLatestVersionGlobalConfig(ctx context.Context) (queries.GlobalConfig, error)

	// Agent
	CreateAgent(ctx context.Context, name string) (uuid.UUID, error)
}
