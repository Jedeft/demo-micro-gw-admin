# RPC 风格 HTTP API 设计规范

> **核心理念**：本项目采用 RPC（Remote Procedure Call）风格的 HTTP API 设计。API 代表的是**函数/方法调用**，而非资源操作。设计目标是与本地方法调用保持一致，让开发者专注于业务逻辑而非通信细节。

---

## 一、核心原则

### 1.1 动词优先，而非资源
- URL 表达的是**动作/命令**，而非资源。
- RPC 接口类似于本地方法调用，每个接口对应一个明确的业务操作。

### 1.2 接口即方法
- 每个 API 接口应等同于一个 Service 层方法，命名直接反映业务意图。
- 避免将 RPC 返回值设计成包含 `code`/`message`/`data` 的包装结构——这违背了 RPC 的设计初衷。异常应通过标准 HTTP 状态码 + 错误体传递。

### 1.3 简单直接
- 返回值就是业务数据本身，不额外包裹。
- 错误通过 HTTP 状态码和错误响应体传达，不混入正常响应结构。

---

## 二、URL 设计规范

### 2.1 格式
```
/{服务名}/{方法名}
```
或
```
/{业务域}/{动作}
```

### 2.2 命名规则
| 规则 | 说明 | 示例 |
|------|------|------|
| 全部小写 | 禁止驼峰 | `order` ✅ / `Order` ❌ |
| 单词间用 `/` | 系统可用 `/` | `/order/cancel` |
| 方法名用动词 | 体现业务动作 | `cancel`、`refund`、`batchDelete`、`export` |
| 避免 CRUD 式命名 | 除非确实仅为增删改查 | `createOrder` ✅ / `add` ❌ |

### 2.3 好的示例 vs 坏的示例
```text
# ✅ 好的示例（业务语义明确）
POST /order/cancel          # 取消订单
POST /payment/refund        # 退款
POST /user/batchDelete      # 批量删除用户
GET  /order/getDetail       # 获取订单详情
POST /report/export         # 导出报表

# ❌ 坏的示例（REST 风格，不符合本规范）
DELETE /orders/123          # 看不出是物理删除还是逻辑删除
PATCH /orders/123           # 看不出要改什么
GET  /orders                # 看不出要查列表还是统计
```

---

## 三、HTTP 方法使用规范

RPC 风格 API **仅支持 GET 和 POST** 两种 HTTP 方法。

| 方法 | 使用场景 | 参数位置 |
|------|----------|----------|
| **GET** | 纯查询操作，无副作用 | URL Query String |
| **POST** | 所有写入、修改、删除、触发类操作 | Request Body (JSON) |

```text
# 查询类 → GET
GET /user/get?userId=123

# 所有修改/操作类 → POST
POST /order/cancel
Body: {"orderId": "ORD-001", "reason": "用户申请"}
```

> **注意**：**禁止使用** PUT、DELETE、PATCH。这些方法的幂等性语义在复杂业务中难以保证，统一使用 POST 更清晰可控。

---

## 四、参数传递规范

### 4.1 GET 请求参数
- 所有参数放在 **URL Query String** 中。
- 参数名使用 **camelCase**。
- 复杂查询条件（嵌套对象、数组）—— 如果参数过多或结构复杂，改用 **POST + Body**。

```bash
# ✅ 简单查询用 GET
GET /user/list?page=1&pageSize=20&status=active

# ✅ 复杂查询用 POST（参数放 Body）
POST /user/search
Body: {"page":1, "pageSize":20, "filters":{"name":"张三","role":"admin","createTime":{"start":"2026-01-01","end":"2026-06-30"}}}
```

### 4.2 POST 请求参数
- 所有参数放在 **Request Body** 中，格式为 **application/json**。
- Body 结构即为方法参数结构，与服务端方法签名一一对应。

```json
POST /order/submitApproval
{
  "orderId": "ORD-001",
  "approverId": "user_123",
  "comment": "已核实，同意"
}
```

### 4.3 参数命名
- 使用 **camelCase**（如 `userId`、`orderId`、`createTime`）。
- 布尔类型以 `is`/`has`/`can` 开头（如 `isActive`、`hasPermission`）。

---

## 五、响应规范

### 5.1 成功响应
**直接返回业务数据**，不额外包裹 `code`/`message`/`data`。

```json
// 查询单个对象
{
  "userId": "user_123",
  "userName": "张三",
  "email": "zhangsan@example.com",
  "status": "active"
}

// 查询列表
{
  "list": [
    {"userId": "user_001", "userName": "张三"},
    {"userId": "user_002", "userName": "李四"}
  ],
  "total": 100,
  "page": 1,
  "pageSize": 20
}

// 操作成功（无返回数据）
{
  "success": true
}
// 或直接返回空对象
{}
```

### 5.2 错误响应
错误通过 **HTTP 状态码** + **统一错误体** 返回。

