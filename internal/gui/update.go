package gui

// resolveSelectorSelection 在列表刷新后恢复光标：有 path 则按 path；否则落在 selected（通常为 0）。
func resolveSelectorSelection(entries []EntryView, selectedPath string, selected int) (int, string) {
	if selectedPath != "" {
		for i, e := range entries {
			if e.Path == selectedPath {
				return i, selectedPath
			}
		}
	}
	if len(entries) == 0 {
		return 0, ""
	}
	if selected < 0 || selected >= len(entries) {
		selected = 0
	}
	return selected, entries[selected].Path
}

// cycleSource 在来源列表中按 delta 循环切换，返回新来源 id。
func cycleSource(sources []string, current string, delta int) string {
	if len(sources) == 0 {
		return current
	}
	idx := 0
	for i, src := range sources {
		if src == current {
			idx = i
			break
		}
	}
	return sources[wrapIndex(idx+delta, len(sources))]
}

// selectorOpenAction 描述选择器 Enter 键应触发的动作。
type selectorOpenAction int

const (
	selectorOpenNone selectorOpenAction = iota
	selectorOpenFiles
	selectorOpenDelete
)

// decideSelectorOpen 根据标记状态与选中项决定 Enter 行为。
func decideSelectorOpen(markedCount, selected, entryCount int) selectorOpenAction {
	if markedCount > 0 {
		return selectorOpenDelete
	}
	if selected < 0 || selected >= entryCount {
		return selectorOpenNone
	}
	return selectorOpenFiles
}
