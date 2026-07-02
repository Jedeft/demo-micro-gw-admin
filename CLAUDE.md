# Golang 网关微服务

## 1. 核心命令 (Essential Commands)
- **运行所有测试**: `go test ./... -v`
- **运行集成测试（需要数据库等）**: `go test -tags=integration ./...`
- **代码检查与格式化**: `golangci-lint run` 和 `go fmt ./...`
- **构建服务**: `make build`

## 2. 项目架构 (Architecture)
本项目遵循整洁架构（Clean Architecture）原则，目录结构如下：
- `cmd/`: 
    - `cmd.go`: 命令入口，当前文件存放服务启动时接收参数相关代码。
    - `server/`: 各微服务的入口 (`main.go`)，微服务.toml配置项默认也放在此目录下 。
- `internal/`: 核心业务代码，**不应被外部项目导入**。
    - `configs/`: 配置项，微服务的所有配置数据结构存储位置。
    - `handlers/`: 业务接口入口，HTTP接口实现处，主要职责为实现接口参数校验。
    - `services/`: 核心业务逻辑处理层，如果一个接口存在有多个业务微服务的rpc调用，那么将在这一层做处理
    - `supports/`:
        - `utils/`: 当前服务所使用的工具类。
        - `ierr/`: 服务内部定义的错误。
    - `application.go`: 服务资源处理文件，**在这里实现优雅关闭相关逻辑**。
        * 服务启动参数获取
        * 服务资源初始化
        * 服务启动
        * 服务资源销毁
- `deployments/`: 构建所依赖的相关文件。
- `scripts/`: 项目配套脚本存储位置。

**架构约束 (Architecture Constraints)**:
- **依赖方向**: `handlers` -> `services` -> `下游业务服务rpc调用`。**严禁** `services` 依赖 `handlers` 的具体实现。
- **`internal` 包隔离**: 新代码必须放在正确的 `internal/` 子包中。
- **上下文传递**: 所有涉及 I/O 的函数（如数据库、HTTP 请求）的第一个参数必须是 `context.Context`。
- **数据库**: 在网关服务中禁止直接对db进行操作，所有的数据操作都依赖于下游业务服务实现。

## 3. HTTP API 定义规范 (JSON Buffers)

本项目所有 API 接口统一以 http 为基础，通过 json 进行协议定义。

> **核心规范文档**：@docs/api-rps-http-design-standards.md

## 4. 代码规范 (Coding Standards)

### 4.1 通用规范引用

本项目遵循 Go 社区通用规范，核心规范文档详见：

> **@docs/go-coding-standards.md**

**优先级**：Code Review Comments > Effective Go > Google Go Style Guide

