# 模块路径迁移（demo/micro/* → Jedeft/*，hades → xuanwu）

- 日期：2026-07-02
- 触发：用户将本工程 module 改为 `github.com/Jedeft/demo-micro-gw-admin`，hades 引用改为 `github.com/Jedeft/xuanwu v0.0.1`，base-user 引用改为 `github.com/Jedeft/demo-micro-base-user`（与 base-user 项目的同类改动一致）。
- 状态：**构建/静态检查全通过**，`go build` / `go vet` / `golangci-lint` 退出码 0，全工程无残留 `hades` 或旧 `github.com/demo/micro/*` 引用。

## 1. 执行命令

```bash
# 0. 批量替换 .go 文件中三条过时模块路径（见第 2 节映射）
find . -name "*.go" -not -path "./.git/*" -exec sed -i \
  's|github.com/demo/micro/gw/admin|github.com/Jedeft/demo-micro-gw-admin|g;
   s|github.com/demo/micro/base/user|github.com/Jedeft/demo-micro-base-user|g;
   s|github.com/Jedeft/hades/v3|github.com/Jedeft/xuanwu|g' {} +

# 1. import 重排（模块路径变更后本地前缀组需重排）
goimports -w .

# 2. 代码检查与格式化（CLAUDE.md 要求每次改完必做）
gofmt -l .                 # 应无输出
go vet ./...

# 3. 构建
go build ./...

# 4. 静态分析门禁
golangci-lint run ./...
```

最终结果：`gofmt -l .` 无输出；`go vet ./...` 退出 0；`go build ./...` 退出 0；`golangci-lint run ./...` 退出 0（无告警）。

## 2. 迁移映射

### 2.1 模块路径（三条，仅作用于 `.go` import 语句）

| 旧路径 | 新路径 |
|--------|--------|
| `github.com/demo/micro/gw/admin` | `github.com/Jedeft/demo-micro-gw-admin` |
| `github.com/demo/micro/base/user` | `github.com/Jedeft/demo-micro-base-user` |
| `github.com/Jedeft/hades/v3` | `github.com/Jedeft/xuanwu` |

> **hades → xuanwu 关键差异**：hades 历史版本走 `/v3` 主版本后缀（`github.com/Jedeft/hades/v3`）；xuanwu 为 `v0.0.1`（v0 主版本），模块路径为 `github.com/Jedeft/xuanwu`，**无 `/v3` 后缀**。已在 base-user 项目以同样方式迁移并验证，作为参照基准。

### 2.2 标识符与命令行 flag

| 旧 | 新 | 位置 |
|----|----|------|
| `hades.Init()` / `hades.Destroy()` | `xuanwu.Init()` / `xuanwu.Destroy()` | `internal/application.go` |
| `hades.SetConfig(...)` | `xuanwu.SetConfig(...)` | `internal/configs/config.go` |
| `cmd.FlagsHadesConfig` | `cmd.FlagsXuanwuConfig` | `cmd/cmd.go`、`internal/configs/config.go` |
| `./hades.toml`（flag 默认值） | `./xuanwu.toml` | `cmd/cmd.go` |

### 2.3 配置文件 section 前缀（**强制**，xuanwu 不兼容旧 `hades-*`）

| 旧 section | 新 section |
|------------|------------|
| `[hades-service-info]` | `[xuanwu-service-info]` |
| `[hades-supports]` | `[xuanwu-supports]` |
| `[hades-log]` | `[xuanwu-log]` |
| `[hades-tracer]` | `[xuanwu-tracer]` |
| `[[hades-cache]]` | `[[xuanwu-cache]]` |

> **依据**：xuanwu `config/config.go` 的 toml tag 固定为 `xuanwu-*` 前缀（如 `toml:"xuanwu-supports"`、`toml:"xuanwu-cache"`），且 base-user 项目的 `xuanwu.toml` 头部注释明确写明"section 前缀为 `xuanwu-*`（xuanwu 不兼容旧 `hades-*` 前缀）"。沿用旧前缀会导致框架初始化时配置项全部为零值。
>
> `cache.Config` 字段名（`alias`/`addr`/`password`/`db`/`pool_size`/`min_idle`/`read_timeout`/`write_timeout`）在 xuanwu 中未变，原 `[[hades-cache]]` 字段直接平移到 `[[xuanwu-cache]]`，仅改 section 名。

## 3. 改动明细

