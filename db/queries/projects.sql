-- name: CreateProject :exec
INSERT INTO projects (path, code) VALUES (?, ?);

-- name: ListProjects :many
SELECT id, code, path FROM projects ORDER BY code;

