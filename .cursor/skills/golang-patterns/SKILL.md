---
name: golang-patterns
description: >-
  Go 惯用模式和最佳实践参考手册。涵盖错误处理、并发、接口设计、包组织、
  结构体设计、内存优化等。当用户编写新 Go 代码、审查代码、重构、
  或设计 Go 包/模块时触发。
---

# Go 开发模式

惯用 Go 模式和最佳实践，用于构建健壮、高效、可维护的应用程序。

## 核心原则

### 1. 简洁与清晰

Go 偏好简洁而非花哨。代码应该一目了然、容易阅读。

```go
// 好：清晰直接
func GetUser(id string) (*User, error) {
    user, err := db.FindUser(id)
    if err != nil {
        return nil, fmt.Errorf("获取用户 %s: %w", id, err)
    }
    return user, nil
}
```

### 2. 让零值可用

设计类型时确保零值无需初始化即可使用。

```go
// 好：零值可用
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

### 3. 接受接口，返回结构体

函数应接受接口参数，返回具体类型。

## 错误处理模式

### 带上下文的错误包装

```go
if err != nil {
    return nil, fmt.Errorf("加载配置 %s: %w", path, err)
}
```

### 自定义错误类型

```go
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("%s 验证失败: %s", e.Field, e.Message)
}

var (
    ErrNotFound     = errors.New("资源未找到")
    ErrUnauthorized = errors.New("未授权")
)
```

### 检查错误

使用 `errors.Is` / `errors.As`，不要用 `==` 比较。

## 并发模式

### Worker Pool

```go
func WorkerPool(jobs <-chan Job, results chan<- Result, numWorkers int) {
    var wg sync.WaitGroup
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for job := range jobs {
                results <- process(job)
            }
        }()
    }
    wg.Wait()
    close(results)
}
```

### Context 取消和超时

```go
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()
req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
```

### errgroup 协调

```go
g, ctx := errgroup.WithContext(ctx)
for i, url := range urls {
    i, url := i, url
    g.Go(func() error {
        data, err := FetchWithTimeout(ctx, url)
        if err != nil { return err }
        results[i] = data
        return nil
    })
}
if err := g.Wait(); err != nil { return nil, err }
```

### 避免 goroutine 泄漏

使用带缓冲的 channel + select + `ctx.Done()` 防止阻塞。

## 接口设计

- 单方法接口，按需组合
- 在使用方定义接口，而非提供方
- 通过类型断言实现可选行为

## 包组织

```text
cmd/myapp/main.go       # 入口
internal/handler/       # HTTP 处理器
internal/service/       # 业务逻辑
internal/repository/    # 数据访问
internal/config/        # 配置
```

- 包名：短、小写、无下划线
- 避免包级可变状态，使用依赖注入

## 结构体设计

### 函数选项模式

```go
type Option func(*Server)

func WithTimeout(d time.Duration) Option {
    return func(s *Server) { s.timeout = d }
}

func NewServer(addr string, opts ...Option) *Server {
    s := &Server{addr: addr, timeout: 30 * time.Second}
    for _, opt := range opts { opt(s) }
    return s
}
```

## 内存与性能

- 已知大小时预分配切片：`make([]T, 0, len(items))`
- 频繁分配使用 `sync.Pool`
- 循环中拼接字符串使用 `strings.Builder` 或 `strings.Join`

## 惯用法速查表

| 惯用法 | 说明 |
|--------|------|
| 接受接口，返回结构体 | 函数接受接口参数，返回具体类型 |
| 错误是值 | 将错误视为一等值，而非异常 |
| 不要通过共享内存通信 | 使用 channel 在 goroutine 间协调 |
| 让零值可用 | 类型无需显式初始化即可工作 |
| 清晰优于巧妙 | 可读性优先于精巧 |
| 尽早返回 | 先处理错误，保持正常路径不缩进 |

## 应避免的反模式

- 长函数中的裸返回
- 用 panic 做流程控制
- 在结构体中传递 context（应作为第一个参数）
- 混用值接收器和指针接收器
