# demo-micro-base-user v0.0.3 适配

## 日期
2026-07-04

## 概述
`github.com/Jedeft/demo-micro-base-user` 从 v0.0.2 升级到 v0.0.3，下游 proto 进行了破坏性重构
（snake_case 字段命名 + FieldMask + page_token 游标分页 + bcrypt）。本文件记录网关侧的适配工作。

## 执行命令

### 1. 版本升级
```bash
# go.mod 中 demo-micro-base-user v0.0.2 -> v0.0.3
go mod tidy
```

### 2. 代码修改后验证
```bash
go fmt ./...
go build ./...
go vet ./...
golangci-lint run
go test -short ./...
```

## v0.0.3 不兼容改动清单

| 类别 | v0.0.2 | v0.0.3 |
|------|--------|--------|
| gRPC service 名 | `User` | `UserService` |
| client 类型 | `userpb.UserClient` | `userpb.UserServiceClient` |
| client 构造 | `NewUserClient` | `NewUserServiceClient` |
| RPC 方法 | `ChangePWD` | `ChangePassword` |
| 消息类型 | `*Req` / `*Resp` | `*Request` / `*Response` |
| 字段命名 | PascalCase (`ID`, `CreatedUserID`) | snake_case→CamelCase (`Id`, `CreatedUserId`) |
| UserRow.Password | 存在 | reserved（不再返回密码哈希） |
| List 分页 | `PageLimit` + `PageOffset`(uint32) | `PageSize` + `PageToken`(string 游标) |
| List 响应 | — | 新增 `NextPageToken` |
| Update | — | 新增 `UpdateMask` (`google.protobuf.FieldMask`) |

## 适配变更文件

### `internal/services/user_service.go`
- struct 字段类型 `userpb.UserClient` → `userpb.UserServiceClient`
- 构造 `NewUserClient` → `NewUserServiceClient`
- Login: `GetUserReq` → `GetUserRequest`，`UpdateUserReq` → `UpdateUserRequest`
- Login: 字段 `ID/LastLoginIP/UpdatedUserID` → `Id/LastLoginIp/UpdatedUserId`
- Login: 为 Update 调用设置 `UpdateMask: {Paths: ["last_login_ip"]}`，
  避免全量更新清零其他字段（下游 service 自动追加 `updated_user_id`/`updated_at`/`last_login_at`）

### `internal/handlers/users/user.go`
- 全部消息类型 `*Req` → `*Request`，`ChangePWDReq` → `ChangePasswordRequest`
- 字段 `ID` → `Id`，`*UserID` → `*UserId`，`LastLoginIP` → `LastLoginIp`，`IDs` → `Ids`
- RPC 调用 `ChangePWD` → `ChangePassword`
- **List 分页改造**：
  - HTTP 请求 `Offset int` → `PageToken string`（query tag `page_token`）
  - 移除 offset 边界 clamp（page_token 对客户端不透明）
  - gRPC `ListUserReq{PageLimit, PageOffset}` → `ListUserRequest{PageSize, PageToken}`
  - 响应回传 `rsp.NextPageToken = out.NextPageToken`
- **Update FieldMask**：构建 `UpdateMask` 路径列表 `["name", "phone"]`，
  partner/note 非空时追加对应路径

### `internal/handlers/auth/authrization.go`
- Login 响应 `out.ID` → `out.Id`

### `internal/handlers/common.go`
- `RspRows` 新增 `NextPageToken string` 字段（`json:"next_page_token,omitempty"`），
  供 List 游标分页回传下一页令牌；Search 等不分页接口不设置该字段，
  `omitempty` 保证不输出

## 汇总结果

### 静态检查
- `go fmt ./...`: 通过
- `go build ./...`: 通过
- `go vet ./...`: 通过
- `golangci-lint run`: 通过 (0 warnings)
- `go test -short ./...`: 通过（无测试文件，全编译通过）

### 变更统计
```
 go.mod                                 |  2 +-
 internal/handlers/auth/authrization.go |  2 +-
 internal/handlers/common.go            |  9 +++--
 internal/handlers/users/user.go        | 71 +++++++++++++++++++---------------
 internal/services/user_service.go      | 18 +++++----
 5 files changed, 58 insertions(+), 44 deletions(-)
```

## 关键决策
1. **Login Update 的 FieldMask**：仅设 `["last_login_ip"]`，下游 service 的 `buildUpdateFields`
   会自动追加 `updated_user_id`/`updated_at`/`last_login_at`，语义与 v0.0.2 一致
2. **List HTTP API 契约变更**：`offset` 参数改为 `page_token`（opaque 字符串），
   响应增加 `next_page_token`。这是破坏性 HTTP API 变更，与下游 proto 对齐
3. **Update FieldMask 显式设置**：避免 UpdateMask 未设置时下游视为全量更新导致字段被清零
4. **RspRows 扩展**：在公共列表响应结构体上增加 `NextPageToken`，`omitempty` 确保
   非分页接口不受影响