补充参考：
- [Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- [Effective Go](https://go.dev/doc/effective_go)
- [Google Go Style Guide](https://google.github.io/styleguide/go/decisions)

> **自动化强制**：所有代码提交前必须通过 `gofmt` 或 `goimports` 格式化，机械性风格问题由工具自动修复，人工 Review 不再讨论。

---

### 4.2 核心规则速查（Code Review Comments 硬性规则）

| # | 规则 | 说明 |
|---|------|------|
| 1 | **格式化** | 所有代码必须通过 `gofmt` / `goimports` 格式化 |
| 2 | **注释** | 所有导出元素必须有 doc comment，完整句子，以英文句号结尾 |
| 3 | **Context** | 作为第一个参数传递，命名为 `ctx`，**不存入结构体** |
| 4 | **空 slice** | 优先 `var s []T` 而非 `s := []T{}` |
| 5 | **Error strings** | 首字母不大写，末尾不加标点 |
| 6 | **错误检查** | 每个 error 必须检查或用 `_` 明确忽略 |
| 7 | **正常路径** | 不缩进（indent error flow），错误路径尽早 return |
| 8 | **Goroutine** | 必须明确生命周期管理，确保可退出 |
| 9 | **Receiver 一致性** | 不要混用值/指针 receiver |
| 10 | **测试** | 用表驱动测试 + `t.Helper()` + 有用的失败信息 |

### 4.3 强烈推荐（Effective Go + Google Style Guide）

| # | 规则 | 说明 |
|---|------|------|
| 11 | **包命名** | 简短、小写、单单词，避免 `util` / `common` |
| 12 | **接口** | 偏好小接口（1-3 方法） |
| 13 | **错误包装** | 使用 `fmt.Errorf("...: %w", err)` 保留错误链 |
| 14 | **Mutex 零值** | 直接使用，不需要显式初始化 |
| 15 | **defer** | 紧跟在资源获取之后使用 |
| 16 | **命名缩写** | 全大写或全小写（`HTTP` 而非 `Http`） |
| 17 | **函数短小** | 一个函数只做一件事 |
| 18 | **表驱动测试** | 结构体字段包含 `name` / `input` / `want` |

---

### 4.4 项目特定约定

以下规则为本项目强制执行，优先级高于社区通用规范：

#### 配置管理
- **必须**使用 `config.toml` 文件管理所有配置项
- **禁止**在代码中硬编码任何配置值（端口、地址、超时、凭证等）
- 配置结构体定义与 toml 文件结构保持一一对应

#### 业务服务层 (Services)
- **必须**在 `services` 层完成所有下游业务服务调用
- **禁止**在 `handlers` 直接调用下游业务服务

#### 测试规范
| 维度 | 要求 |
|------|------|
| 文件命名 | `*_test.go`，与被测文件同包 |
| 断言库 | 统一使用 `testify/assert`（非致命断言）和 `testify/require`（致命断言） |
| Mock 生成 | 外部依赖（HTTP 客户端、缓存等）使用 `gomock` 生成 Mock 对象 |
| 覆盖率 | 核心业务逻辑（`services` 层）行覆盖率不低于 **90%** |

## 5. 红线规则 (Rules & Gotchas)
- **绝对禁止 (NEVER)**:
    - 修改 `internal/handlers/` 下的核心实体和接口，除非有明确指令。
    - 提交包含密码、API 密钥等敏感信息的代码。
    - 在 `go.mod` 中引入非必要的、未经评估的第三方库。
    - 在业务代码中使用 `panic`，除非是在 `init` 函数中处理不可恢复的错误。

- **必须执行 (ALWAYS)**:
    - 新增或修改 `services` 接口后，**必须先**更新对应的 Mock 实现 (`go generate ./...`)。
    - 所有新增的 API 端点，必须同时提供单元测试和集成测试。
    - 在修改公共函数或方法时，**必须**更新其对应的 Go Doc 注释。
    - 每一次的任务执行命令，以及汇总执行结果都需要放到docs/process下。
    - 每次修改完代码都要运行代码检查与格式化，确认当前编写的代码是否满足静态编译要求。
    - 当前项目目录结构如果发生了变更（增删改）那么必须同步更新README.md

- **特别注意 (Gotchas)**:
    - 项目的分布式事务管理是通过 `services` 层实现的。
    - 时间处理统一使用unix时间戳。

## Session Context Recovery (IMPORTANT)

**FIRST ACTION on every new conversation**: Before responding to the user, silently read these two files to recover context from previous sessions:
1. `.checkpoints/project-state.md` — current project state
2. The most recent file in `.checkpoints/sessions/` — latest session summary

Then briefly tell the user what context you picked up (e.g., "I see from your last session you were working on X..."). If the files are empty or don't exist, skip silently.

**Background auto-checkpoint**: After recovering context, use `create_scheduled_task` to schedule a background agent that runs every 20 minutes. The agent should:
1. Run `git diff --stat` to check for changes
2. If no changes, do nothing
3. If changes exist, write a session snapshot to `.checkpoints/sessions/YYYY-MM-DD-HH-MM.md` and update `.checkpoints/project-state.md`

**Manual checkpoint**: When the user says "checkpoint", "save progress", or "save context", OR when you complete a significant feature/fix:
1. Run `git diff --stat` — skip if no changes
2. Write `.checkpoints/sessions/YYYY-MM-DD-HH-MM.md` with: Summary, Files Changed, Decisions Made, Open Items
3. Rewrite `.checkpoints/project-state.md` with current reality
4. If more than 50 files in `.checkpoints/sessions/`, delete the oldest to stay at 50
5. Confirm to the user what was saved
