-- name: CreateProject :exec
INSERT INTO projects (path, code) VALUES (?, ?);

-- name: ListProjects :many
SELECT id, code, path, removed FROM projects ORDER BY code;

-- name: SetProjectRemoved :exec
UPDATE projects SET removed = ? WHERE id = ?;

