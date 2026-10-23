-- migrate:up
CREATE TABLE model_prices (
    model TEXT NOT NULL,
    effective_from TEXT NOT NULL,
    input REAL,
    cache_write_5m REAL,
    cache_write_1h REAL,
    cache_read REAL,
    output REAL,
    PRIMARY KEY (model, effective_from)
);

INSERT INTO model_prices (model, effective_from, input, cache_write_5m, cache_write_1h, cache_read, output) VALUES
    ('claude-opus-5-5', '2026-01-01', 4, 5, 8, 0.20, 20),
    ('claude-sonnet-5-5', '2026-01-01', 2, 2.5, 4, 0.10, 10),
    ('claude-sonnet-5', '2026-01-01', 2, 2.5, 4, 0.20, 10),
    ('claude-haiku-5-5', '2026-01-01', 0.10, 0.125, 0.20, 0.01, 0.50),
    ('claude-haiku-4-5-20251001', '2026-01-01', 1, 1.25, 2, 0.10, 5);

-- migrate:down
DROP TABLE model_prices;
