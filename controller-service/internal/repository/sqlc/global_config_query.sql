-- name: GetLatestVersionGlobalConfig :one
SELECT * 
FROM 
    global_config 
ORDER BY 
    version DESC LIMIT 1;

-- name: CreateGlobalConfig :one
INSERT INTO global_config (config, version)
VALUES ($1, $2)
RETURNING *;