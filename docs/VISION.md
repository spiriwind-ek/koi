# Koi — Long-Term Vision

> **Status**: This document describes the long-term architectural vision for Koi.
> The current implementation (v0.1.0-mvp) covers Phase 1 only. See [ROADMAP.md](ROADMAP.md)
> for the phased implementation plan.

## What Koi Is

Koi is a Lua-based numerical computation operating system. Users write Lua scripts
in a browser, call math APIs for matrix/tensor/statistics operations, and persist
data in a virtual filesystem backed by SQLite.

**Name**: Koi (鯉) — A koi fish grows from a small pond to a vast ocean.

## Current Implementation (v0.1.0-mvp)

```
Browser (HTML/JS)
    │ HTTP
    ▼
Go Gateway
    ├── HTTP Server (net/http)
    ├── Lua VM (gopher-lua, sandboxed)
    ├── Math APIs (gonum: matrix, tensor, vector)
    ├── Virtual Filesystem (SQLite)
    └── API Key Authentication
```

Single binary, single process, no external dependencies beyond SQLite.

## Long-Term Architecture

The goal is to evolve Koi into a dual-backend system:

```
Browser (HTML/JS)
    │ HTTP / WebSocket
    ▼
Gateway (Go)
    ├── HTTP/WebSocket Server
    ├── Lua VM Pool (gopher-lua)
    ├── LuaOS API Layer
    ├── SQL Storage
    ├── Task Queue (async)
    └── gRPC Client ──▶ Engine (Zig)
                            ├── Matrix / Tensor (OpenBLAS)
                            ├── Calculus
                            ├── Statistics
                            ├── Geometry
                            ├── Special Functions
                            ├── FFT (FFTW)
                            ├── Handle Manager
                            ├── Compute Cache (LRU)
                            └── GPU Backend (optional)
```

### Why Two Languages?

| Component | Language | Reason |
|-----------|----------|--------|
| Gateway | Go | Fast compilation, good concurrency, mature ecosystem for HTTP/gRPC |
| Engine | Zig | Zero-cost C interop, compile-time safety, direct OpenBLAS/FFTW calls |
| Web UI | Vanilla JS | Zero framework, minimal bundle size |

### Why Not Pure Go for Math?

gonum is good for moderate workloads, but:
- No SIMD optimization (OpenBLAS has AVX/NEON)
- No LAPACK routines (SVD, eigenvalue decomposition for large matrices)
- No GPU offloading

The Zig Engine wraps battle-tested C libraries (OpenBLAS, FFTW) with zero FFI
overhead via `@cImport`.

## LuaOS API

The Lua environment provides OS-like APIs:

| Module | Description | MVP Status |
|--------|-------------|------------|
| `fs` | Virtual filesystem (mkdir, read, write, ls, rm, exists) | ✅ Implemented |
| `io` | Input/output (print) | ✅ Implemented |
| `os` | System info (time, clock, version, edition) | ✅ Implemented |
| `proc` | Task management (start, wait, status, kill) | ❌ Planned |
| `settings` | Runtime configuration (get, set, preset) | ❌ Planned |
| `matrix` | Linear algebra | ✅ Partial (gonum) |
| `tensor` | N-dimensional arrays | ✅ Partial (gonum) |
| `calculus` | Numerical analysis | ❌ Planned |
| `stats` | Statistics | ❌ Planned |
| `geometry` | Geometric operations | ❌ Planned |
| `special` | Special functions | ❌ Planned |
| `fft` | Fast Fourier Transform | ❌ Planned |

## Dual Editions (Planned)

### Full Edition

For machines with sufficient resources (4GB+ RAM, modern CPU).

- Full Zig Engine with all math modules
- OpenBLAS for linear algebra
- FFTW for Fourier transforms
- GPU acceleration (CUDA/OpenCL, optional)

### Lite Edition

For resource-constrained devices (256MB RAM, 800MHz dual-core).

- Same Gateway codebase (build tag or runtime flag)
- Embedded simplified math engine (Zig compiled as .so)
- Same LuaOS API surface, unsupported ops return `nil, "unsupported"`
- Matrix size limit: ≤1000×1000
- No SVD, eigenvalue, high-dimensional tensors, ODE solvers

```lua
-- Lua code can check at runtime:
if os.edition() == "lite" then
    -- use simplified approach
else
    -- use full computation
end
```

## Storage Model

SQLite. S3-style flat key-value. One table for everything.

```sql
CREATE TABLE nodes (
    key         TEXT PRIMARY KEY,    -- "/math/matrix/A"
    parent_key  TEXT,                -- parent node key
    name        TEXT NOT NULL,       -- display name
    obj_type    TEXT NOT NULL,       -- dir, matrix, tensor, number, string, handle
    value       TEXT,                -- JSON for small objects (<1KB)
    blob_data   BLOB,               -- compressed binary for large objects
    blob_meta   TEXT,                -- JSON metadata
    created_at  INTEGER,
    updated_at  INTEGER
);
```

### Rules

1. All nodes equal — dirs and files in one table
2. Auto mkdir-p — write "/a/b/c" auto-creates missing dirs
3. Lua object protocol — write: serialize to JSON; read: deserialize with metatable
4. Atomicity — mkdir + write wrapped in SQLite transaction
