-- migrate:up
CREATE TABLE queue (
    id INTEGER PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES projects(id),
    target_id TEXT NOT NULL,
    target_typ TEXT NOT NULL
);

-- migrate:down
DROP TABLE queue;
