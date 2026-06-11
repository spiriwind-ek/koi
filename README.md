# Koi — 鯉

一个基于 Lua 的高维数值计算操作系统。

在浏览器中编写 Lua 脚本，调用数学 API 进行矩阵、张量、微积分、统计等计算，数据持久化存储在 SQLite 中。

## 特性

- **Lua 脚本执行** — 沙箱化的 Lua 运行环境
- **矩阵运算** — 创建、乘法、转置、行列式、逆矩阵、SVD 等
- **张量运算** — 任意维度的张量创建与操作
- **持久化存储** — SQLite 存储，数据自动保存
- **Web 界面** — 浏览器中编写和执行代码
- **双版本架构** — 完整版（4GB+ RAM）与精简版（256MB RAM）

## 快速开始

### 依赖

- Go 1.21+
- GCC（CGO 编译 SQLite 需要）

### 构建与运行

```bash
go mod tidy
CGO_ENABLED=1 go build -o koi .
./koi                              # 启动 Web 服务器
./koi shell                        # 启动交互式 Lua Shell
```

浏览器打开 `http://localhost:8080`。

### 使用示例

```lua
-- 定义常量
c = 299792458
G = 6.67430e-11

-- 创建矩阵
math.mat_new('/data/A', 3, 3, {1,2,3, 4,5,6, 7,8,9})
math.mat_new('/data/B', 3, 3, {9,8,7, 6,5,4, 3,2,1})

-- 矩阵乘法
math.mat_mul('/data/A', '/data/B', '/data/C')

-- 打印结果
io.print(math.mat_print('/data/C'))
io.print('det = ' .. math.mat_determinant('/data/A'))
```

## 配置

编辑 `config/koi.toml`：

```toml
[server]
listen = ":8080"
timeout = "30s"

[database]
path = "./data/koi.db"

[security]
max_timeout = 60
max_memory = 1073741824
max_matrix_size = 10000

[engine]
edition = "full"

[ui]
theme = "dark"
font_size = 14
```

## API

### Lua 执行

```
POST /api/lua/execute
Body: Lua 代码
Response: {"output": "..."} 或 {"error": "..."}
```

### 文件系统

```
GET /api/filesystem/{path}
Response: 节点信息或子节点列表
```

### 设置

```
GET /api/settings
POST /api/settings/update
```

## 数学 API

| API | 说明 |
|-----|------|
| `math.mat_new(path, rows, cols, data)` | 创建矩阵 |
| `math.mat_mul(path_a, path_b, path_result)` | 矩阵乘法 |
| `math.mat_transpose(path, result)` | 转置 |
| `math.mat_determinant(path)` | 行列式 |
| `math.mat_inverse(path, result)` | 逆矩阵 |
| `math.mat_print(path)` | 打印矩阵 |
| `math.mat_shape(path)` | 获取形状 |
| `math.tensor_new(path, shape, data)` | 创建张量 |
| `math.tensor_print(path)` | 打印张量 |
| `math.tensor_shape(path)` | 获取张量形状 |
| `math.dot(a, b)` | 向量点积 |
| `math.norm(a)` | 向量范数 |
| `math.cross(a, b)` | 叉积（3D） |

## 文档

- [VISION.md](docs/VISION.md) — 长期愿景与架构目标
- [ROADMAP.md](docs/ROADMAP.md) — 分阶段实施计划
- [GATEWAY_DESIGN.md](docs/GATEWAY_DESIGN.md) — Gateway 设计
- [ENGINE_DESIGN.md](docs/ENGINE_DESIGN.md) — Engine 设计（计划中）
- [MATH_API.md](docs/MATH_API.md) — 数学 API 参考
- [SQL_CONFIG.md](docs/SQL_CONFIG.md) — 数据库与配置
- [PERFORMANCE.md](docs/PERFORMANCE.md) — 性能与优化
- [WEBUI_DESIGN.md](docs/WEBUI_DESIGN.md) — 前端设计

## 项目结构

```
koi/
├── main.go              # 入口
├── shell.go             # CLI Shell
├── config/
│   ├── config.go        # 配置加载器
│   └── koi.toml         # 默认配置
├── gateway/
│   ├── core.go          # 核心结构体、路由注册
│   ├── auth.go          # 认证中间件、CORS
│   └── handlers.go      # API 处理器
├── lua/
│   ├── vm.go            # Lua VM 池
│   ├── sandbox.go       # 沙箱环境
│   ├── api_fs.go        # 文件系统 API
│   └── api_math.go      # 数学 API
├── storage/
│   └── db.go            # SQLite 存储
├── web/
│   └── index.html       # 前端界面
├── sql/
│   └── schema.sql       # 数据库 Schema
├── docs/                # 文档
├── Makefile             # 构建脚本
├── LICENSE              # MPL-2.0
└── CONTRIBUTING.md      # 贡献指南
```

## 许可证

本项目采用 [MPL-2.0](LICENSE)（Mozilla Public License 2.0）许可证。

## 贡献

欢迎贡献！请阅读 [CONTRIBUTING.md](CONTRIBUTING.md) 了解详情。

## 致谢

- [gopher-lua](https://github.com/yuin/gopher-lua) — Go 的 Lua VM
- [gonum](https://gonum.org/) — Go 数值计算库
- [go-sqlite3](https://github.com/mattn/go-sqlite3) — SQLite 驱动
