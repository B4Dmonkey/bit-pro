-- migrate:up
CREATE UNIQUE INDEX queue_project_target ON queue (project_id, target_id);

-- migrate:down
DROP INDEX queue_project_target;
