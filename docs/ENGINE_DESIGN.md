# Koi — Engine Design (Planned)

> **Status**: Not implemented. Current math is done via gonum in the Gateway process.
> This document describes the planned Zig-based Engine for high-performance computation.

## Motivation

gonum works for moderate workloads, but has limitations:
- No SIMD optimization (OpenBLAS has AVX/NEON auto-detection)
- No LAPACK routines for large matrices (SVD, eigenvalue, QR, Cholesky)
- No GPU offloading
- Single-process bottleneck (Gateway handles both I/O and computation)

The Engine separates computation into a dedicated process with access to
battle-tested C libraries.

## Architecture

```
Engine (Zig)
├── main.zig              gRPC server entry point
├── thread_pool.zig       Worker thread pool
├── handle_manager.zig    Object lifecycle (ref counting)
├── compute_cache.zig     LRU cache for hot results
├── gpu_backend.zig       CUDA/OpenCL (optional)
├── matrix.zig            Matrix operations (OpenBLAS)
├── tensor.zig            Tensor operations
├── calculus.zig          Numerical analysis
├── stats.zig             Statistics
├── geometry.zig          Geometric operations
├── special.zig           Special functions (gamma, Bessel, etc.)
├── fft.zig               FFT operations (FFTW)
└── lib/
    ├── openblas.zig      OpenBLAS FFI bindings
    └── fftw.zig          FFTW FFI bindings
```

## Why Zig?

| Feature | Benefit |
|---------|---------|
| `@cImport` | Zero-cost FFI to OpenBLAS/FFTW, no wrapper code needed |
| Compile-time safety | Catch errors at build time, not runtime |
| No hidden allocations | Predictable memory behavior |
| Cross-compilation | Build for ARM (Lite) from x86 host |
| No runtime overhead | Comparable to C performance |

## Communication Protocol

Gateway ↔ Engine via gRPC with Protocol Buffers.

```
Gateway                          Engine
  │ gRPC request                   │
  │ (matrix data, operation)       │
  ├───────────────────────────────▶│
  │                                │ Execute with OpenBLAS
  │                                │ Cache result (LRU)
  │◀───────────────────────────────┤
  │ gRPC response                  │
  │ (result or handle ID)          │
```

### Handle Mechanism

Large objects (matrices >1MB) stay in Engine memory. Gateway receives only a
handle ID (uint64). Subsequent operations pass the handle ID instead of
serializing the full matrix.

```
mat_new("/data/big", 10000, 10000, data)
  → Gateway sends data to Engine
  → Engine stores matrix, returns handle=42
  → Gateway stores handle=42 in SQLite

mat_mul(handle=42, handle=43, "/data/result")
  → Gateway sends handles to Engine
  → Engine reads from memory (no data transfer)
  → Engine returns result handle=44
```

### Compute Cache

LRU cache for expensive operations (matrix inverse, SVD, eigenvalue).
Keyed by operation + input hash.

```
Cache hit:  O(1) lookup, return cached result
Cache miss: Compute, store in cache, evict LRU if full
```

## Library Dependencies

| Library | Purpose | Why |
|---------|---------|-----|
| OpenBLAS | BLAS/LAPACK | Gold standard for linear algebra, SIMD optimized |
| FFTW | Fast Fourier Transform | Best FFT implementation, used by MATLAB/NumPy |
| (Lite) libm | Basic math | Standard C math, no external deps |

## Lite Edition

Same Zig code, compiled as `.so` with reduced feature set:
- Linked into Gateway via CGO (no gRPC overhead)
- Matrix size limit: ≤1000×1000
- No SVD, eigenvalue, high-dimensional tensors
- No GPU backend
- Supported: basic matrix ops, 1D/2D tensors, derivatives, basic stats
