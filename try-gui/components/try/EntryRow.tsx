'use client'

import { TryEntry, relativeTime, formatSize, fuzzyMatch } from '@/lib/try-types'
import { MatchText } from './MatchText'
import { Trash2, RotateCcw, FolderOpen } from 'lucide-react'
import { cn } from '@/lib/utils'

interface EntryRowProps {
  entry: TryEntry
  isSelected: boolean
  isMarked: boolean
  deleteMode: boolean
  searchQuery: string
  onClick: () => void
  onOpen: () => void
  onToggleMark: () => void
  index: number
}

export function EntryRow({
  entry,
  isSelected,
  isMarked,
  deleteMode,
  searchQuery,
  onClick,
  onOpen,
  onToggleMark,
  index,
}: EntryRowProps) {
  // Don't show entries that don't match the fuzzy query
  if (searchQuery && !fuzzyMatch(entry.baseName, searchQuery) && !fuzzyMatch(entry.date, searchQuery)) {
    return null
  }

  const timeStr = relativeTime(entry.lastModified)
  const sizeStr = formatSize(entry.sizeKB)

  return (
    <div
      onClick={onClick}
      onDoubleClick={onOpen}
      className={cn(
        'group relative flex items-center gap-3 px-4 py-2.5 cursor-pointer select-none transition-colors border-b border-[var(--try-line)]',
        isMarked
          ? 'bg-[var(--try-danger)]/8 hover:bg-[var(--try-danger)]/12'
          : isSelected
          ? 'bg-[var(--try-surface-selected)]'
          : 'hover:bg-[var(--try-surface-hover)]',
      )}
      role="option"
      aria-selected={isSelected}
    >
      {/* Selection arrow */}
      <div className="flex-none w-3 flex items-center justify-center">
        {isSelected && !isMarked && (
          <span className="text-[var(--try-blue)] text-xs font-bold">›</span>
        )}
        {isMarked && (
          <span className="text-[var(--try-danger)] text-xs font-bold">✕</span>
        )}
      </div>

      {/* Icon */}
      <div className="flex-none">
        <FolderOpen
          size={15}
          className={cn(
            'transition-colors',
            isMarked
              ? 'text-[var(--try-danger)]'
              : isSelected
              ? 'text-[var(--try-blue)]'
              : 'text-muted-foreground group-hover:text-foreground',
          )}
        />
      </div>

      {/* Name + date */}
      <div className="flex-1 min-w-0">
        <div className="flex items-baseline gap-2 min-w-0">
          <MatchText
            text={entry.baseName}
            query={searchQuery}
            className={cn(
              'font-mono text-sm font-medium truncate',
              isMarked
                ? 'text-[var(--try-danger)] line-through opacity-70'
                : isSelected
                ? 'text-foreground font-semibold'
                : 'text-foreground',
            )}
          />
          <span className="flex-none font-mono text-[11px] text-muted-foreground/70">
            <MatchText text={entry.date} query={searchQuery} />
          </span>
        </div>

        {/* Tags */}
        {entry.tags && entry.tags.length > 0 && (
          <div className="flex items-center gap-1 mt-0.5">
            {entry.tags.map((tag) => (
              <span
                key={tag}
                className="text-[10px] px-1.5 py-0 rounded font-mono bg-[var(--try-surface-hover)] text-muted-foreground border border-border/50"
              >
                {tag}
              </span>
            ))}
          </div>
        )}
      </div>

      {/* Metadata right side */}
      <div className="flex-none flex flex-col items-end gap-0.5">
        {/* Score bar */}
        <div className="w-12 score-bar">
          <div className="score-bar-fill" style={{ width: `${entry.score}%` }} />
        </div>
        <div className="flex items-center gap-2">
          <span className="font-mono text-[11px] text-muted-foreground">{sizeStr}</span>
          <span className="font-mono text-[11px] text-muted-foreground">{timeStr}</span>
        </div>
      </div>

      {/* Actions (hover) */}
      <div
        className={cn(
          'flex-none flex items-center gap-1 ml-1',
          deleteMode ? 'opacity-100' : 'opacity-0 group-hover:opacity-100 transition-opacity',
        )}
      >
        {deleteMode && (
          <button
            onClick={(e) => {
              e.stopPropagation()
              onToggleMark()
            }}
            className={cn(
              'flex items-center justify-center w-6 h-6 rounded-md transition-colors',
              isMarked
                ? 'text-[var(--try-danger)] bg-[var(--try-danger)]/15 hover:bg-[var(--try-danger)]/25'
                : 'text-muted-foreground hover:text-[var(--try-danger)] hover:bg-[var(--try-danger)]/10',
            )}
            aria-label={isMarked ? '取消标记删除' : '标记删除'}
          >
            {isMarked ? <RotateCcw size={11} /> : <Trash2 size={11} />}
          </button>
        )}
        {!deleteMode && (
          <button
            onClick={(e) => {
              e.stopPropagation()
              onOpen()
            }}
            className="flex items-center justify-center w-6 h-6 rounded-md text-muted-foreground hover:text-[var(--try-blue)] hover:bg-[var(--try-blue-dim)] transition-colors"
            aria-label="打开目录"
          >
            <FolderOpen size={11} />
          </button>
        )}
      </div>
    </div>
  )
}
