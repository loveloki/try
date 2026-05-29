---
name: go-build-resolver
description: >-
  Go 构建错误诊断与修复。以最小改动修复编译错误、go vet 告警、
  staticcheck 问题和模块依赖问题。当 go build 失败、go vet 报告问题、
  或模块依赖损坏时触发。
---

# Go 构建错误解决器

## 诊断命令

按顺序运行：

```bash
go build ./...
go vet ./...
staticcheck ./... 2>/dev/null || echo "staticcheck 未安装"
go mod verify
go mod tidy -v
```

## 修复工作流

```
1. go build ./...     -> 解析错误消息
2. 读取受影响文件      -> 理解上下文
3. 应用最小修复        -> 只改必要的
4. go build ./...     -> 验证修复
5. go vet ./...       -> 检查告警
6. go test ./...      -> 确保无副作用
```

## 常见修复模式

| 错误 | 修复 |
|------|------|
| `undefined: X` | 添加导入或修正大小写 |
| `cannot use X as type Y` | 类型转换或解引用 |
| `X does not implement Y` | 用正确的接收器实现方法 |
| `import cycle not allowed` | 将共享类型提取到新包 |
| `cannot find package` | `go get pkg@version` 或 `go mod tidy` |
| `missing return` | 添加 return 语句 |
| `declared but not used` | 删除或使用空白标识符 |
| `multiple-value in single-value context` | `result, err := func()` |
| `cannot assign to struct field in map` | 使用指针 map 或复制-修改-回写 |

## 模块排障

```bash
grep "replace" go.mod              # 检查本地替换
go mod why -m package              # 为什么选择了某个版本
go get package@v1.2.3              # 固定特定版本
go clean -modcache && go mod download  # 修复校验和问题
```

## 修复策略

1. 先修构建错误 → 再修 vet 告警 → 后修 lint 告警
2. 每次改一个，验证每次变更
3. **只做最小修复** —— 不重构，只修复错误
4. 始终在添加/删除导入后运行 `go mod tidy`
5. 修复根本原因而非压制症状

## 停止条件

- 3 次修复尝试后同一错误仍存在
- 修复引入了比解决的更多的错误
- 错误需要超出范围的架构变更
