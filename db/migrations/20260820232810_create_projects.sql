-- migrate:up
CREATE TABLE projects (
    id INTEGER PRIMARY KEY,
    path TEXT NOT NULL UNIQUE,
    code TEXT NOT NULL
);

-- migrate:down
DROP TABLE projects;
