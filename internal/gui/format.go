package gui

import (
	"fmt"
	"math"
	"time"

	"github.com/loveloki/try/internal/selector"
)

// formatSizeKB 将 KB 格式化为人类可读大小。
func formatSizeKB(kb float64) string {
	if kb < 1 {
		return fmt.Sprintf("%dB", int(math.Round(kb*1024)))
	}
	if kb < 1024 {
		return fmt.Sprintf("%.1fKB", kb)
	}
	return fmt.Sprintf("%.1fMB", kb/1024)
}

// formatModTime 将修改时间格式化为相对时间文案。
func formatModTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return selector.FormatTimeAgo(time.Since(t))
}
