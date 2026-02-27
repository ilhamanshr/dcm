-- name: CreateAgent :one
INSERT INTO agents (name) VALUES ($1) RETURNING id;