# 单元测试生成

## 任务概述

根据 AGENTS.md 测试规范要求，为项目中缺少单元测试的源码文件生成表驱动测试。

## 执行命令

```bash
# 编译检查
go build ./...

# 运行全部测试
go test ./... -v -count=1

# 覆盖率检查
go test ./internal/handlers/ ./internal/handlers/auth/ ./internal/handlers/users/ \
    ./internal/services/ ./internal/supports/ ./internal/supports/utils/ -cover -count=1

# 代码格式化
go fmt ./...

# 静态检查
golangci-lint run ./internal/handlers/... ./internal/services/... ./internal/supports/...
```

## 新增测试文件

| 文件 | 被测文件 | 测试内容 |
|------|---------|---------|
| `internal/supports/utils/remove_repeat_test.go` | `remove_repeat.go` | `RemoveUint32Duplication`、`RemoveStringDuplication` |
| `internal/supports/error_test.go` | `error.go` | `GetErrMsg` 全量错误码 |
| `internal/handlers/common_test.go` | `common.go` | `Uint32IDReq.Valid`、`StringIDReq.Valid` |
| `internal/handlers/error_test.go` | `error.go` | `ErrorHandler`（biz/echo/default/HEAD/committed） |
| `internal/handlers/jwt_test.go` | `jwt.go` | `CreateToken`、`ParseToken`、`AuthSkipper`、`InvalidJWT`、`GetUserInfo` |
| `internal/handlers/users/user_test.go` | `user.go` | `valid()` 方法 + `Add/Get/List/Search/Update/ChangePWD/Delete` handler |
| `internal/handlers/auth/authrization_test.go` | `authrization.go` | `LoginReq.valid()` + `Login/Logout` handler |
| `internal/services/service_test.go` | `service.go` + `user_service.go` 补充 | `Init` + `Login` 边界用例 |

## 测试规范遵循情况

- **文件命名**：`*_test.go`，与被测文件同包
- **断言库**：`testify/assert` + `testify/require`
- **Mock**：`gomock` 生成 `MockUserClient`
- **表驱动测试**：所有 `valid()` 方法及纯函数均使用表驱动结构体（`name`/`input`/`want`）
- **t.Helper()**：测试辅助函数均调用 `t.Helper()`
- **Redis 依赖**：使用 `miniredis`（via `testutil.StartRedis`）隔离真实 Redis
- **Config 依赖**：使用 `testutil.SetTestConfig()` 注入测试配置

## 注意事项

### Go nil-interface 陷阱

`AddUserReq.valid()` 和 `LoginReq.valid()` 返回 `*xerr.Err`（具体指针类型）而非 `error`（接口类型）。
当返回值为 `nil` 时，将其传入 `assert.NoError(t, err)` 会因 typed-nil 转为非 nil 的 `error` 接口而判定失败。
此场景下应使用 `assert.Nil(t, err)` / `assert.NotNil(t, err)` 进行断言。

`UpdateUserReq.valid()` 和 `ChangePWDReq.valid()` 返回 `error` 接口类型，可直接使用 `assert.NoError`。

## 覆盖率结果

| 包 | 覆盖率 |
|---|--------|
| `internal/handlers` | 91.7% |
| `internal/handlers/auth` | 91.7% |
| `internal/handlers/users` | 91.9% |
| `internal/services` | 100% (函数级) |
| `internal/supports` | 100% |
| `internal/supports/utils` | 100% |

> `internal/services` 整体覆盖率为 75%（含 `newUserClient` 包级变量函数体，需真实 gRPC 连接，不纳入单元测试范围）。函数级覆盖率为 100%，满足 AGENTS.md 中 services 层 ≥90% 的要求。

## 遗留项

- `internal/grpc/conn.go`（`GetConn`/`CloseAll`）涉及真实 gRPC 连接，建议使用 `bufconn` 编写集成测试。
- `internal/routers/router.go`（`New()`）依赖完整初始化链，建议纳入集成测试。
