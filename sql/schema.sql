-- Koi Storage Schema

PRAGMA journal_mode = WAL;
PRAGMA synchronous = NORMAL;
PRAGMA cache_size = 10000;
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS nodes (
    key         TEXT PRIMARY KEY,
    parent_key  TEXT,
    name        TEXT NOT NULL,
    obj_type    TEXT NOT NULL,
    value       TEXT,
    blob_data   BLOB,
    blob_meta   TEXT,
    created_at  INTEGER NOT NULL,
    updated_at  INTEGER NOT NULL,
    FOREIGN KEY (parent_key) REFERENCES nodes(key) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_nodes_parent ON nodes(parent_key);
CREATE INDEX IF NOT EXISTS idx_nodes_type ON nodes(obj_type);

CREATE TABLE IF NOT EXISTS settings (
    key         TEXT PRIMARY KEY,
    value       TEXT NOT NULL,
    category    TEXT NOT NULL,
    description TEXT,
    updated_at  INTEGER NOT NULL
);

-- Seed initial directories
INSERT OR IGNORE INTO nodes (key, parent_key, name, obj_type, created_at, updated_at) VALUES
('/', NULL, '/', 'dir', strftime('%s', 'now'), strftime('%s', 'now')),
('/home', '/', 'home', 'dir', strftime('%s', 'now'), strftime('%s', 'now')),
('/home/user', '/home', 'user', 'dir', strftime('%s', 'now'), strftime('%s', 'now')),
('/data', '/', 'data', 'dir', strftime('%s', 'now'), strftime('%s', 'now')),
('/tmp', '/', 'tmp', 'dir', strftime('%s', 'now'), strftime('%s', 'now'));