```json
{
  "code": "ORDER_NOT_FOUND",
  "message": "订单不存在",
  "details": "未找到订单ID为 ORD-001 的订单",
  "traceId": "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `code` | string | ✅ | 错误码，便于前端程序化处理 |
| `message` | string | ✅ | 用户友好的错误描述 |
| `details` | string | ❌ | 详细的错误信息（便于排查），仅在网关服务开启Debug模式时才会有详细错误信息展示 |
| `traceId` | string | ✅ | 全链路追踪 ID，用于问题定位 |

### 5.3 HTTP 状态码使用
| 状态码 | 语义 | 使用场景 |
|--------|------|----------|
| 200 | OK | 请求成功 |
| 400 | Bad Request | 参数校验失败、格式错误 |
| 401 | Unauthorized | 未认证（Token 缺失或无效） |
| 403 | Forbidden | 无权限 |
| 404 | Not Found | 资源不存在 |
| 409 | Conflict | 业务冲突（如重复提交、状态不允许操作） |
| 422 | Unprocessable Entity | 业务逻辑校验失败（如库存不足、余额不够） |
| 429 | Too Many Requests | 请求频率超限 |
| 500 | Internal Server Error | 服务端未知错误 |

---

## 六、请求头规范

### 6.1 必须包含的 Header
| Header | 说明 | 示例 |
|--------|------|------|
| `Content-Type` | 固定为 `application/json` | `application/json; charset=utf-8` |
| `Authorization` | Bearer Token 认证 | `Bearer eyJhbGciOiJIUzI1NiIs...` |
| `X-Request-Id` | 请求唯一标识（由客户端生成或网关注入） | `req-20260101-001` |

### 6.2 建议包含的 Header
| Header | 说明 | 示例 |
|--------|------|------|
| `X-Client-Version` | 客户端版本号 | `v2.1.0` |
| `X-User-Id` | 操作用户 ID（网关层注入，前端无需关心） | `user_123` |

---

## 七、错误码规范

### 7.1 错误码格式
`{模块}_{具体错误}`，全大写，下划线分隔。

```text
# 示例
USER_NOT_FOUND
USER_ALREADY_EXISTS
ORDER_STATUS_INVALID
ORDER_CANNOT_CANCEL
INSUFFICIENT_BALANCE
PERMISSION_DENIED
INVALID_PARAMETER
RATE_LIMIT_EXCEEDED
```

### 7.2 错误码分类
| 前缀 | 说明 | HTTP 状态码 |
|------|------|-------------|
| `INVALID_*` | 参数/格式错误 | 400 |
| `NOT_FOUND` | 资源不存在 | 404 |
| `USER_*` | 用户相关错误 | 401/403 |
| `PERMISSION_*` | 权限相关错误 | 403 |
| `*_STATUS_*` | 状态机相关错误 | 409 |
| `INSUFFICIENT_*` | 资源不足（余额、库存等） | 422 |
| `RATE_*` | 限流相关 | 429 |
| `SYSTEM_*` | 系统内部错误 | 500 |

---

## 八、安全规范

| 规范 | 说明 |
|------|------|
| **HTTPS** | 生产环境**必须**使用 HTTPS |
| **认证** | 所有非公开接口必须通过 Bearer Token（JWT）认证 |
| **敏感信息** | 响应和日志中严禁包含密码、密钥、手机号明文等敏感信息 |
| **操作审计** | 所有写操作必须记录操作人、操作时间、操作内容 |

---

## 九、版本管理

### 9.1 版本策略
- 接口版本号放在 **URL 路径**中：`/v1/{service}/{method}`
- 主版本号变更表示**不兼容的破坏性变更**
- 向后兼容的新增字段不升级版本号

```text
POST /v1/order/cancel
POST /v2/order/cancel   # v2 签名或行为与 v1 不兼容
```

### 9.2 兼容性规则
- ✅ 允许：新增可选字段、新增接口、放宽校验规则
- ❌ 禁止：删除字段、修改字段类型、修改必填性（可选→必填）、修改错误码语义

---

## 十、接口设计检查清单

在新增或修改接口时，逐项确认：

- [ ] URL 是否表达了明确的业务动作？
- [ ] 方法选择是否正确（查询用 GET，其他用 POST）？
- [ ] 参数命名是否使用 camelCase？
- [ ] 成功响应是否直接返回业务数据（无多余包裹）？
- [ ] 错误响应是否包含 `code`、`message`、`traceId`？
- [ ] 错误码是否在错误码表中已定义？
- [ ] 是否添加了必要的认证和鉴权？
- [ ] 是否记录了操作审计日志？
- [ ] 接口变更是否向后兼容？

---

## 十一、参考示例

### 查询订单详情
```bash
GET /v1/order/getDetail?orderId=ORD-001
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```
**成功响应 (200)**：
```json
{
  "orderId": "ORD-001",
  "userId": "user_123",
  "amount": 299.00,
  "status": "paid",
  "createTime": "2026-01-01T10:00:00Z"
}
```
**错误响应 (404)**：
```json
{
  "code": "ORDER_NOT_FOUND",
  "message": "订单不存在",
  "traceId": "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
}
```

### 取消订单
```bash
POST /v1/order/cancel
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Content-Type: application/json
{
  "orderId": "ORD-001",
  "reason": "用户申请取消"
}
```
**成功响应 (200)**：
```json
{
  "success": true
}
```
**错误响应 (409)**：
```json
{
  "code": "ORDER_STATUS_INVALID",
  "message": "订单状态不允许取消",
  "details": "当前订单状态为 'shipped'，无法取消",
  "traceId": "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
}
```

---

> **参考资料**：本规范参考了阿里云 OpenAPI RPC 风格、Azure API 设计指南、京东云 RPC 接口设计实践等现网标准。