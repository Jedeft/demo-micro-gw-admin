# 全工程静态代码分析（golangci-lint 配置更新后）

- 日期：2026-07-01
- 触发：`.golangci.yml` 重构（新增 gosec/errorlint/gocritic/unparam/misspell/nolintlint/revive/unused，移除已停用 linter，`gocyclo.min-complexity` 15→30，`issues-exit-code` 改为默认失败即退出）
- 状态：**已修复全部 19 项告警**，`golangci-lint run ./...` 退出码 0。

## 1. 执行命令

```bash
# 0. 环境前置：依赖可解析、构建通过（详见第 2 节）
go build ./...

# 1. 全工程静态分析
golangci-lint run ./...

# 2. 代码检查与格式化（CLAUDE.md 要求每次改完必做）
gofmt -l .                # 应无输出
goimports -w -local github.com/demo/micro/gw/admin .
go vet ./...
go test ./...
```

最终结果：`gofmt -l .` 无输出；`go build ./...` 退出 0；`go vet ./...` 退出 0；`golangci-lint run ./...` 退出 0（无告警）；`go test ./...` 退出 0（项目暂无测试文件，属历史遗留，不在本次范围）。

## 2. 环境前置处理（非代码质量问题的阻塞项）

首次 `golangci-lint run ./...` 出现大量 `(typecheck)` 告警（`undefined: echo/jwt/hades`），根因是依赖不可解析，而非代码本身有问题。排查与修复过程：

| 问题 | 根因 | 处理 |
|------|------|------|
| `typecheck: undefined` 大面积报错 | 缺 `go.sum`、缺 `vendor/`、`go.mod` 中 `replace base/user` 指向他机本地路径 `/media/ut001165/...`（不存在） | 由用户将 `replace` 改为正式模块 `github.com/Jedeft/demo-micro-base-user v0.0.1` |
| `google.golang.org/genproto` ambiguous import | `hades v3.1.0` 与 `base/user v0.0.1` 均依赖旧版整体式 `genproto@v0.0.0-2021...`（含 `googleapis/*` 包），与 `grpc-gateway`/`otel` 依赖的拆分模块 `genproto/googleapis/{api,rpc}` 冲突 | 在 `go.mod` 增加 `replace google.golang.org/genproto => ...@v0.0.0-20260630182238-925bb5da69e7`（拆分后版本，已移除 `googleapis/*`）。用 `replace` 而非 require 钉子：`replace` 不受 `go mod tidy` 自动移除，更稳健 |
| `go mod tidy` 失败 | `github.com` 被防火墙拦截，`tidy` 遍历 `base/user` 的测试依赖（gonum/tempredis）时回退到 `direct` 失败 | 改用 `go mod download` + `GOFLAGS=-mod=mod go build ./...` 触发源码哈希写入 `go.sum` |
| `go.sum` 仅有 `/go.mod` 哈希、缺源码 `h1:` 哈希 | `go mod download` 默认只拉取 `go.mod`，未下载源码包 | `GOFLAGS=-mod=mod go build ./...` 自动补齐（`go.sum` 现 245 行） |

> 说明：`go.sum` 与 `vendor/` 均在 `.gitignore` 中，属本地生成物，不提交。`go.mod` 最终改动为：新增 `replace google.golang.org/genproto => ...`（消除 ambiguous import）+ 用户已设的 `replace base/user` 指向正式模块。
>
> **重要约束**：`go mod tidy` 在本环境无法运行（防火墙），且即便可运行也会移除 genproto 的 require 钉子——但因改用 `replace` 指令，`tidy` 不会移除 `replace`，故构建仍然安全。**请勿删除该 `replace`**，否则会恢复 ambiguous import。

## 3. 告警修复明细（原 19 项 → 0 项）

### 3.1 goimports —— import 分组（11 文件）

