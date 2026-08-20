-- name: CreateProject :exec
INSERT INTO projects (path, code) VALUES (?, ?);

-- name: ListProjects :many
SELECT id, path, code FROM projects;
