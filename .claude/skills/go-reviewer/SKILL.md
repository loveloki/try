---
name: go-reviewer
description: >-
  Go 代码审查，按优先级检查安全、错误处理、并发、代码质量和性能。
  当用户请求代码审查、提交前检查、或审查 PR 时触发。
---

# Go 代码审查器

## 审查优先级

### 严重 —— 安全
- SQL/命令注入：查询/exec 中的字符串拼接
- 路径遍历：未经 `filepath.Clean` + 前缀检查的用户路径
- 竞态条件：共享状态无同步
- 硬编码密钥、不安全 TLS

### 严重 —— 错误处理
- 使用 `_` 丢弃错误
- `return err` 缺少 `fmt.Errorf("上下文: %w", err)`
- 可恢复错误使用 panic
- `err == target` 代替 `errors.Is`

### 高 —— 并发
- goroutine 泄漏（无 context 取消）
- 无缓冲 channel 死锁
- 缺少 sync.WaitGroup
- 未 defer mu.Unlock()

### 高 —— 代码质量
- 函数超过 50 行、嵌套超过 4 层
- if/else 代替提前返回
- 包级可变全局状态
- 接口膨胀

### 中 —— 性能
- 循环中字符串拼接（用 `strings.Builder`）
- 切片未预分配
- 热路径中不必要的分配

### 中 —— 最佳实践
- Context 不是第一个参数
- 测试未使用表驱动模式
- 包命名不规范
- 循环中使用 defer

## 诊断命令

```bash
go vet ./...
staticcheck ./...
go build -race ./...
go test -race ./...
```

## 审批标准

| 状态 | 条件 |
|------|------|
| 通过 | 无严重或高优问题 |
| 警告 | 仅有中优问题 |
| 阻止 | 发现严重或高优问题 |
