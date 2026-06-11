# Koi — Roadmap

> Phased implementation plan. Each phase builds on the previous one.

## Phase 1: MVP ✅ (Current)

**Goal**: Working system with basic math operations.

| Component | Status | Details |
|-----------|--------|---------|
| Go Gateway | ✅ | HTTP server, API key auth, CORS |
| Lua VM | ✅ | gopher-lua, sandboxed, VM pool |
| Matrix API | ✅ | gonum: new, mul, transpose, determinant, inverse, print, shape |
| Tensor API | ✅ | gonum: new, print, shape (basic) |
| Vector API | ✅ | dot, norm, cross |
| Filesystem API | ✅ | mkdir, read, write, ls, rm, exists |
| SQLite Storage | ✅ | nodes table, WAL mode, indexes |
| Web UI | ✅ | Code editor, file tree, console, settings |
| Config | ✅ | TOML file, runtime updates |
| Security | ✅ | API key, size limits, timeout |

## Phase 2: Enhanced Math (Planned)

**Goal**: Expand mathematical capabilities without changing architecture.

| Feature | Description |
|---------|-------------|
| Matrix decomposition | SVD, LU, QR, Cholesky (via gonum/lapack) |
| Eigenvalues | Eigenvalue/eigenvector computation |
| Statistics | mean, variance, stddev, correlation, regression |
| Calculus | Numerical derivative, integral, root finding |
| FFT | Forward/inverse FFT (via gonum/dft) |
| Geometry | Distance metrics, projections, convex hull |

## Phase 3: Zig Engine (Planned)

**Goal**: High-performance math backend for large-scale computation.

| Feature | Description |
|---------|-------------|
| Zig Engine binary | gRPC server wrapping OpenBLAS/FFTW |
| gRPC protocol | Binary serialization, streaming for large matrices |
| Handle Manager | Large objects in Engine memory, Gateway passes handle IDs |
| Compute Cache | LRU cache for hot results (matrix inverse, SVD) |
| Thread Pool | Concurrent request handling in Engine |

## Phase 4: Advanced Features (Planned)

**Goal**: Production-ready system with advanced capabilities.

| Feature | Description |
|---------|-------------|
| WebSocket | Real-time console output, live file updates |
| Task Queue | Async computation with priority, timeout, cancellation |
| Lite Edition | Embedded math for resource-constrained devices |
| GPU Backend | CUDA/OpenCL for large matrix/tensor operations |
| Monaco Editor | Professional code editing with Lua syntax support |
| Data Visualization | Charts for matrix data, function plots |

## Phase 5: Ecosystem (Future)

**Goal**: Community adoption and extensibility.

| Feature | Description |
|---------|-------------|
| Plugin System | Load custom Lua modules and C extensions |
| Data Import/Export | NumPy, MATLAB, CSV formats |
| Docker/K8s | Containerized deployment |
| Multi-tenant | User isolation and resource quotas |
| Monitoring | Prometheus metrics, Grafana dashboards |

## Dependencies Between Phases

```
Phase 1 (MVP)          ← Done
    ↓
Phase 2 (Math)         ← Can start now
    ↓
Phase 3 (Zig Engine)   ← Needs Phase 2 for API design
    ↓
Phase 4 (Advanced)     ← Needs Phase 3 for performance
    ↓
Phase 5 (Ecosystem)    ← Needs stable core
```

Phase 2 can be done entirely in Go (gonum). Phase 3 introduces Zig.
