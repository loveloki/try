'use client'

import { useState, useEffect, useCallback, useRef } from 'react'
import { TryEntry, MOCK_ENTRIES, SourceTab, fuzzyMatch } from '@/lib/try-types'
import { TryHeader } from './TryHeader'
import { SourceTabs } from './SourceTabs'
import { EntryRow } from './EntryRow'
import { SelectorFooter } from './SelectorFooter'
import { DeleteDialog } from './DeleteDialog'
import { FolderSearch } from 'lucide-react'

interface SelectorViewProps {
  theme: 'dark' | 'light'
  onThemeToggle: () => void
  onOpen: (entry: TryEntry) => void
}

export function SelectorView({ theme, onThemeToggle, onOpen }: SelectorViewProps) {
  const [entries, setEntries] = useState<TryEntry[]>(MOCK_ENTRIES)
  const [searchQuery, setSearchQuery] = useState('')
  const [activeTab, setActiveTab] = useState<SourceTab>('all')
  const [selectedId, setSelectedId] = useState<string>(MOCK_ENTRIES[0]?.id ?? '')
  const [deleteMode, setDeleteMode] = useState(false)
  const [markedIds, setMarkedIds] = useState<Set<string>>(new Set())
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)
  const [showCreateInput, setShowCreateInput] = useState(false)
  const [createName, setCreateName] = useState('')
  const listRef = useRef<HTMLDivElement>(null)

  // Filter entries
  const filtered = entries.filter((e) => {
    const tabOk = activeTab === 'all' || e.source === activeTab
    if (!tabOk) return false
    if (!searchQuery) return true
    return !!fuzzyMatch(e.baseName, searchQuery) || !!fuzzyMatch(e.date, searchQuery)
  })

  const counts: Record<SourceTab, number> = {
    all: entries.length,
    tries: entries.filter((e) => e.source === 'tries').length,
    ship: entries.filter((e) => e.source === 'ship').length,
    bug: entries.filter((e) => e.source === 'bug').length,
  }

  // Keyboard navigation
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (showDeleteConfirm) {
        if (e.key === 'Escape') setShowDeleteConfirm(false)
        if (e.key === 'y' || e.key === 'Y' || e.key === 'Enter') handleDeleteConfirm()
        return
      }
      if (showCreateInput) return

      switch (e.key) {
        case 'ArrowDown': {
          e.preventDefault()
          const idx = filtered.findIndex((x) => x.id === selectedId)
          if (idx < filtered.length - 1) setSelectedId(filtered[idx + 1].id)
          break
        }
        case 'ArrowUp': {
          e.preventDefault()
          const idx = filtered.findIndex((x) => x.id === selectedId)
          if (idx > 0) setSelectedId(filtered[idx - 1].id)
          break
        }
        case 'Enter': {
          if (deleteMode && markedIds.size > 0) {
            setShowDeleteConfirm(true)
          } else {
            const sel = filtered.find((x) => x.id === selectedId)
            if (sel) onOpen(sel)
          }
          break
        }
        case 'Escape': {
          if (deleteMode) {
            setDeleteMode(false)
            setMarkedIds(new Set())
          }
          break
        }
        case 'Tab': {
          e.preventDefault()
          const tabs: SourceTab[] = e.shiftKey
            ? ['all', 'bug', 'ship', 'tries']
            : ['all', 'tries', 'ship', 'bug']
          const idx = tabs.indexOf(activeTab)
          setActiveTab(tabs[(idx + 1) % tabs.length])
          break
        }
        default:
          if (e.ctrlKey && e.key === 't') {
            e.preventDefault()
            setShowCreateInput((v) => !v)
          }
          if (e.ctrlKey && e.key === 'd') {
            e.preventDefault()
            setDeleteMode((v) => !v)
            if (deleteMode) setMarkedIds(new Set())
          }
      }
    }
    window.addEventListener('keydown', handler)
    return () => window.removeEventListener('keydown', handler)
  }, [filtered, selectedId, deleteMode, markedIds, showDeleteConfirm, showCreateInput, activeTab, onOpen])

  const handleDeleteConfirm = useCallback(() => {
    setEntries((prev) => prev.filter((e) => !markedIds.has(e.id)))
    setMarkedIds(new Set())
    setDeleteMode(false)
    setShowDeleteConfirm(false)
  }, [markedIds])

  const handleCreate = useCallback(() => {
    if (!createName.trim()) return
    const today = new Date().toISOString().slice(0, 10)
    const newEntry: TryEntry = {
      id: Date.now().toString(),
      name: `${createName.trim()}-${today}`,
      baseName: createName.trim(),
      date: today,
      source: 'tries',
      score: 100,
      lastModified: new Date(),
      fileCount: 0,
      sizeKB: 0,
      tags: [],
    }
    setEntries((prev) => [newEntry, ...prev])
    setSelectedId(newEntry.id)
    setCreateName('')
    setShowCreateInput(false)
  }, [createName])

  const markedEntries = entries.filter((e) => markedIds.has(e.id))

  return (
    <div className="flex flex-col h-full">
      <TryHeader
        title="try"
        subtitle="~/src/tries"
        searchQuery={searchQuery}
        onSearchChange={setSearchQuery}
        searchPlaceholder="模糊搜索目录…"
        theme={theme}
        onThemeToggle={onThemeToggle}
        showSearch
      />

      <SourceTabs active={activeTab} counts={counts} onChange={setActiveTab} />

      {/* Entry list */}
      <div
        ref={listRef}
        className="flex-1 overflow-y-auto"
        role="listbox"
        aria-label="实验目录列表"
      >
        {filtered.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full gap-3 text-muted-foreground">
            <FolderSearch size={32} className="opacity-30" />
            <div className="text-center">
              <p className="text-sm font-medium">
                {searchQuery ? `无匹配结果 "${searchQuery}"` : '暂无目录'}
              </p>
              {searchQuery && (
                <p className="text-xs mt-1 font-mono text-[var(--try-accent)]">
                  ⌃T 创建"{searchQuery}"
                </p>
              )}
            </div>
          </div>
        ) : (
          filtered.map((entry, idx) => (
            <EntryRow
              key={entry.id}
              entry={entry}
              isSelected={entry.id === selectedId}
              isMarked={markedIds.has(entry.id)}
              deleteMode={deleteMode}
              searchQuery={searchQuery}
              index={idx}
              onClick={() => setSelectedId(entry.id)}
              onOpen={() => onOpen(entry)}
              onToggleMark={() => {
                setMarkedIds((prev) => {
                  const next = new Set(prev)
                  next.has(entry.id) ? next.delete(entry.id) : next.add(entry.id)
                  return next
                })
              }}
            />
          ))
        )}
      </div>

      <SelectorFooter
        totalCount={entries.length}
        filteredCount={filtered.length}
        deleteMode={deleteMode}
        markedCount={markedIds.size}
        showCreateInput={showCreateInput}
        createName={createName}
        onCreateNameChange={setCreateName}
        onCreateSubmit={handleCreate}
        onDeleteModeToggle={() => {
          setDeleteMode(false)
          setMarkedIds(new Set())
        }}
        onDeleteConfirm={() => setShowDeleteConfirm(true)}
      />

      {showDeleteConfirm && (
        <DeleteDialog
          entries={markedEntries}
          onConfirm={handleDeleteConfirm}
          onCancel={() => setShowDeleteConfirm(false)}
        />
      )}
    </div>
  )
}
