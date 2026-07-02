# Go 编码规范参考

> 本文档整合了三份官方 Go 规范指南，作为项目 Go 编码的统一参考标准。
>
> **规范优先级（从上到下逐级覆盖）**：
> 1. **[Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)** — 最高优先级，Google 官方代码评审检查清单
> 2. **[Effective Go](https://go.dev/doc/effective_go)** — 补充性语言风格与设计原则
> 3. **[Google Go Style Guide](https://google.github.io/styleguide/go/decisions)** — 补充性参考，涵盖更全面的风格决策
>
> **冲突处理规则**：当规范之间存在冲突时，**优先级高的规范覆盖优先级低的**。例如：
> - CodeReviewComments 说 error strings 不应大写 → 以它为准
> - CodeReviewComments 说 Context 不存入结构体 → 以它为准
> - Google Style Guide 仅作为前两者未涉及场景的补充

---

## 目录

1. [格式化](#1-格式化)
2. [注释](#2-注释)
3. [命名约定](#3-命名约定)
4. [包组织与导入](#4-包组织与导入)
5. [声明与初始化](#5-声明与初始化)
6. [函数与方法](#6-函数与方法)
7. [错误处理](#7-错误处理)
8. [并发与同步](#8-并发与同步)
9. [Context 约定](#9-context-约定)
10. [测试](#10-测试)
11. [类型与接口](#11-类型与接口)
12. [性能与优化](#12-性能与优化)
13. [其他约定](#13-其他约定)
14. [CLAUDE.md 引用配置](#14-claudemd-引用配置)

---

## 1. 格式化

### 1.1 Gofmt（#1 优先级 — CodeReviewComments）

**规则**：所有 Go 代码在提交前必须通过 `gofmt`（或 `goimports`）格式化。

```bash
gofmt -l -w <file>     # 格式化并原地写入
goimports -w <file>    # 推荐：同时处理导入排序
```

**原因**：消除所有格式争论——`gofmt` 的风格就是官方标准，没有商量余地。缩进、空格、换行全部由工具决定。

### 1.2 行长（#3 优先级 — Google Style Guide）

- GitHub 上会显示垂直分割线，建议行长不超过 **100 字符**
- 超出时手动换行，保持清晰的结构
- 不要为了缩短行长而使用荒谬的缩写

```go
// OK：如果超出行长，合理换行
func (s *LongServiceName) ProcessWithContext(ctx context.Context,
    req *longpackage.LongRequestName) (*longpackage.LongResponseName, error) {
```

## 2. 注释

### 2.1 Comment Sentences（#1 优先级 — CodeReviewComments）

- 注释应使用**完整的英文句子**，以句号结尾
- 首句以被注释的元素名称开头

```go
// Request represents a request to run a command.
type Request struct{ ... }

// Encode writes the JSON encoding of req to w.
func Encode(w io.Writer, req *Request) error { ... }
```

### 2.2 Doc Comments（#1 优先级 — CodeReviewComments）

- 所有导出的（Exported）名字（类型、函数、方法、常量、变量、包）必须写 doc comment
- 格式：`// <Name> <verb-phrase>.` 或 `// <Name> <is/represents/holds> ...`

```go
// Package fmt implements formatted I/O.
package fmt

// MaxSize is the maximum allowed size in bytes.
const MaxSize = 1024

// IsExist reports whether the file or directory exists.
func IsExist(path string) bool { ... }
```

**Good / Bad 对照**：

```go
// Bad：不是完整句子
// This function reads a file
func Read(path string) ([]byte, error)

// Good：完整句子，以名字开头
// Read reads the file at path and returns its contents.
func Read(path string) ([]byte, error)
```

### 2.3 Package Comment（#1 优先级 — CodeReviewComments）

- 每个包必须有一个包级注释
- 格式：`// Package <name> provides ...`
- 对于 `main` 包：注释描述二进制文件的用途

```go
// Package route provides HTTP route matching for the web framework.
package route

// 或简写（包内容简单时）：
// Package math provides basic math constants and functions.
package math
```

### 2.4 内联注释（#3 优先级 — Google Style Guide）

- 解释 **为什么** 这样写，而不是 **做了什么**
- 代码应该自解释做了什么
- 不要在明显的代码上加注释

```go
// Good：解释为什么
// Re-sort the list because the input order is not guaranteed.
sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })

// Bad：注释复述了代码已经表达的信息
// Sort items by ID
sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
```

## 3. 命名约定

### 3.1 MixedCaps（#1 优先级 — CodeReviewComments）

- 使用 `MixedCaps`（驼峰），而不是下划线
- 缩写词保持大写（除非是包的内部缩写）

```go
// Good
var UserID int
var RequestID string
var HTTPClient *http.Client
var JSONEncoder *json.Encoder

// Bad
var user_id int
var http_client *http.Client
var JsonEncoder *json.Encoder
```

### 3.2 Initialisms / 缩写（#1 优先级 — CodeReviewComments）

- 缩写/首字母缩略词要么全大写，要么全小写
- 不要出现 `HttpRequest`、`UrlParser` 这样的写法

```go
// Good
func ServeHTTP()
func ParseURL()
var userID string
var idMap map[string]int

// Bad
func ServeHttp()
func ParseUrl()
var userId string
```

**常见缩写列表**：

| 正确 | 错误 |
|------|------|
| `HTTP`, `http` | `Http`, `http` |
| `URL`, `url` | `Url` |
| `ID`, `id` | `Id` |
| `JSON`, `json` | `Json` |
| `API`, `api` | `Api` |
| `ACL`, `acl` | `Acl` |
| `HTTPS`, `https` | `Https` |

### 3.3 Variable Names（#1 优先级 — CodeReviewComments）

- **短作用域用短名称**：`i`、`r`、`w`、`err`、`ch`
- **长作用域用描述性名称**：`userCount`、`requestBody`
- 不要使用匈牙利命名法（`u32Count`）
- 不要使用蛇形命名（`user_count`）

```go
// 短作用域（推荐）
for i, r := range rows { ... }

// 长作用域（推荐）
userCount := len(activeUsers)

// Bad：作用域很短却用了长名字
for theCurrentIndex := range items { ... }
```

### 3.4 Receiver Names（#1 优先级 — CodeReviewComments）

- 使用一个或两个字母的缩写，反映 receiver 的类型
- 在同一个包内保持**一致性**——同一个类型的 receiver 名称必须一致
- 不要使用 `self` 或 `this`

```go
// Good
func (s *Server) Start() { ... }
func (s *Server) Stop() { ... }
func (t *Token) Validate() error { ... }
func (tok *Token) Expire() { ... }  // 也 OK，扩展单字母会有更多上下文

// Bad
func (self *Server) Start() { ... }       // 不要用 self/this
func (server *Server) Start() { ... }     // 太啰嗦
func (s *Server) Start() { ... }
func (this *Server) Stop() { ... }        // 不一致！
```

### 3.5 Receiver Type 选择（#1 优先级 — CodeReviewComments）

| 使用指针 receiver | 使用值 receiver |
|-------------------|-----------------|
| 方法需要修改 receiver | 方法不修改 receiver |
| receiver 是大结构体（或 slice/map 字段） | receiver 是小值类型 |
| 一致性要求——同一个 type 的方法应保持一致 | — |

**核心规则**：
- **不要混用**：对同一个类型，不要一部分方法用值 receiver，另一部分用指针 receiver（除非绝对必要）
- 如果某个方法必须用指针 receiver，那么该类型的所有方法都应该用指针 receiver

```go
// 只有一种 receiver 类型 —— 推荐
type Server struct { ... }
func (s *Server) Start() error { ... }     // 指针
func (s *Server) Stop() error { ... }      // 指针（一致）

// 对简单的小值类型，可以用值 receiver
type Point struct { X, Y float64 }
func (p Point) Distance() float64 { ... }  // 值
func (p *Point) Scale(f float64) { ... }   // ❌ 不要混用
```

### 3.6 Function & Method Names（#1 优先级 — CodeReviewComments）

- 命名描述效果（effect），而非实现细节
- 导出（Exported）函数：`PascalCase`
- 未导出（Unexported）函数：`camelCase`

```go
// Good：描述了效果
func (s *Store) LookupUser(id string) (*User, error)
func (s *Store) DeleteUser(id string) error

// Bad：描述了实现
func (s *Store) QueryUserByIDFromPostgres(id string) (*User, error)
```

### 3.7 Package Names（#1 优先级 — CodeReviewComments + Effective Go）

- 简短、小写、单单词
- 不要用下划线或驼峰
- 名称应该描述包的内容，而不是文件结构
- **避免**使用 `common`、`util`、`base`、`shared` 等模糊名称

```go
// Good
package http
package zip
package user

// Bad
package common_lib
package HttpHandler
package utils
package myCompany
package shared
```

**包命名黄金法则**（Effective Go）：包的使用者在引用时，应该不需要查找：`user.LookupByID` 而不是 `user.UserLookup`。

### 3.8 Named Result Parameters（#1 优先级 — CodeReviewComments）

**何时使用**：
- 多个返回值类型相同（用名字区分语义）
- 函数简短且在错误路径中明显可见
- 返回值语义不直观时

```go
// Good：两个 string 类型相同，用名字区分
func LookupName(id string) (firstName, lastName string, err error)

// OK：简短函数中可使用裸返回
func Split(path string) (dir, file string) {
    dir, file = path.Split(path)
    return  // 裸返回 —— 仅在短函数中可接受
}

// Bad：长函数中使用裸返回（难以追踪哪个返回值被设置）
func (s *Server) process(data []byte) (result *Result, err error) {
    // ... 很多行代码 ...
    // ... 更多代码 ...
    return  // 这个裸返回让读者困惑
}
```

### 3.9 Naked Returns（#1 优先级 — CodeReviewComments）

- **裸返回（naked return）仅在短函数中可接受**
- **导出函数不要使用裸返回**
- 函数长度超过 10 行时禁止裸返回

```go
// Good：短函数，裸返回清晰
func add(a, b int) (sum int) {
    sum = a + b
    return
}

// Bad：长函数中的裸返回令人困惑
func process(ctx context.Context, input *Request) (result *Response, err error) {
    if err = validate(input); err != nil {
        return  // 读者需要往上翻才知道返回了什么
    }
    // ... 50行代码 ...
    return
}
```

## 4. 包组织与导入

### 4.1 Imports — 分组与排序（#1 优先级 — CodeReviewComments）

导入分为**三个分组**，组间空行分隔：

1. 标准库
2. 第三方包
3. 项目内部包

```go
import (
    "fmt"
    "io"
    "net/http"
    "os"

    "github.com/google/uuid"
    "golang.org/x/sync/errgroup"

    "your-project/internal/auth"
    "your-project/internal/config"
)
```

### 4.2 Import Dot / 点导入（#1 优先级 — CodeReviewComments）

- **不推荐点导入（`.` import）**
- 仅在极少数特殊场景（如测试对外部包的 mock）下使用
- 不要在常规生产代码中使用

```go
// 不推荐
import (
    . "fmt"  // 导致 Println() 而不是 fmt.Println()
)
```

### 4.3 Blank Imports（#1 优先级 — CodeReviewComments）

- 仅用于副作用（side-effect），如注册数据库驱动
- 必须附带注释解释原因

```go
import (
    _ "github.com/lib/pq"  // Register PostgreSQL driver in init()
)
```

### 4.4 Package Comment（#1 优先级 — CodeReviewComments）

- 每个包必须有 package-level 注释
- 格式见 [2.3 Package Comment](#23-package-comment-1-优先级--codereviewcomments)

### 4.5 Goimports 工具

推荐使用 `goimports` 而非手动管理导入：

```bash
go install golang.org/x/tools/cmd/goimports@latest
goimports -w .
```

它会自动格式化 + 排序导入 + 移除未使用的导入。

## 5. 声明与初始化

### 5.1 Declaring Empty Slices（#1 优先级 — CodeReviewComments）

**规则**：优先使用 `var s []T` 而不是 `s := []T{}`。

```go
// 推荐（nil slice，无内存分配）
var s []int

// 不推荐（非 nil 的空 slice，有分配）
s := []int{}
```

**为什么**：
- `var s []int` 没有分配底层数组（nil slice）
- `s := []int{}` 分配了一个空的底层数组（non-nil but zero-length）
- 对几乎所有标准操作（`append`、`len`、`cap`、`range`、JSON 编码），两者行为完全一致

**例外**：如果 JSON 序列化必须输出 `[]` 而不是 `null`，使用 `[]int{}` 或 `make([]int, 0)`。

### 5.2 Zero-value Structs（#1 优先级 — CodeReviewComments）

- 如果声明一个零值结构体，使用 `var` 而不是 `:=` 如果后者需要显式 `{}`

```go
// 推荐
var myStruct StructType

// 也是可以的，但 var 更简洁
myStruct := StructType{}

// 如果结构体有字段需初始化，用 `:=` + 复合字面量
conf := Config{
    Timeout: 30,
    MaxSize: 1024,
}
```

### 5.3 Variable Declarations（#1 优先级 — CodeReviewComments）

```go
// 零值声明 —— 用 var（更简洁）
var count int
var buf bytes.Buffer

// 非零值声明 —— 用 :=
name := "hello"
maxSize := 1024

// 包级变量 —— 用 var
var globalConfig = loadConfig()
```

### 5.4 Composite Literals（#2 优先级 — Effective Go）

**构造器替代**：Go 不需要显式的构造器函数，用复合字面量代替。

```go
// 不需要 NewPoint 构造器
p := Point{X: 3, Y: 4}

// 命名参数版本（推荐，更清晰）
vf := Vertex{X: 3, Y: 4}

// 无命名版本（谨慎使用，依赖字段顺序）
vf := Vertex{3, 4}
```

### 5.5 make vs new（#2 优先级 — Effective Go）

| `new(T)` | `make(T, args)` |
|----------|----------------|
| 返回 `*T`（零值指针） | 返回 `T`（初始化后的值） |
| 对所有类型可用 | **仅**用于 **slice / map / channel** |
| 分配零值内存 | 初始化内部数据结构 |

```go
// new
p := new(Point)        // *Point，字段为零值
m := new(map[string]int) // *map[string]int——指向 nil map

// make
m := make(map[string]int)     // 可用的空 map
s := make([]int, 0, 10)       // len=0, cap=10 的 slice
ch := make(chan int, 100)     // 带缓冲的 channel
```

### 5.6 Map 初始化（#2 优先级 — Effective Go）

```go
// 推荐 —— 字面量
m := map[string]int{
    "one": 1,
    "two": 2,
}

// 空 map —— 用 make（而非 nil map）
m := make(map[string]int)
m["key"] = 1  // nil map 这里会 panic

// 不要声明为 nil map
var m map[string]int  // nil，写入会 panic
```

## 6. 函数与方法

### 6.1 函数参数（#3 优先级 — Google Style Guide）

- 参数列表**保持简短**——如果超过 3-4 个参数，使用配置结构体
- 避免使用 `bool` 参数改变函数行为（改用拆分函数或枚举）

```go
// Bad：bool 参数让调用者困惑
func Save(w io.Writer, data []byte, compress bool) error
Save(w, data, true)  // 这个 true 是什么意思？

// Good：拆分为不同函数
func Save(w io.Writer, data []byte) error
func SaveCompressed(w io.Writer, data []byte) error

// 或者：Options 模式
type SaveOptions struct {
    Compress bool
    FileMode os.FileMode
}
func SaveWithOptions(w io.Writer, data []byte, opts SaveOptions) error
```

### 6.2 可变参数（#3 优先级 — Google Style Guide）

- 对可选的零个或多个参数使用可变参数
- 可变参数应放在参数列表的最后

```go
// Good
func Sum(nums ...int) int

// 结合 Options 模式的变体
func Open(name string, opts ...Option) (*File, error)
```

### 6.3 Defer（#1 优先级 — CodeReviewComments）

- 用 `defer` 释放资源（关闭文件、解锁互斥锁等）
- defer 应紧跟在资源获取之后——让读者一眼看出生命周期

```go
// Good：defer 紧跟在成功获取资源之后
func (s *Store) Write(data []byte) error {
    f, err := os.Create(filepath)
    if err != nil {
        return err
    }
    defer f.Close()  // 紧跟在 os.Create 检查之后
    // ... 使用文件 ...
}

// Good：defer 解锁
mu.Lock()
defer mu.Unlock()
```

### 6.4 函数长度（#1 优先级 — CodeReviewComments）

- 函数应当**做一件事，且把它做好**
- 如果函数超过 40-50 行，考虑拆分解构
- 这不是严格的硬性规定——但长函数通常是重构的信号

### 6.5 函数参数 vs 方法接收者（#1 优先级 — CodeReviewComments + Effective Go）

```go
// 如果函数主要操作某个类型的实例，用方法（receiver）
func (u *User) Activate() error   // ✅

// 如果函数操作多个类型的组合，用独立函数
func SendNotification(user *User, msg *Message) error  // ✅
```

## 7. 错误处理

### 7.1 Handle Error（#1 优先级 — CodeReviewComments）

- **每个 error 都必须被检查或明确忽略**
- 不要使用 `_` 忽略错误——**除非**你深思熟虑决定忽略

```go
// Good
f, err := os.Open(path)
if err != nil {
    return fmt.Errorf("open %s: %w", path, err)
}

// Bad：忽略了错误
f, _ := os.Open(path)

// 当确实确定可以忽略时，用 _ 并加注释
_ = r.Close()  // 忽略 close 错误，因为文件将以只读方式打开
```

### 7.2 Error Strings（#1 优先级 — CodeReviewComments）

- 错误信息**首字母不要大写**（除非是专有名词）
- 错误信息**末尾不要加标点符号**

```go
// Good
fmt.Errorf("invalid input: %s", input)
errors.New("user not found")

// Bad
fmt.Errorf("Invalid input: %s", input)     // 大写首字母 ❌
errors.New("user not found.")              // 末尾标点 ❌
```

**原因**：错误信息经常被链式包装或通过 `errors.Is` / `errors.As` 比较，大写/标点会在链式组装后看起来不自然：
`something failed: Invalid input: foo.` ← 难看。

### 7.3 Error Wrapping（#1 优先级 — CodeReviewComments + #3 Google Style Guide）

- 使用 `fmt.Errorf("context: %w", err)` 包装错误
- 这保留了原始错误的类型信息，使 `errors.Is` / `errors.As` 可以工作

```go
// Good：用 %w 包装
if err := doSomething(); err != nil {
    return fmt.Errorf("doSomething failed: %w", err)
}

// 当不需要 unwrap 时，用 %v（而非 %w）也可以
if err := doSomething(); err != nil {
    return fmt.Errorf("doSomething failed: %v", err)
}
```

### 7.4 Indent Error Flow（#1 优先级 — CodeReviewComments）

- 正常代码路径不应缩进过深
- 处理错误的代码应该缩进，然后**立即返回**

```go
// Good：正常路径无缩进
f, err := os.Open(path)
if err != nil {
    return err
}
defer f.Close()
// 正常代码在这里，无额外缩进

// Bad：正常路径被过度缩进
f, err := os.Open(path)
if err == nil {
    defer f.Close()
    // ... 大量代码 ...
    return nil
}
return err
```

### 7.5 Sentinel Errors vs Error Types（#1 优先级 — CodeReviewComments）

```go
// Sentinel error（包级导出，用 errors.Is 比较）
var ErrNotFound = errors.New("item not found")

// Error type（用 errors.As 比较）
type ValidationError struct {
    Field string
    Err   error
}
func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed on %s: %v", e.Field, e.Err)
}
func (e *ValidationError) Unwrap() error { return e.Err }

// 使用
if errors.Is(err, ErrNotFound) { ... }

var valErr *ValidationError
if errors.As(err, &valErr) { ... }
```

### 7.6 Don't Panic（#1 优先级 — CodeReviewComments）

- `panic` 表示程序遇到了不可恢复的错误
- 不要在正常错误处理中使用 `panic`
- 库函数应避免 panic（让调用者决定如何处理错误）
- `panic` 仅在 `main` 包中用于致命初始化错误

### 7.7 In-Band Errors（#1 优先级 — CodeReviewComments）

- 不要将结果和错误混在一个值里（如用 -1 表示 "not found"）
- 使用多返回值：`(result, error)`
- Go 不鼓励 C 风格的 in-band 错误信号

```go
// Good
func Lookup(id string) (*User, error)

// Bad：用特殊值表示错误
func Lookup(id string) *User  // nil 表示 not found？读者不知道
```

## 8. 并发与同步

### 8.1 Goroutine Lifetimes（#1 优先级 — CodeReviewComments）

**规则**：启动 goroutine 时，必须知道它何时退出。

```go
// Good：goroutine 生命周期可控
func (s *Server) Start(ctx context.Context) {
    s.wg.Add(1)
    go func() {
        defer s.wg.Done()
        for {
            select {
            case <-ctx.Done():
                return
            case job := <-s.jobCh:
                s.process(job)
            }
        }
    }()
}

func (s *Server) Shutdown(ctx context.Context) error {
    // 等待所有 goroutine 退出
    done := make(chan struct{})
    go func() {
        s.wg.Wait()
        close(done)
    }()
    select {
    case <-done:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}

// Bad：goroutine 泄漏
func (s *Server) Start() {
    go func() {
        for {
            select {
            case job := <-s.jobCh:
                s.process(job)
            }
            // 没有退出条件！如果 jobCh 关闭，这个 goroutine 永远不会结束
        }
    }()
}
```

### 8.2 Mutex 零值可用（#1 优先级 — CodeReviewComments）

- `sync.Mutex` 和 `sync.RWMutex` 的**零值是可用的**，不需要初始化

```go
// Good：零值直接用
type Counter struct {
    mu    sync.Mutex
    count int
}

func (c *Counter) Inc() {
    c.mu.Lock()
    c.count++
    c.mu.Unlock()
}
```

### 8.3 Channel 用法（#2 优先级 — Effective Go）

- 默认是无缓冲 channel（同步通信）
- 有缓冲 channel 用于限速或解耦生产者/消费者
- 发送者关闭 channel，接收者不关闭
- 使用 `range` 遍历 channel 直到关闭

```go
ch := make(chan int)        // 无缓冲
ch := make(chan int, 100)   // 有缓冲

// 发送者关闭 channel
close(ch)

// 接收者用 range 读取
for v := range ch {
    fmt.Println(v)
}
```

### 8.4 errgroup 模式

- 多个独立 goroutine 返回错误时，使用 `errgroup`
- 这比手动管理 `sync.WaitGroup` + channel 更简洁

```go
g, ctx := errgroup.WithContext(ctx)

g.Go(func() error {
    return doSomething(ctx)
})

g.Go(func() error {
    return doSomethingElse(ctx)
})

if err := g.Wait(); err != nil {
    return fmt.Errorf("group failed: %w", err)
}
```

### 8.5 同步函数优先（#1 优先级 — CodeReviewComments）

- 如果函数内部不需要并发，优先暴露同步 API
- 只在必要时暴露异步版本
- 同步函数更容易测试、推理和调试

```go
// 优先：同步版本
func (s *Store) Get(key string) (*Item, error)

// 只在必要时提供异步版本
func (s *Store) GetAsync(ctx context.Context, key string) chan GetResult {
    ch := make(chan GetResult, 1)
    go func() {
        item, err := s.Get(key)
        ch <- GetResult{item, err}
    }()
    return ch
}
```

## 9. Context 约定

### 9.1 Contexts — 作为第一个参数（#1 优先级 — CodeReviewComments）

- `context.Context` 必须是函数的**第一个参数**
- 参数名通常为 `ctx`

```go
// Good
func DoSomething(ctx context.Context, arg string) error {
    // ...
}

// Bad
func DoSomething(arg string, ctx context.Context) error { ... }
func DoSomething(arg string, other int, ctx context.Context) error { ... }
```

### 9.2 Context 不存入结构体（#1 优先级 — CodeReviewComments）

- **不要**将 `context.Context` 存入结构体字段
- Context 是函数的参数，应该随请求流显式传递

```go
// Bad：Context 存入结构体
type Handler struct {
    ctx context.Context  // ❌ 不要这样做
}

// Good：在需要时传参
type Handler struct {
    // 不存 context
}
func (h *Handler) Handle(ctx context.Context, req *Request) error { ... }
```

### 9.3 Context Key 类型（#1 优先级 — CodeReviewComments）

- 如果需要在 context 中存自定义值，创建一个私有类型作为 key
- 不要直接用 `string` 类型作为 context key（包之间会冲突）

```go
// Good：自定义类型避免冲突
type requestIDKey struct{}

func WithRequestID(ctx context.Context, id string) context.Context {
    return context.WithValue(ctx, requestIDKey{}, id)
}

func RequestIDFromContext(ctx context.Context) (string, bool) {
    id, ok := ctx.Value(requestIDKey{}).(string)
    return id, ok
}

// Bad：string key 可能与其他包冲突
ctx = context.WithValue(ctx, "requestID", id)
```

### 9.4 Context 传递规则

- Context 应贯穿调用链——当函数的调用链中有潜在的超时、取消或追踪需求时，始终传递 ctx
- 即使当前函数"不需要"，也接受 ctx——这是面向未来的设计

## 10. 测试

### 10.1 Table Driven Tests（#1 优先级 — CodeReviewComments）

- 使用表驱动测试覆盖多个输入/输出组合
- 表驱动测试易于扩增新的测试用例

```go
func TestSplit(t *testing.T) {
    tests := []struct {
        name  string
        input string
        sep   string
        want  []string
    }{
        {name: "comma", input: "a,b,c", sep: ",", want: []string{"a", "b", "c"}},
        {name: "space", input: "a b c", sep: " ", want: []string{"a", "b", "c"}},
        {name: "none", input: "abc", sep: ",", want: []string{"abc"}},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := strings.Split(tt.input, tt.sep)
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("Split(%q, %q) = %v, want %v", tt.input, tt.sep, got, tt.want)
            }
        })
    }
}
```

**推荐结构**：

| 字段 | 说明 |
|------|------|
| `name` | 测试案例名称——传给 `t.Run` |
| `input` / `args` | 输入参数 |
| `want` / `expected` | 期望输出 |
| `err` | 如果测试错误场景，期望的错误 |

### 10.2 Useful Test Failures（#1 优先级 — CodeReviewComments）

- 测试失败消息应包含**所有相关上下文**——输入、期望值、实际值
- 使用格式化的错误信息：`t.Errorf("method(%v) = %v, want %v", input, got, want)`
- 不要只打印"test failed"或"expected X"

```go
// Good
t.Errorf("Split(%q, %q) = %v, want %v", tt.input, tt.sep, got, tt.want)

// Bad —— 缺少输入上下文
t.Errorf("got %v, want %v", got, tt.want)
```

### 10.3 t.Helper（#1 优先级 — CodeReviewComments）

- 测试辅助函数中必须调用 `t.Helper()`
- 这样错误报告会指向调用辅助函数的测试行，而非辅助函数内部

```go
// Good
func assertEqual(t *testing.T, got, want int) {
    t.Helper()
    if got != want {
        t.Errorf("got %d, want %d", got, want)
    }
}

// 调用时错误报告指向这里，而非 assertEqual 内部
assertEqual(t, result, 42)
```

### 10.4 Test Cleanup（#1 优先级 — CodeReviewComments）

- 使用 `t.Cleanup` 注册资源清理（不需要在 helper 中要求 caller 自己 defer）
- 这简化了辅助函数的调用模式

```go
func withTempDir(t *testing.T) string {
    t.Helper()
    dir := t.TempDir()  // 自动清理（Go 1.15+）
    // 或手动：
    tmpDir, err := os.MkdirTemp("", "test-*")
    if err != nil {
        t.Fatal(err)
    }
    t.Cleanup(func() { os.RemoveAll(tmpDir) })
    return tmpDir
}
```

### 10.5 Examples（#1 优先级 — CodeReviewComments）

- 为导出的函数和类型编写 Example 测试
- Example 是文档的一部分，同时作为测试运行

```go
func ExampleSplit() {
    got := strings.Split("a,b,c", ",")
    fmt.Println(got)
    // Output: [a b c]
}
```

## 11. 类型与接口

### 11.1 接口大小（#3 优先级 — Google Style Guide）

- 优先小接口（1-3 个方法）
- 接口越大，抽象越弱
- 定义所需行为的**最小集合**

```go
// Good：小接口
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}

// 比单个大接口更好
type ReadWriter interface {
    Reader
    Writer
}
```

### 11.2 Interface Checks — 编译时验证（#1 优先级 — CodeReviewComments）

- 如果类型必须实现特定接口，在代码中添加编译时检查
- 如果没有变量赋值给该接口，Go 会静态检查失败

```go
// 编译时检查：*Server 是否实现了 http.Handler
var _ http.Handler = (*Server)(nil)
```

### 11.3 Struct Embedding（#1 优先级 — CodeReviewComments）

- 嵌入（embedding）是组合，不是继承
- 嵌入应只用于**扩展行为**，而非为了复用字段

```go
// Good：嵌入是为了扩展行为
type BufferedFile struct {
    *os.File
    buf [4096]byte
}
// 自动继承 File 的所有方法

// Bad：只是复用字段
type User struct {
    Name string
    // 嵌入一个只提供字段的类型
    Address  // ❌ —— 应该用字段，而非嵌入
}
```

### 11.4 Generics（#3 优先级 — Google Style Guide + #1 CodeReviewComments）

- Go 1.18+ 支持泛型，但使用要克制
- 优先使用具体类型，除非泛型提供显著的好处
- 适用场景：容器/数据结构、操作 slice/map 的通用函数
- 不适用场景：仅为了"将来可能"而引入泛型

```go
// Good：泛型对通用容器有意义
type Stack[T any] struct {
    items []T
}

func (s *Stack[T]) Push(item T) {
    s.items = append(s.items, item)
}

// ❌ 不必要地使用泛型 —— 使用具体类型更清晰
func PrintAnything[T any](v T) string {
    return fmt.Sprint(v)  // 为什么不直接用 interface{}?
}
```

## 12. 性能与优化

### 12.1 先 Profile，后优化（#1 优先级 — CodeReviewComments）

- 在证明瓶颈存在之前，不要为性能牺牲可读性
- 先实现明显正确的代码，然后用 profile 找瓶颈

### 12.2 String Concatenation（#1 优先级 — CodeReviewComments）

- 在循环中拼接字符串时，使用 `strings.Builder`
- 直接使用 `+=` 会产生大量内存分配

```go
// Good
var b strings.Builder
for _, s := range items {
    b.WriteString(s)
}
result := b.String()

// Bad（大量分配）
result := ""
for _, s := range items {
    result += s
}
```

### 12.3 复用 Slice / Filter Without Allocation（#1 优先级 — CodeReviewComments）

- 过滤 slice 时可以复用原有底层数组，避免额外分配

```go
// Filter 复用 b 的底层数组
func Filter(b []int) []int {
    out := b[:0]
    for _, v := range b {
        if v > 0 {
            out = append(out, v)
        }
    }
    return out
}
```

### 12.4 Range 变量捕获（#1 优先级 — CodeReviewComments）

- Go 1.22+ 中循环变量已改为每次迭代重新绑定，无需担心捕获问题
- 旧版本（Go < 1.22）中需要用局部变量解决闭包捕获：

```go
// Go < 1.22（旧版本）需要这样做
for _, v := range items {
    v := v  // 创建局部副本
    go func() {
        fmt.Println(v)  // 使用局部副本而非循环变量
    }()
}

// Go 1.22+ 中循环变量每次都重新创建，直接使用即可
for _, v := range items {
    go func() {
        fmt.Println(v)  // 正确 —— 每个 goroutine 获取自己的 v
    }()
}
```

### 12.5 传递值 vs 传递指针（#1 优先级 — CodeReviewComments）

- 小值类型（int、bool、小 struct）传值而非指针
- 大结构体、slice、map 传指针
- 猜测"传指针更快"之前先 profile——很多时候传值更快（无 GC 压力）

```go
// 传值
type Point struct { X, Y float64 }
func (p Point) Distance(q Point) float64  // 小结构体传值

// 传指针
type LargeStruct struct { /* 很多字段 */ }
func (s *LargeStruct) Process() error     // 大结构体传指针
```

## 13. 其他约定

### 13.1 不要用 crypto/rand 做伪随机（#1 优先级 — CodeReviewComments）

- 如果不需要加密安全随机，用 `math/rand/v2`（Go 1.22+）或 `math/rand`
- `crypto/rand` 用于密码学场景——它很慢，且是阻塞式的

```go
import "math/rand/v2"

// 非加密随机
n := rand.IntN(100)
```

### 13.2 Copying — Slice & Map（#1 优先级 — CodeReviewComments）

- 从参数接收 slice/map 时，如果需要保留副本，必须深拷贝
- 不要假设调用者不会修改底层数组

```go
// 拷贝 slice
func (s *Store) SetItems(items []string) {
    s.items = make([]string, len(items))
    copy(s.items, items)
}

// 拷贝 map
func (s *Store) SetLabels(labels map[string]string) {
    s.labels = make(map[string]string, len(labels))
    for k, v := range labels {
        s.labels[k] = v
    }
}
```

### 13.3 Line Length（#1 优先级 — CodeReviewComments）

- 实际建议：无严格的硬性限制，但**超过 100 字符应换行**
- 换行时保持可读性，不要让换行本身降低可读性

### 13.4 包级配置标志（#3 优先级 — Google Style Guide）

- 命令行标志在 `main` 包或专用的 `config` 包中定义
- 避免在多个包中分散定义标志
- 标志应有合理的默认值和文档

### 13.5 零值有用性（#2 优先级 — Effective Go）

- Go 的设计原则之一：零值应该直接可用
- 遵循此原则设计自己的类型

```go
// bytes.Buffer 零值可用
var buf bytes.Buffer
buf.WriteString("hello")  // 无需显式初始化

// sync.Mutex 零值可用
var mu sync.Mutex
mu.Lock()

// 不要强制用户调用 Init() 方法——除非必要
type Config struct {
    mu    sync.Mutex
    cache map[string]any
}
// 用户可以直接用，但 map 需要初始化……
// 更好的方式：使用 lazy init 或构造函数
```

### 13.6 不要混合值/指针 receiver（#1 优先级 — CodeReviewComments）

这太重要了，再强调一次：**同一个类型的所有方法应保持 receiver 类型一致。**

```go
// ❌ 不要这样做
type Foo struct { ... }
func (f Foo) Method1() { ... }    // 值 receiver
func (f *Foo) Method2() { ... }   // 指针 receiver（不一致）
```

---

### 工具配置

```yaml
# .golangci.yml 推荐配置
linters:
  enable:
    - gofmt          # 强制 gofmt 格式
    - goimports     # 导入排序
    - errcheck      # 检查未处理的 error
    - govet         # Go vet 静态分析
    - staticcheck   # 更深入的静态分析
    - revive        # 代码风格检查（替代 golint）
```

**常用命令**：

```bash
# 在 CI 中检查
gofmt -l .          # 列出格式错误的文件（非零退出码=失败）
goimports -e -l .   # 检查导入

# 修复
goimports -w .
```

---

> **维护提示**：本文档整合自三份官方 Go 指南。随着 Go 语言版本演进（如泛型在 Go 1.18 引入、循环变量在 Go 1.22 的变更），某些细节可能发生变化。建议定期查阅原文获取最新信息。
>
> **来源链接**：
> - [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
> - [Effective Go](https://go.dev/doc/effective_go)
> - [Google Go Style Guide](https://google.github.io/styleguide/go/decisions)
> - [Google Go Best Practices](https://google.github.io/styleguide/go/best-practices)
