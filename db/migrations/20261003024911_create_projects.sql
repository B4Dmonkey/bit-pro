-- migrate:up
CREATE TABLE projects (
    id INTEGER PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    path TEXT NOT NULL UNIQUE
);

-- migrate:down
DROP TABLE projects;