- 根因：`.golangci.yml` 原 `goimports.local-prefixes: github.com/Jedeft/hades/v3` 把框架包视作"本地前缀"，要求单独成组；常规做法应取项目模块自身路径。
- 修复：将 `local-prefixes` 改为 `github.com/demo/micro/gw/admin`（项目模块），执行 `goimports -w -local github.com/demo/micro/gw/admin .` 重排。现项目自身包单独成组（最后一组），`hades`/echo 等归入第三方组。
- 涉及：`cmd/server/main.go`、`internal/application.go`、`internal/configs/config.go`、`internal/grpc/conn.go`、`internal/handlers/common.go`、`internal/handlers/error.go`、`internal/handlers/jwt.go`、`internal/handlers/auth/authrization.go`、`internal/handlers/users/user.go`、`internal/routers/router.go`、`internal/services/user_service.go`。

### 3.2 staticcheck SA1019 —— 已弃用 API（3 项）

| 位置 | 告警 | 修复 |
|------|------|------|
| `internal/grpc/conn.go:43` | `grpc.DialContext` 已弃用 | 改用 `grpc.NewClient`（非阻塞） |
| `internal/grpc/conn.go:45` | `grpc.WithBlock` 已弃用 | `NewClient` 不支持 `WithBlock`；改用 `conn.Connect()` + `GetState()`/`WaitForStateChange` 在 5s 超时内轮询连接状态，复刻原"服务不可达即快速失败"行为 |
| `internal/routers/router.go:30` | `middleware.Logger()` 已弃用 | 改用 `middleware.RequestLoggerWithConfig` + `LogValuesFunc`，通过 `c.Logger().Infoj/Errorj` 输出结构化 JSON 访问日志（与原 echo logger 输出一致） |

> `NewClient` 迁移行为说明：原 `DialContext+WithBlock` 在 5s 内阻塞至 `Ready` 或失败；新实现主动 `Connect()` 后轮询 `GetState()`，超时未达 `Ready` 则 `conn.Close()` 并返回错误。语义等价。

### 3.3 gosec G115 —— int→uint32 整型溢出（3 项）

| 位置 | 修复 |
|------|------|
| `internal/handlers/error.go:35` | `uint32(httpErr.Code)`：HTTP 状态码恒在 uint32 范围内，加 `//nolint:gosec` 并注明原因（G115 流不敏感，无法识别安全转换） |
| `internal/handlers/users/user.go:175` | `uint32(req.Limit)`：新增 `defaultUserListLimit=20`/`maxUserListLimit=100` 常量，在转换前 clamp 至 `[1,100]`（既防溢出又防资源耗尽），并加 nolint 注明 |
| `internal/handlers/users/user.go:176` | `uint32(req.Offset)`：转换前 floor 至 `>=0`（消除负数→大 uint32 的危险翻转），并加 nolint 注明 |

### 3.4 revive —— 未使用参数（1 项）

- `internal/handlers/jwt.go:35`：`InvalidJWT(err, c)` 两个参数均未使用（注释说明一律返回 401）。改名为 `InvalidJWT(_ error, _ echo.Context)`，签名不变（不影响调用方）。

### 3.5 errorlint —— error 类型断言（1 项）

- `internal/handlers/error.go:24`：`switch err := err.(type)` 对 wrapped error 失败。改为 `errors.As(err, &bizErr)` / `errors.As(err, &httpErr)`，正确处理被包装的 `*xerr.Err` 与 `*echo.HTTPError`。

## 4. 结论

- **19 项告警全部清零**，`golangci-lint run ./...` 退出码 0。
- 新增 linter 未发现高危问题（无硬编码凭证、无 `panic` 滥用、无错误链断裂等）。
- 新配置的门禁生效：`issues-exit-code` 默认为 1，CI 现可真正拦截告警。
- `go.mod` 因外部模块（hades/base-user）依赖旧版整体式 genproto，需保留 `replace` 钉子；已在 go.mod 内附注释说明，请勿删除或运行 `go mod tidy`。
- 遗留（非本次范围）：项目无单元/集成测试文件，与 CLAUDE.md「services 层行覆盖率不低于 90%」要求存在差距，建议后续补齐。
