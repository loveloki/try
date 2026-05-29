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

// 使用方式
results := fuzzy.Match(entries, query, limit)
// results 按 Score 降序排列
```

### 与 selector 包的类型映射

fuzzy 包定义自己的 Entry/MatchResult 类型，与 selector 包的类型分离。调用方负责转换：

```go
// selector.Entry → fuzzy.Entry（loadAllTries 后构建）
fuzzyEntries := make([]fuzzy.Entry, len(allTries))
for i, e := range allTries {
    fuzzyEntries[i] = fuzzy.Entry{
        Text:      e.Basename,
        BaseScore: e.BaseScore,
        Data:      e,  // 透传原始 selector.Entry
    }
}

// fuzzy.MatchResult → selector.MatchedEntry（refreshList 中转换）
results := fuzzy.Match(fuzzyEntries, query, limit)
matched := make([]MatchedEntry, len(results))
for i, r := range results {
    matched[i] = MatchedEntry{
        Entry:              r.Entry.Data.(Entry),  // 还原透传数据
        Score:              r.Score,
        HighlightPositions: r.Positions,
    }
}
```

## 匹配层（委托 sahilm/fuzzy）

子序列匹配和字符位置追踪委托给 `sahilm/fuzzy` 库：

```go
import "github.com/sahilm/fuzzy"

// 实现 fuzzy.Source 接口适配 Entry 切片
type entrySource []Entry
func (s entrySource) String(i int) string { return s[i].Text }
func (s entrySource) Len() int            { return len(s) }

func Match(entries []Entry, query string, limit int) []MatchResult {
    if query == "" {
        return matchAll(entries, limit)  // 空 query 走纯 base_score 排序
    }

    // sahilm/fuzzy 负责：子序列匹配、大小写不敏感、位置追踪
    matches := fuzzy.FindFrom(query, entrySource(entries))

    // 丢弃库的 Score，用自定义公式重新评分
    results := make([]MatchResult, len(matches))
    for i, m := range matches {
        entry := entries[m.Index]
        score := computeScore(entry.BaseScore, m.MatchedIndexes, entry.Text, query)
        results[i] = MatchResult{
            Entry:     entry,
            Positions: m.MatchedIndexes,
            Score:     score,
        }
    }
    return topK(results, limit)
}
```

sahilm/fuzzy 已作为 bubbles 的间接依赖存在，不增加新依赖。我们只使用其匹配功能，评分和排序完全自定义。

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

匹配字符位于字符串开头或紧跟非字母数字字符时触发。

```
WORD_BOUNDARY_RE = /[^a-z0-9]/

if found == 0 || text[found - 1] matches WORD_BOUNDARY_RE:
    score += 1.0
```

例如 query `c` 匹配 `redis-connection` 中的 `c` 会得到边界加成（因为前面是 `-`）。

#### 4. 邻近加成

```
if lastPos >= 0:
    gap = found - lastPos - 1
    score += 2.0 / sqrt(gap + 1)
```

连续匹配（gap=0）加成最大（+2.0），间隔越远加成越小。

预计算 sqrt 表（gap 0-63）避免热路径中的浮点运算：

```
SQRT_TABLE[i] = 2.0 / sqrt(i + 1)  // i = 0..64
```

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

查询未变化时直接返回缓存结果，避免每次按键都重新计算。在 SelectorModel 层实现：

```go
if m.lastQuery == m.textInput.Value() && m.cachedResults != nil {
    return m.cachedResults
}
```

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
