package fuzzy

import (
	"container/heap"
	"math"
	"sort"

	sfuzzy "github.com/sahilm/fuzzy"
)

// Entry 模糊匹配的输入条目
type Entry struct {
	Text      string  // 匹配目标文本（目录名）
	BaseScore float64 // 基础分（时间权重 + 日期后缀加成，由调用方预计算）
	Data      any     // 透传数据（匹配引擎不读取）
}

// MatchResult 匹配结果
type MatchResult struct {
	Entry     Entry
	Positions []int   // 匹配字符在文本中的索引位置（用于高亮）
	Score     float64 // 最终综合评分
}

// 预计算 sqrt 查找表，避免热路径中的浮点运算
var sqrtTable [65]float64

func init() {
	for i := range sqrtTable {
		sqrtTable[i] = 2.0 / math.Sqrt(float64(i)+1)
	}
}

// 词边界正则的等效判断：非字母数字字符
func isWordBoundary(c byte) bool {
	return !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9'))
}

// entrySource 实现 sahilm/fuzzy.Source 接口
type entrySource []Entry

func (s entrySource) String(i int) string { return s[i].Text }
func (s entrySource) Len() int            { return len(s) }

// Match 执行模糊匹配：空 query 返回按 BaseScore 排序的全部条目；
// 非空 query 委托 sahilm/fuzzy 做子序列匹配，用自定义公式评分后 top-k 排序。
func Match(entries []Entry, query string, limit int) []MatchResult {
	if query == "" {
		return matchAll(entries, limit)
	}

	matches := sfuzzy.FindFrom(query, entrySource(entries))

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

// matchAll 空 query 时按 BaseScore 降序返回所有条目
func matchAll(entries []Entry, limit int) []MatchResult {
	results := make([]MatchResult, len(entries))
	for i, e := range entries {
		results[i] = MatchResult{
			Entry:     e,
			Positions: nil,
			Score:     e.BaseScore,
		}
	}
	return topK(results, limit)
}

// computeScore 多维评分公式：(base_score + match_score) × density × length_penalty
func computeScore(baseScore float64, positions []int, text, query string) float64 {
	if len(positions) == 0 {
		return baseScore
	}

	score := baseScore
	queryLen := float64(len(query))
	textLen := float64(len(text))
	lastPos := -1

	for _, found := range positions {
		// 每字符匹配分
		score += 1.0

		// 词边界加成
		if found == 0 || isWordBoundary(text[found-1]) {
			score += 1.0
		}

		// 邻近加成
		if lastPos >= 0 {
			gap := found - lastPos - 1
			if gap < len(sqrtTable) {
				score += sqrtTable[gap]
			} else {
				score += 2.0 / math.Sqrt(float64(gap)+1)
			}
		}
		lastPos = found
	}

	// 密度乘数：query 长度 / (最后匹配位置 + 1)
	if lastPos >= 0 {
		score *= queryLen / float64(lastPos+1)
	}

	// 长度惩罚：偏好短名称
	score *= 10.0 / (textLen + 10.0)

	return score
}

// --- top-k 部分排序 ---

// topK 返回按 Score 降序的前 limit 个结果。
// limit <= 0 或 >= len(results) 时返回全部。
func topK(results []MatchResult, limit int) []MatchResult {
	if limit <= 0 || limit >= len(results) {
		sort.Slice(results, func(i, j int) bool {
			return results[i].Score > results[j].Score
		})
		return results
	}

	// 使用 min-heap 维护 top-k：O(n log k)
	h := &minHeap{}
	heap.Init(h)
	for _, r := range results {
		if h.Len() < limit {
			heap.Push(h, r)
		} else if r.Score > (*h)[0].Score {
			(*h)[0] = r
			heap.Fix(h, 0)
		}
	}

	out := make([]MatchResult, h.Len())
	for i := len(out) - 1; i >= 0; i-- {
		out[i] = heap.Pop(h).(MatchResult)
	}
	return out
}

// minHeap 按 Score 升序排列的最小堆
type minHeap []MatchResult

func (h minHeap) Len() int            { return len(h) }
func (h minHeap) Less(i, j int) bool  { return h[i].Score < h[j].Score }
func (h minHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *minHeap) Push(x any)         { *h = append(*h, x.(MatchResult)) }
func (h *minHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}
