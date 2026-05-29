# 模糊匹配引擎

## 概述

`internal/fuzzy/` 基于 `sahilm/fuzzy` 库做子序列匹配和位置追踪，自己实现多维评分和 top-k 排序。匹配算法委托给库，评分算法完全自定义。

## 输入接口

```go
type Entry struct {
    Text      string  // 匹配目标文本（目录名）
    BaseScore float64 // 基础分（时间权重 + 日期后缀加成，由调用方预计算）
    Data      any     // 透传数据（调用方自定义，匹配引擎不读取）
}
```

## 输出接口

```go
type MatchResult struct {
    Entry      Entry    // 原始条目（透传）
    Positions  []int    // 匹配字符在文本中的索引位置（用于高亮）
    Score      float64  // 最终综合评分
}

func Match(entries []Entry, query string, limit int) []MatchResult
```

返回结果按 `Score` 降序排列。

### 与 selector 包的类型映射

fuzzy 包定义自己的 `Entry`/`MatchResult` 类型，与 selector 包分离。调用方负责：
- `selector.Entry` → `fuzzy.Entry`（`loadAllTries` 后构建，`Data` 存原始 `selector.Entry`）
- `fuzzy.MatchResult` → `selector.MatchedEntry`（`refreshList` 中转换，从 `Data` 还原）

## 匹配层（委托 sahilm/fuzzy）

```go
type entrySource []Entry
func (s entrySource) String(i int) string
func (s entrySource) Len() int
```

实现 `fuzzy.Source` 接口适配 Entry 切片。`Match` 函数调用 `fuzzy.FindFrom` 获取匹配结果，丢弃库的 Score，用自定义 `computeScore` 重新评分，最后通过 `topK` 排序。

空 query 时走 `matchAll` 分支，所有条目匹配，仅按 `BaseScore` 排序。

sahilm/fuzzy 已作为 bubbles 的间接依赖存在，不增加新依赖。

### 评分公式

最终分数 = `(base_score + match_score) × density × length_penalty`

#### 1. 基础分（base_score）

在目录加载时预计算：

```
hours_since_mod = (now - mtime) / 3600.0
base_score = 3.0 / sqrt(hours_since_mod + 1)    // 修改时间权重（mtime）
base_score += 2.0 if name matches /-\d{4}-\d{2}-\d{2}$/  // 日期后缀加成
```

时间权重衰减曲线：
- 刚访问（0h）：+3.0
- 1 小时前：+2.12
- 24 小时前：+0.60
- 7 天前：+0.23

#### 2. 每字符匹配分（+1.0/字符）

每匹配一个 query 字符，+1.0。

#### 3. 词边界加成（+1.0）

匹配字符位于字符串开头或紧跟非字母数字字符时触发。判断使用 `/[^a-z0-9]/` 正则。

#### 4. 邻近加成

```
gap = found - lastPos - 1
score += 2.0 / sqrt(gap + 1)
```

连续匹配（gap=0）加成最大（+2.0），间隔越远加成越小。

预计算 sqrt 表（gap 0-63）避免热路径中的浮点运算：`SQRT_TABLE[i] = 2.0 / sqrt(i + 1)`

典型值：
- gap=0（连续）：+2.0
- gap=1：+1.41
- gap=3：+1.0
- gap=15：+0.5

#### 5. 密度乘数

```
score *= queryLen / (lastMatchPos + 1)
```

query 长度 / 最后一个匹配位置。匹配越紧凑（都在文本前半部分），乘数越大。

#### 6. 长度惩罚

```
score *= 10.0 / (textLen + 10.0)
```

偏好短名称。20 字符的名称乘数约 0.33，10 字符约 0.5。

### 无 query 时的行为

query 为空时所有条目都匹配，只有 base_score 生效（时间权重 + 日期后缀）。最近访问的目录排在最前面。

## 排序优化

### top-k 部分排序

当设置了 `limit` 且 limit 小于总结果数时，使用 `container/heap` 做 O(n log k) 的部分排序，而非 O(n log n) 全排序。

### 结果缓存

查询未变化时直接返回缓存结果，避免每次按键都重新计算。在 SelectorModel 层实现（比较 `lastQuery` 与当前输入值）。

### limit 动态计算

```go
maxResults := max(height - 6, 3)
```

根据终端高度动态限制结果数，不会计算屏幕外不可见的条目评分。

## 评分示例

假设 query = `rds`，目标 = `redis-server-2025-08-14`（23 字符，mtime 1 小时前）：

```
字符位置：r(0) e(1) d(2) i(3) s(4) -(5) s(6) e(7) r(8) v(9) e(10) r(11) -(12) 2(13) ...

base_score = 3.0/√2 + 2.0 = 4.12
match 'r' at 0: +1.0 (char) +1.0 (边界，字符串开头) = +2.0
match 'd' at 2: +1.0 (char) +1.41 (邻近 gap=1)      = +2.41
match 's' at 4: +1.0 (char) +1.41 (邻近 gap=1)       = +2.41
subtotal = 4.12 + 2.0 + 2.41 + 2.41 = 10.94
density = 3/(4+1) = 0.6
length_penalty = 10/(23+10) = 0.303
final = 10.94 × 0.6 × 0.303 ≈ 1.99
```

名称在日期之前的好处：密度乘数 = 3/(4+1) = 0.6。若日期在前（如 `2025-08-14-redis-server`），密度仅为 3/(15+1) = 0.1875，分数差约 3 倍。
