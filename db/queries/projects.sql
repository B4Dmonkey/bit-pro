-- name: CreateProject :exec
INSERT INTO projects (path, code) VALUES (?, ?);

-- name: ProjectExists :one
SELECT EXISTS(SELECT 1 FROM projects WHERE path = ?);

-- name: ListProjects :many
SELECT id, path, code FROM projects ORDER BY code;

-- name: UpdateProjectCounts :exec
UPDATE projects SET backlog = ?, todo = ?, done = ?, completed = ? WHERE id = ?;
