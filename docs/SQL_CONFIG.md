# Koi — SQL & Configuration

## Database Schema

SQLite with WAL mode for concurrent read performance.

### nodes table

Stores all filesystem objects (directories, matrices, tensors, etc.).

```sql
CREATE TABLE nodes (
    key         TEXT PRIMARY KEY,    -- Unique path: "/data/matrix/A"
    parent_key  TEXT,                -- Parent path (NULL for root)
    name        TEXT NOT NULL,       -- Display name
    obj_type    TEXT NOT NULL,       -- dir, matrix, tensor, number, string, blob
    value       TEXT,                -- JSON for small objects (<1KB)
    blob_data   BLOB,               -- Compressed binary for large objects
    blob_meta   TEXT,                -- JSON metadata (format, shape, compression)
    created_at  INTEGER NOT NULL,
    updated_at  INTEGER NOT NULL,
    FOREIGN KEY (parent_key) REFERENCES nodes(key) ON DELETE CASCADE
);
```

### Indexes

```sql
CREATE INDEX idx_nodes_parent ON nodes(parent_key);
CREATE INDEX idx_nodes_type ON nodes(obj_type);
```

### Concurrency

```sql
PRAGMA journal_mode = WAL;
PRAGMA synchronous = NORMAL;
PRAGMA cache_size = 10000;
PRAGMA foreign_keys = ON;
```

## Configuration

TOML file at `config/koi.toml`.

```toml
[server]
listen = ":8080"
timeout = "30s"

[database]
path = "./data/koi.db"

[security]
max_timeout = 60           # Max Lua execution timeout (seconds)
max_memory = 1073741824    # Max memory (bytes)
max_matrix_size = 10000    # Max matrix dimension
max_tensor_ndim = 8        # Max tensor dimensions

[engine]
edition = "full"           # "full" or "lite" (lite not yet implemented)

[ui]
theme = "dark"
font_size = 14
tab_size = 4
```

### Command-Line Overrides

```
./koi -config config/koi.toml -listen :9090 -db ./data/koi.db -api-key secret
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `KOI_API_KEY` | API key for authentication |

### Runtime Updates

Settings can be updated via the `/api/settings/update` endpoint (UI only,
not database path or listen address for security).
