-- migrate:up
CREATE TABLE usage (
    id INTEGER PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    session_id TEXT NOT NULL,
    agent_id TEXT NOT NULL DEFAULT '',
    agent_type TEXT NOT NULL DEFAULT '',
    parent_agent_id TEXT NOT NULL DEFAULT '',
    invoked_agent TEXT NOT NULL DEFAULT '',
    skill TEXT NOT NULL DEFAULT '',
    skill_invocation TEXT NOT NULL DEFAULT '',
    track TEXT NOT NULL,
    bar TEXT NOT NULL DEFAULT '',
    message_id TEXT NOT NULL,
    request_id TEXT NOT NULL DEFAULT '',
    entry INTEGER NOT NULL DEFAULT 0,
    ts TEXT NOT NULL,
    model TEXT NOT NULL,
    speed TEXT NOT NULL DEFAULT '',
    is_sidechain INTEGER NOT NULL DEFAULT 0,
    cc_version TEXT NOT NULL DEFAULT '',
    input_tokens INTEGER NOT NULL DEFAULT 0,
    cache_write_5m_tokens INTEGER NOT NULL DEFAULT 0,
    cache_write_1h_tokens INTEGER NOT NULL DEFAULT 0,
    cache_read_tokens INTEGER NOT NULL DEFAULT 0,
    output_tokens INTEGER NOT NULL DEFAULT 0,
    UNIQUE (message_id, request_id, entry)
);

CREATE INDEX usage_project_track ON usage (project_id, track);

-- migrate:down
DROP INDEX usage_project_track;
DROP TABLE usage;
