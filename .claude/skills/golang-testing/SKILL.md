---
name: golang-testing
description: >-
  Go 测试模式和 TDD 工作流参考手册。涵盖表驱动测试、基准测试、
  模糊测试、Mock、Golden Files、覆盖率分析等。当用户编写测试、
  为已有代码添加覆盖、或遵循 TDD 流程时触发。
---

# Go 测试模式

## TDD 工作流

```
RED     → 先写一个失败的测试
GREEN   → 写最少的代码使测试通过
REFACTOR → 改进代码，保持测试通过
REPEAT  → 继续下一个需求
```

## 表驱动测试

```go
func TestParseConfig(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    *Config
        wantErr bool
    }{
        {"有效配置", `{"host":"localhost"}`, &Config{Host: "localhost"}, false},
        {"无效 JSON", `{invalid}`, nil, true},
        {"空输入", "", nil, true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseConfig(tt.input)
            if tt.wantErr {
                if err == nil { t.Error("期望错误，得到 nil") }
                return
            }
            if err != nil { t.Fatalf("意外错误: %v", err) }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("got %+v; want %+v", got, tt.want)
            }
        })
    }
}
```

## 测试辅助函数

```go
func setupTestDB(t *testing.T) *sql.DB {
    t.Helper()
    db, err := sql.Open("sqlite3", ":memory:")
    if err != nil { t.Fatalf("打开数据库失败: %v", err) }
    t.Cleanup(func() { db.Close() })
    return db
}
```

## 并行子测试

```go
for _, tt := range tests {
    tt := tt
    t.Run(tt.name, func(t *testing.T) {
        t.Parallel()
        // 测试体
    })
}
```

## Golden Files

使用 `testdata/` 存储预期输出，`-update` 标志更新：

```go
var update = flag.Bool("update", false, "更新黄金文件")
// ... 比对 got 与 golden 文件内容
```

## 基于接口的 Mock

```go
type MockUserRepo struct {
    GetUserFunc func(id string) (*User, error)
}
func (m *MockUserRepo) GetUser(id string) (*User, error) {
    return m.GetUserFunc(id)
}
```

## 基准测试

```go
func BenchmarkProcess(b *testing.B) {
    data := generateTestData(1000)
    b.ResetTimer()
    for i := 0; i < b.N; i++ { Process(data) }
}
// 运行: go test -bench=BenchmarkProcess -benchmem
```

## 模糊测试

```go
func FuzzParseJSON(f *testing.F) {
    f.Add(`{"name": "test"}`)
    f.Fuzz(func(t *testing.T, input string) {
        var result map[string]interface{}
        if err := json.Unmarshal([]byte(input), &result); err != nil {
            return
        }
        if _, err := json.Marshal(result); err != nil {
            t.Errorf("Unmarshal 成功后 Marshal 失败: %v", err)
        }
    })
}
// 运行: go test -fuzz=FuzzParseJSON -fuzztime=30s
```

## 覆盖率命令

```bash
go test -cover ./...                        # 基础覆盖率
go test -coverprofile=coverage.out ./...    # 生成文件
go tool cover -html=coverage.out            # 浏览器查看
go test -race -cover ./...                  # 带竞态检测
```

## 最佳实践

**应该做的：**
- 先写测试（TDD），测试行为而非实现
- 使用 `t.Helper()` / `t.Cleanup()` / `t.Parallel()`
- 使用 `t.TempDir()` 和 `t.Setenv()` 隔离环境

**不应该做的：**
- 直接测试私有函数、在测试中使用 `time.Sleep`
- 忽视不稳定的测试、过度 mock