### 3.1 代码文件（import 路径 + 标识符）

`cmd/server/main.go`、`internal/application.go`、`internal/configs/config.go`、`internal/grpc/conn.go`、`internal/routers/router.go`、`internal/routers/jwt.go`、`internal/handlers/common.go`、`internal/handlers/error.go`、`internal/handlers/jwt.go`、`internal/handlers/auth/authrization.go`、`internal/handlers/users/user.go`、`internal/services/user_service.go`（共 12 文件）。

- import：按 2.1 三条映射替换。
- `userpb "github.com/demo/micro/base/user/api/protobuf"` → `userpb "github.com/Jedeft/demo-micro-base-user/api/protobuf"`（`user_service.go`、`handlers/users/user.go`）。
- 标识符与注释：按 2.2 替换；同步更新行内中文注释（`// hades库初始化` → `// xuanwu库初始化` 等）与 error string（`hades config set err` → `xuanwu config set err`，符合 error strings 不大写规范）。

### 3.2 配置文件

- 新增 `cmd/server/xuanwu.toml`（内容沿用原 `hades.toml`，按 2.3 改 section 前缀）。
- 删除 `cmd/server/hades.toml`。

### 3.3 辅助文件

| 文件 | 改动 |
|------|------|
| `.golangci.yml` | `goimports.local-prefixes`: `github.com/demo/micro/gw/admin` → `github.com/Jedeft/demo-micro-gw-admin`（自身包单独成组） |
| `Makefile` | `run` 目标 `-hc ./cmd/server/hades.toml` → `-hc ./cmd/server/xuanwu.toml` |
| `README.md` | 目录结构图 `hades.toml` → `xuanwu.toml` |
| `go.mod` | genproto `replace` 钉子的注释：`hades/base/user 依赖` → `xuanwu/demo-micro-base-user 依赖`（指令本身不变，见第 4 节） |

## 4. 依赖解析（环境前置，非代码改动）

迁移代码后 `go build` 报两类问题，均已解决：

| 问题 | 根因 | 处理 |
|------|------|------|
| `missing go.sum entry for github.com/Jedeft/xuanwu/*` | 模块路径变更后 `go.sum` 未同步 | 由用户 `go mod tidy` 补齐 |
| `demo-micro-base-user@v0.0.1: module declares its path as: github.com/demo/micro/base/user but was required as: github.com/Jedeft/demo-micro-base-user` | 已发布的 `v0.0.1` tag 仍声明旧模块路径（`go clean -modcache` + `GOPROXY=direct` 重下确认） | 用户重新发布 `v0.0.2`（go.mod 已声明新路径），`go.mod` require 升至 `github.com/Jedeft/demo-micro-base-user v0.0.2`，`go clean -modcache` + `go mod tidy` 后构建通过 |

### 4.1 genproto `replace` 钉子保留（**勿删**）

`go mod graph` 验证：`github.com/Jedeft/xuanwu@v0.0.1` 与 `github.com/Jedeft/demo-micro-base-user@v0.0.2` **仍依赖旧版整体式** `google.golang.org/genproto@v0.0.0-20211129164237-f09f9a12af12`（含 `googleapis/*` 包），与 `grpc-gateway`/`otel` 依赖的拆分模块 `genproto/googleapis/{api,rpc}` 冲突，导致 ambiguous import。故 `go.mod` 末尾的 `replace google.golang.org/genproto => ...@v0.0.0-20260630182238-925bb5da69e7`（拆分后版本，已移除 `googleapis/*`）**仍然必需**，已在注释中将"hades/base/user"更新为"xuanwu/demo-micro-base-user"。`replace` 不受 `go mod tidy` 自动移除，**请勿删除**。

## 5. 结论

- 三条模块路径（gw-admin / base-user / hades→xuanwu）及对应标识符、配置文件、辅助文件均已迁移完毕，全工程无 `hades` 或旧 `github.com/demo/micro/*` 残留引用（`docs/process/` 历史文档除外，保留为当时状态记录）。
- `go build ./...`、`go vet ./...`、`golangci-lint run ./...` 均退出 0。
- genproto `replace` 钉子经 `go mod graph` 复核仍为必需，保留并更新注释。
- 遗留（非本次范围）：项目无单元/集成测试文件，与 CLAUDE.md「services 层行覆盖率不低于 90%」要求存在差距，建议后续补齐。
