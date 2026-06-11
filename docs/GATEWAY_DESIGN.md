# Koi — Gateway Design

> Current: Go HTTP server with Lua VM and SQLite.
> Future: Add WebSocket, task queue, gRPC client to Zig Engine.

## Current Architecture

```
Gateway (Go)
├── main.go              Config loading, server startup
├── config/              TOML config with defaults
├── gateway/server.go    HTTP routes, auth, CORS
├── lua/
│   ├── vm.go            VM pool (Get/Put/Discard/Close)
│   ├── sandbox.go       Sandbox setup, API registration
│   ├── api_fs.go        Virtual filesystem API
│   └── api_math.go      Math API (gonum)
├── storage/db.go        SQLite CRUD + AutoMkdir
└── web/index.html       Single-page frontend
```

## API Endpoints

| Endpoint | Method | Auth | Description |
|----------|--------|------|-------------|
| `/api/health` | GET | No | Health check |
| `/api/lua/execute` | POST | Yes | Execute Lua code |
| `/api/filesystem/{path}` | GET | Yes | Read filesystem |
| `/api/settings` | GET | Yes | Read settings (safe subset) |
| `/api/settings/update` | POST | Yes | Update settings |

## Authentication

API key via `Authorization: Bearer <key>` header or `?api_key=<key>` query parameter.
Configured via `-api-key` flag or `KOI_API_KEY` environment variable.

## Lua Execution Flow

```
Request → Auth check → Get VM from pool
    → Override io.print to capture output
    → L.DoString(code) in goroutine
    → Wait for completion or timeout
    → On success: return VM to pool
    → On timeout/error: discard VM (close)
    → Return captured output as JSON
```

## Future: WebSocket

```
/ws → Upgrade to WebSocket
    → Bidirectional communication
    → Real-time console output
    → Live filesystem updates
    → Session management
```

## Future: Task Queue

```
POST /api/tasks/submit → Queue task
GET  /api/tasks/{id}   → Check status
POST /api/tasks/{id}/cancel → Cancel task

Queue features:
- Priority levels
- Timeout per task
- Persistence (survive restart)
- Progress callbacks
```
