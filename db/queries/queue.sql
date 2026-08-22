-- name: EnqueueTask :exec
INSERT OR IGNORE INTO queue (project_id, target_id, target_typ) VALUES (?, ?, ?);

-- name: ListQueueByProject :many
SELECT id, project_id, target_id, target_typ FROM queue WHERE project_id = ? ORDER BY id;
