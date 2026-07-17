import { useCallback, useMemo } from 'react'

interface Selectable {
  id: string
}

interface UseListNavigationOptions<T extends Selectable> {
  items: T[]
  selectedId: string
  onSelect: (id: string) => void
}

export function useListNavigation<T extends Selectable>({
  items,
  selectedId,
  onSelect,
}: UseListNavigationOptions<T>) {
  const selectedIndex = useMemo(
    () => items.findIndex((item) => item.id === selectedId),
    [items, selectedId],
  )

  const selectedItem = useMemo(
    () => items[selectedIndex] ?? null,
    [items, selectedIndex],
  )

  const moveUp = useCallback(() => {
    if (items.length === 0) return
    const nextIndex = selectedIndex > 0 ? selectedIndex - 1 : items.length - 1
    onSelect(items[nextIndex].id)
  }, [items, selectedIndex, onSelect])

  const moveDown = useCallback(() => {
    if (items.length === 0) return
    const nextIndex = selectedIndex < items.length - 1 ? selectedIndex + 1 : 0
    onSelect(items[nextIndex].id)
  }, [items, selectedIndex, onSelect])

  return {
    selectedIndex,
    selectedItem,
    moveUp,
    moveDown,
  }
}

export function toggleMarkedId(markedIds: Set<string>, id: string): Set<string> {
  const next = new Set(markedIds)
  if (next.has(id)) {
    next.delete(id)
  } else {
    next.add(id)
  }
  return next
}
