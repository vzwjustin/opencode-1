-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS coding_threads (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    raw_goal TEXT NOT NULL,
    narrowed_goal TEXT NOT NULL,
    thread_type TEXT NOT NULL DEFAULT 'feature',
    status TEXT NOT NULL DEFAULT 'active',
    next_step TEXT,
    next_step_reason TEXT,
    confidence REAL NOT NULL DEFAULT 0.5,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    FOREIGN KEY (session_id) REFERENCES sessions (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_coding_threads_session ON coding_threads (session_id);

CREATE TABLE IF NOT EXISTS checkpoints (
    id TEXT PRIMARY KEY,
    thread_id TEXT NOT NULL,
    summary TEXT NOT NULL,
    next_step TEXT,
    created_at INTEGER NOT NULL,
    FOREIGN KEY (thread_id) REFERENCES coding_threads (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS side_quests (
    id TEXT PRIMARY KEY,
    thread_id TEXT NOT NULL,
    description TEXT NOT NULL,
    resumed INTEGER NOT NULL DEFAULT 0,
    created_at INTEGER NOT NULL,
    FOREIGN KEY (thread_id) REFERENCES coding_threads (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS drift_events (
    id TEXT PRIMARY KEY,
    thread_id TEXT NOT NULL,
    signal TEXT NOT NULL,
    description TEXT NOT NULL,
    return_point TEXT,
    created_at INTEGER NOT NULL,
    FOREIGN KEY (thread_id) REFERENCES coding_threads (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS notes (
    id TEXT PRIMARY KEY,
    thread_id TEXT NOT NULL,
    text TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    FOREIGN KEY (thread_id) REFERENCES coding_threads (id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS notes;
DROP TABLE IF EXISTS drift_events;
DROP TABLE IF EXISTS side_quests;
DROP TABLE IF EXISTS checkpoints;
DROP TABLE IF EXISTS coding_threads;
-- +goose StatementEnd
