'use client'

import { useState, useEffect, useCallback, useRef } from 'react'
import { TryEntry, FileEntry, MOCK_FILES, relativeTime, formatSize } from '@/lib/try-types'
import { TryHeader } from './TryHeader'
import { FileIcon } from './FileIcon'
import { cn } from '@/lib/utils'
import {
  ArrowLeft,
  Upload,
  Trash2,
  Edit3,
  Package,
  PackageOpen,
  FolderOpen,
  Plus,
  RotateCcw,
  AlertTriangle,
} from 'lucide-react'

interface FilesViewProps {
  entry: TryEntry
  theme: 'dark' | 'light'
  onThemeToggle: () => void
  onBack: () => void
}

export function FilesView({ entry, theme, onThemeToggle, onBack }: FilesViewProps) {
  const [files, setFiles] = useState<FileEntry[]>(MOCK_FILES)
  const [selectedId, setSelectedId] = useState<string>(MOCK_FILES[0]?.id ?? '')
  const [searchQuery, setSearchQuery] = useState('')
  const [dropActive, setDropActive] = useState(false)
  const [markedIds, setMarkedIds] = useState<Set<string>>(new Set())
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)
  const [notification, setNotification] = useState<string | null>(null)
  const dropRef = useRef<HTMLDivElement>(null)

  const notify = (msg: string) => {
    setNotification(msg)
    setTimeout(() => setNotification(null), 2000)
  }

  const filtered = files.filter((f) =>
    searchQuery ? f.name.toLowerCase().includes(searchQuery.toLowerCase()) : true,
  )

  // Keyboard navigation
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (showDeleteConfirm) {
        if (e.key === 'Escape') setShowDeleteConfirm(false)
        if (e.key === 'y' || e.key === 'Y' || e.key === 'Enter') handleDeleteConfirm()
        return
      }
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
        case 'Escape': {
          if (!searchQuery) onBack()
          break
        }
      }
    }
    window.addEventListener('keydown', handler)
    return () => window.removeEventListener('keydown', handler)
  }, [filtered, selectedId, showDeleteConfirm, searchQuery, onBack])

  // Drag & Drop
  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    setDropActive(true)
  }, [])
  const handleDragLeave = useCallback(() => setDropActive(false), [])
  const handleDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault()
      setDropActive(false)
      const droppedFiles = Array.from(e.dataTransfer.files)
      if (droppedFiles.length === 0) return
      const newEntries: FileEntry[] = droppedFiles.map((f) => ({
        id: Date.now().toString() + Math.random(),
        name: f.name,
        type: 'unknown' as const,
        sizeKB: f.size / 1024,
        modified: new Date(),
        isDir: false,
      }))
      setFiles((prev) => [...newEntries, ...prev])
      notify(`已上传 ${droppedFiles.length} 个文件`)
    },
    [],
  )

  const handleDeleteConfirm = useCallback(() => {
    setFiles((prev) => prev.filter((f) => !markedIds.has(f.id)))
    setMarkedIds(new Set())
    setShowDeleteConfirm(false)
    notify('已删除')
  }, [markedIds])

  const handleDocxOp = (op: 'pack' | 'unpack') => {
    notify(op === 'pack' ? '已打包为 .docx（模拟）' : '已解压 .docx（模拟）')
  }

  const markedFiles = files.filter((f) => markedIds.has(f.id))
  const selectedFile = files.find((f) => f.id === selectedId)

  return (
    <div
      ref={dropRef}
      className={cn('flex flex-col h-full transition-all', dropActive && 'drop-active')}
      onDragOver={handleDragOver}
      onDragLeave={handleDragLeave}
      onDrop={handleDrop}
    >
      <TryHeader
        title={entry.baseName}
        subtitle={entry.name}
        searchQuery={searchQuery}
        onSearchChange={setSearchQuery}
        searchPlaceholder="过滤文件…"
        theme={theme}
        onThemeToggle={onThemeToggle}
        showSearch
      />

      {/* Toolbar */}
      <div className="flex-none flex items-center gap-1.5 px-4 py-2 border-b border-[var(--try-line)] bg-background overflow-x-auto">
        <button
          onClick={onBack}
          className="flex items-center gap-1.5 px-2.5 py-1.5 rounded-md text-xs font-mono text-muted-foreground hover:text-foreground hover:bg-[var(--try-surface-hover)] transition-colors border border-transparent hover:border-border/50 flex-none"
        >
          <ArrowLeft size={12} />
          返回
        </button>

        <div className="w-px h-4 bg-border mx-1 flex-none" />

        {/* File actions */}
        {[
          {
            icon: Upload,
            label: '上传',
            color: 'text-[var(--try-blue)]',
            bg: 'hover:bg-[var(--try-blue-dim)]',
            action: () => notify('拖拽文件到此窗口即可上传'),
          },
          {
            icon: Edit3,
            label: '编辑',
            color: 'text-[var(--try-accent)]',
            bg: 'hover:bg-[var(--try-accent)]/10',
            action: () => notify(`已在系统编辑器中打开 ${selectedFile?.name ?? ''}`),
          },
          {
            icon: Package,
            label: '打包 .docx',
            color: 'text-muted-foreground',
            bg: 'hover:bg-[var(--try-surface-hover)]',
            action: () => handleDocxOp('pack'),
          },
          {
            icon: PackageOpen,
            label: '解压 .docx',
            color: 'text-muted-foreground',
            bg: 'hover:bg-[var(--try-surface-hover)]',
            action: () => handleDocxOp('unpack'),
          },
        ].map(({ icon: Icon, label, color, bg, action }) => (
          <button
            key={label}
            onClick={action}
            className={cn(
              'flex items-center gap-1.5 px-2.5 py-1.5 rounded-md text-xs font-mono transition-colors border border-transparent flex-none',
              color,
              bg,
            )}
          >
            <Icon size={12} />
            <span className="hidden sm:inline">{label}</span>
          </button>
        ))}

        <div className="w-px h-4 bg-border mx-1 flex-none" />

        <button
          onClick={() => {
            if (markedIds.size > 0) {
              setShowDeleteConfirm(true)
            } else {
              notify('先选中要删除的文件（单击选中，再点删除）')
            }
          }}
          className={cn(
            'flex items-center gap-1.5 px-2.5 py-1.5 rounded-md text-xs font-mono transition-colors border flex-none',
            markedIds.size > 0
              ? 'text-[var(--try-danger)] border-[var(--try-danger)]/30 hover:bg-[var(--try-danger)]/10'
              : 'text-muted-foreground border-transparent hover:bg-[var(--try-surface-hover)]',
          )}
        >
          <Trash2 size={12} />
          <span className="hidden sm:inline">删除{markedIds.size > 0 ? ` (${markedIds.size})` : ''}</span>
        </button>
      </div>

      {/* Drop overlay hint */}
      {dropActive && (
        <div className="absolute inset-0 z-10 flex flex-col items-center justify-center pointer-events-none">
          <div className="flex flex-col items-center gap-3 p-8 rounded-2xl bg-background/90 border-2 border-dashed border-[var(--try-blue)] shadow-2xl">
            <Upload size={32} className="text-[var(--try-blue)]" />
            <p className="text-sm font-semibold text-[var(--try-blue)]">松开以上传文件</p>
            <p className="text-xs text-muted-foreground">文件将复制到 {entry.name}</p>
          </div>
        </div>
      )}

      {/* File list */}
      <div className="flex-1 overflow-y-auto" role="listbox" aria-label="文件列表">
        {/* Column header */}
        <div className="flex items-center px-4 py-1.5 border-b border-[var(--try-line)] sticky top-0 bg-background/95 backdrop-blur-sm z-10">
          <div className="w-3 flex-none mr-3" />
          <div className="w-4 flex-none mr-3" />
          <span className="flex-1 min-w-0 text-[10px] font-mono text-muted-foreground/70 uppercase tracking-wider">
            名称
          </span>
          <span className="w-20 flex-none text-[10px] font-mono text-muted-foreground/70 uppercase tracking-wider text-right">
            大小
          </span>
          <span className="w-20 flex-none text-[10px] font-mono text-muted-foreground/70 uppercase tracking-wider text-right hidden sm:block">
            修改时间
          </span>
          <div className="w-12 flex-none" />
        </div>

        {filtered.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-16 gap-2 text-muted-foreground">
            <FolderOpen size={28} className="opacity-30" />
            <p className="text-sm">{searchQuery ? `无匹配文件` : '目录为空'}</p>
          </div>
        ) : (
          filtered.map((file) => (
            <FileRow
              key={file.id}
              file={file}
              isSelected={file.id === selectedId}
              isMarked={markedIds.has(file.id)}
              onClick={() => {
                setSelectedId(file.id)
                setMarkedIds((prev) => {
                  const next = new Set(prev)
                  if (next.has(file.id)) {
                    next.delete(file.id)
                  } else {
                    next.add(file.id)
                  }
                  return next
                })
              }}
            />
          ))
        )}
      </div>

      {/* Footer status */}
      <footer className="flex-none flex items-center justify-between px-4 py-2 border-t border-border bg-background text-[11px] font-mono text-muted-foreground">
        <span>
          {filtered.length} 个项目
          {markedIds.size > 0 && (
            <span className="ml-2 text-[var(--try-danger)]">· 已选 {markedIds.size} 项</span>
          )}
        </span>
        <div className="flex items-center gap-3">
          <span className="hidden sm:block">
            <kbd className="border border-border/40 rounded px-1">↑↓</kbd> 导航
          </span>
          <span className="hidden sm:block">
            <kbd className="border border-border/40 rounded px-1">Esc</kbd> 返回
          </span>
          <span className="text-[var(--try-blue)]/60">拖拽上传</span>
        </div>
      </footer>

      {/* Delete confirmation */}
      {showDeleteConfirm && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
          <div className="w-72 rounded-xl border border-[var(--try-danger)]/30 bg-card shadow-2xl">
            <div className="flex items-center gap-2.5 px-5 py-4 border-b border-[var(--try-danger)]/20">
              <AlertTriangle size={14} className="text-[var(--try-danger)]" />
              <h2 className="text-sm font-semibold text-[var(--try-danger)]">
                删除 {markedFiles.length} 个文件？
              </h2>
            </div>
            <div className="px-5 py-3 max-h-40 overflow-y-auto">
              {markedFiles.map((f) => (
                <div key={f.id} className="flex items-center gap-2 py-1.5 border-b border-[var(--try-line)] last:border-0">
                  <FileIcon type={f.type} size={12} />
                  <span className="font-mono text-xs text-[var(--try-danger)] line-through opacity-80">
                    {f.name}
                  </span>
                </div>
              ))}
            </div>
            <div className="flex gap-2 px-5 py-4 border-t border-[var(--try-line)]">
              <button
                onClick={() => setShowDeleteConfirm(false)}
                className="flex-1 h-8 rounded-md text-xs font-mono border border-border text-muted-foreground hover:text-foreground hover:bg-[var(--try-surface-hover)] transition-colors"
              >
                取消
              </button>
              <button
                onClick={handleDeleteConfirm}
                className="flex-1 h-8 rounded-md text-xs font-mono font-semibold bg-[var(--try-danger)] text-white hover:opacity-90 transition-opacity"
              >
                确认删除
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Notification toast */}
      {notification && (
        <div className="fixed bottom-4 left-1/2 -translate-x-1/2 z-50">
          <div className="px-4 py-2 rounded-lg bg-card border border-border shadow-xl text-xs font-mono text-foreground">
            {notification}
          </div>
        </div>
      )}
    </div>
  )
}

