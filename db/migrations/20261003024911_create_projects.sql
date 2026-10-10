-- migrate:up
CREATE TABLE projects (
    id INTEGER PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    path TEXT NOT NULL UNIQUE,
    removed INTEGER NOT NULL DEFAULT 0
);

-- migrate:down
DROP TABLE projects;
