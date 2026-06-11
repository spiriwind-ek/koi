# 贡献指南

感谢您对 Koi 项目的兴趣！

## 如何贡献

### 报告问题

- 使用 GitHub Issues 报告 bug
- 请包含复现步骤、期望行为和实际行为
- 如果可能，请提供错误日志

### 提交代码

1. Fork 本仓库
2. 创建您的特性分支：`git checkout -b feature/my-feature`
3. 提交您的修改：`git commit -m 'Add my feature'`
4. 推送到分支：`git push origin feature/my-feature`
5. 创建 Pull Request

### 代码规范

- Go 代码请遵循 [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- 使用 `gofmt` 格式化代码
- 使用 `go vet` 检查代码
- 为新功能添加测试

### 提交信息规范

使用简洁的提交信息：

```
类型: 简短描述

详细描述（可选）
```

类型包括：
- `feat`: 新功能
- `fix`: 修复 bug
- `docs`: 文档
- `style`: 代码格式（不影响功能）
- `refactor`: 重构
- `test`: 测试
- `chore`: 构建/工具

### 开发环境

```bash
# 克隆仓库
git clone https://github.com/YOUR_USERNAME/koi.git
cd koi

# 安装依赖
go mod tidy

# 构建
CGO_ENABLED=1 go build -o koi .

# 运行测试
go test ./...
```

### 项目结构

```
koi/
├── main.go              # 入口
├── config/              # 配置
├── gateway/             # HTTP 服务器
├── lua/                 # Lua VM 和 API
├── storage/             # 数据存储
├── web/                 # 前端
├── sql/                 # 数据库 Schema
└── docs/                # 文档
```

### 添加新的数学 API

1. 在 `lua/api_math.go` 中添加函数
2. 在 `lua/sandbox.go` 中注册到 Lua 环境
3. 在 `README.md` 中添加文档
4. 添加测试

示例：

```go
// math.my_function(param) -> result
func mathMyFunction(L *lua.LState) int {
    param := L.CheckNumber(1)
    // 计算逻辑
    result := float64(param) * 2
    L.Push(lua.LNumber(result))
    return 1
}
```

在 `sandbox.go` 中注册：

```go
L.SetField(mathTable, "my_function", L.NewFunction(mathMyFunction))
```

## 行为准则

- 尊重所有参与者
- 接受建设性批评
- 专注于对社区最有利的事情
- 对他人表示同理心

## 许可证

贡献的代码将在项目的双许可证（MPL-2.0 / MulanPSL-2.0）下发布。