// ─── File Row ─────────────────────────────────────────

function FileRow({
  file,
  isSelected,
  isMarked,
  onClick,
}: {
  file: FileEntry
  isSelected: boolean
  isMarked: boolean
  onClick: () => void
}) {
  return (
    <div
      onClick={onClick}
      className={cn(
        'group flex items-center px-4 py-2 cursor-pointer select-none transition-colors border-b border-[var(--try-line)]',
        isMarked
          ? 'bg-[var(--try-danger)]/8 hover:bg-[var(--try-danger)]/12'
          : isSelected
          ? 'bg-[var(--try-surface-selected)]'
          : 'hover:bg-[var(--try-surface-hover)]',
      )}
      role="option"
      aria-selected={isSelected}
    >
      {/* Checkbox indicator */}
      <div className="flex-none w-3 mr-3 flex items-center justify-center">
        {isMarked ? (
          <div className="w-3 h-3 rounded-sm bg-[var(--try-danger)] flex items-center justify-center">
            <span className="text-white text-[8px] leading-none">✓</span>
          </div>
        ) : (
          <div className="w-3 h-3 rounded-sm border border-border/50 group-hover:border-border transition-colors" />
        )}
      </div>

      {/* File icon */}
      <div className="flex-none mr-3">
        <FileIcon type={file.type} size={15} />
      </div>

      {/* Name */}
      <div className="flex-1 min-w-0">
        <span
          className={cn(
            'font-mono text-sm truncate block',
            isMarked
              ? 'text-[var(--try-danger)] line-through opacity-70'
              : isSelected
              ? 'text-foreground font-medium'
              : 'text-foreground',
          )}
        >
          {file.name}
        </span>
      </div>

      {/* Size */}
      <div className="w-20 flex-none text-right">
        <span className="font-mono text-xs text-muted-foreground">
          {file.isDir ? '—' : formatSize(file.sizeKB)}
        </span>
      </div>

      {/* Modified */}
      <div className="w-20 flex-none text-right hidden sm:block">
        <span className="font-mono text-xs text-muted-foreground">
          {relativeTime(file.modified)}
        </span>
      </div>

      {/* Quick actions */}
      <div className="flex-none w-12 flex items-center justify-end gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
        <button
          onClick={(e) => {
            e.stopPropagation()
          }}
          className="flex items-center justify-center w-6 h-6 rounded text-muted-foreground hover:text-[var(--try-accent)] hover:bg-[var(--try-accent)]/10 transition-colors"
          aria-label="编辑"
        >
          <Edit3 size={11} />
        </button>
      </div>
    </div>
  )
}
