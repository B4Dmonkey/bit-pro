-- name: UpsertUsage :exec
INSERT INTO usage (
    project_id, session_id, agent_id, agent_type, parent_agent_id, invoked_agent,
    skill, skill_invocation, track, bar,
    message_id, request_id, entry, ts, model, speed, is_sidechain, cc_version,
    input_tokens, cache_write_5m_tokens, cache_write_1h_tokens, cache_read_tokens, output_tokens
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT (message_id, request_id, entry) DO UPDATE SET
    project_id = excluded.project_id,
    session_id = excluded.session_id,
    agent_id = excluded.agent_id,
    agent_type = excluded.agent_type,
    parent_agent_id = excluded.parent_agent_id,
    invoked_agent = excluded.invoked_agent,
    skill = excluded.skill,
    skill_invocation = excluded.skill_invocation,
    track = excluded.track,
    bar = excluded.bar,
    ts = excluded.ts,
    model = excluded.model,
    speed = excluded.speed,
    is_sidechain = excluded.is_sidechain,
    cc_version = excluded.cc_version,
    input_tokens = excluded.input_tokens,
    cache_write_5m_tokens = excluded.cache_write_5m_tokens,
    cache_write_1h_tokens = excluded.cache_write_1h_tokens,
    cache_read_tokens = excluded.cache_read_tokens,
    output_tokens = excluded.output_tokens;
