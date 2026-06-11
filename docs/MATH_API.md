# Koi — Math API Reference

> Current implementation uses gonum. Future versions may use Zig Engine for
> high-performance operations.

## Matrix Operations

| API | Description | MVP |
|-----|-------------|-----|
| `math.mat_new(path, rows, cols, data)` | Create matrix from flat array | ✅ |
| `math.mat_mul(path_a, path_b, path_result)` | Matrix multiplication | ✅ |
| `math.mat_transpose(path, result)` | Transpose | ✅ |
| `math.mat_determinant(path)` | Determinant (square only) | ✅ |
| `math.mat_inverse(path, result)` | Inverse (square only) | ✅ |
| `math.mat_print(path)` | String representation | ✅ |
| `math.mat_shape(path)` | Returns {rows, cols} | ✅ |
| `math.mat_svd(path)` | SVD decomposition | ❌ |
| `math.mat_eigenvalues(path)` | Eigenvalue decomposition | ❌ |
| `math.mat_lu(path)` | LU decomposition | ❌ |
| `math.mat_qr(path)` | QR decomposition | ❌ |
| `math.mat_cholesky(path)` | Cholesky decomposition | ❌ |
| `math.mat_rank(path)` | Matrix rank | ❌ |
| `math.mat_norm(path)` | Frobenius norm | ❌ |
| `math.mat_trace(path)` | Trace | ❌ |

## Tensor Operations

| API | Description | MVP |
|-----|-------------|-----|
| `math.tensor_new(path, shape, data)` | Create N-dimensional tensor | ✅ |
| `math.tensor_print(path)` | String representation | ✅ |
| `math.tensor_shape(path)` | Returns shape table | ✅ |
| `math.tensor_reshape(path, new_shape)` | Reshape | ❌ |
| `math.tensor_contract(a, b, axes)` | Tensor contraction | ❌ |
| `math.tensor_reduce(path, axis, op)` | Reduction | ❌ |

## Vector Operations

| API | Description | MVP |
|-----|-------------|-----|
| `math.dot(a, b)` | Dot product | ✅ |
| `math.norm(a)` | L2 norm | ✅ |
| `math.cross(a, b)` | Cross product (3D) | ✅ |
| `math.normalize(a)` | Unit vector | ❌ |
| `math.angle(a, b)` | Angle between vectors | ❌ |
| `math.project(a, b)` | Projection | ❌ |

## Calculus (Planned)

| API | Description |
|-----|-------------|
| `calculus.derivative(f, x, h)` | Numerical derivative |
| `calculus.gradient(f, point)` | Gradient |
| `calculus.integral(f, a, b)` | 1D integral |
| `calculus.root(f, x0, tol)` | Root finding |
| `calculus.ode_rk4(f, y0, t_range)` | ODE solver (RK4) |

## Statistics (Planned)

| API | Description |
|-----|-------------|
| `stats.mean(data)` | Mean |
| `stats.variance(data)` | Variance |
| `stats.stddev(data)` | Standard deviation |
| `stats.correlation(x, y)` | Pearson correlation |
| `stats.linreg(x, y)` | Linear regression |
| `stats.normal.pdf(x, mu, sigma)` | Normal PDF |

## FFT (Planned)

| API | Description |
|-----|-------------|
| `fft.forward(data)` | Forward FFT |
| `fft.inverse(data)` | Inverse FFT |
| `fft.powerspectrum(data)` | Power spectrum |

## Security Limits

All matrix/tensor creation functions enforce configurable limits:

```toml
[security]
max_matrix_size = 10000    # Max rows or columns
max_tensor_ndim = 8        # Max tensor dimensions
```

Operations exceeding these limits return `nil, "exceeds limit"`.
