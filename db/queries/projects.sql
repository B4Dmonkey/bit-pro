-- name: CreateProject :exec
INSERT INTO projects (path, code) VALUES (?, ?);

-- name: ProjectExists :one
SELECT EXISTS(SELECT 1 FROM projects WHERE path = ?);

-- name: ListProjects :many
SELECT id, code, path FROM projects ORDER BY code;

