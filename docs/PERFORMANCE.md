# Koi — Performance

> Current: gonum-based computation in Go.
> Future: Zig Engine with OpenBLAS/FFTW for higher performance.

## Current Performance (gonum)

| Operation | Size | Typical Time |
|-----------|------|--------------|
| Matrix multiply | 100×100 | ~1ms |
| Matrix multiply | 1000×1000 | ~500ms |
| Determinant | 1000×1000 | ~200ms |
| SVD | Not implemented | — |

gonum uses pure Go with some assembly optimizations. Adequate for most use cases
but limited for very large matrices or high-throughput workloads.

## Planned Performance (Zig + OpenBLAS)

| Operation | Size | Target Time |
|-----------|------|-------------|
| Matrix multiply | 1000×1000 | ~10ms |
| SVD | 1000×1000 | ~50ms |
| FFT | 1M points | ~50ms |

OpenBLAS provides:
- SIMD auto-detection (SSE, AVX, NEON)
- Multi-threaded BLAS routines
- Optimized LAPACK implementations

## Optimization Strategies

### Storage

- WAL mode for concurrent reads
- Connection pooling
- Large objects compressed with zstd
- Indexes on frequently queried columns

### Computation (Planned)

- Handle mechanism: large objects in Engine memory, pass only IDs
- LRU cache: avoid recomputing expensive operations
- Thread pool: concurrent request handling
- GPU offloading: CUDA/OpenCL for massive parallelism

### Network (Planned)

- gRPC binary protocol (60-80% less overhead than JSON)
- Streaming for large matrix transfers
- WebSocket for real-time output

## Resource Limits

Configurable via `config/koi.toml`:

| Limit | Default | Description |
|-------|---------|-------------|
| `max_timeout` | 60s | Max Lua execution time |
| `max_memory` | 1GB | Max memory usage |
| `max_matrix_size` | 10000 | Max matrix dimension |
| `max_tensor_ndim` | 8 | Max tensor dimensions |
