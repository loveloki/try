// ─── Core Types ───────────────────────────────────────

export type SourceTab = 'tries' | 'ship' | 'bug' | 'all'
export type EntrySource = 'tries' | 'ship' | 'bug'

export type FileType =
  | 'dir'
  | 'go'
  | 'ts'
  | 'js'
  | 'md'
  | 'json'
  | 'txt'
  | 'docx'
  | 'zip'
  | 'image'
  | 'unknown'

export interface TryEntry {
  id: string
  name: string        // 原始目录名（含日期）
  baseName: string    // 去除日期后缀的名称
  date: string        // YYYY-MM-DD
  source: EntrySource
  score: number       // 时间权重 + 模糊匹配分
  lastModified: Date
  fileCount: number
  sizeKB: number
  tags?: string[]
  markedForDelete?: boolean
}

export interface FileEntry {
  id: string
  name: string
  type: FileType
  sizeKB: number
  modified: Date
  isDir: boolean
  children?: FileEntry[]
}

export interface AppState {
  view: 'selector' | 'files'
  activeTab: SourceTab
  searchQuery: string
  selectedEntry: TryEntry | null
  deleteMode: boolean
  markedIds: Set<string>
  showDeleteConfirm: boolean
  showCreateInput: boolean
  createName: string
  showRenameInput: boolean
  renameName: string
  dropHover: boolean
}

// ─── Mock Data ────────────────────────────────────────

const now = new Date()
const daysAgo = (d: number) => new Date(now.getTime() - d * 86400000)

export const MOCK_ENTRIES: TryEntry[] = [
  {
    id: '1',
    name: 'axum-middleware-2025-07-15',
    baseName: 'axum-middleware',
    date: '2025-07-15',
    source: 'tries',
    score: 98,
    lastModified: daysAgo(2),
    fileCount: 12,
    sizeKB: 48,
    tags: ['rust', 'web'],
  },
  {
    id: '2',
    name: 'try-gui-fyne-2025-07-14',
    baseName: 'try-gui-fyne',
    date: '2025-07-14',
    source: 'tries',
    score: 95,
    lastModified: daysAgo(3),
    fileCount: 7,
    sizeKB: 22,
    tags: ['go', 'gui'],
  },
  {
    id: '3',
    name: 'sqlx-pool-bench-2025-07-10',
    baseName: 'sqlx-pool-bench',
    date: '2025-07-10',
    source: 'tries',
    score: 87,
    lastModified: daysAgo(7),
    fileCount: 5,
    sizeKB: 18,
    tags: ['rust', 'database'],
  },
  {
    id: '4',
    name: 'openai-stream-2025-07-08',
    baseName: 'openai-stream',
    date: '2025-07-08',
    source: 'tries',
    score: 83,
    lastModified: daysAgo(9),
    fileCount: 9,
    sizeKB: 34,
    tags: ['ai', 'go'],
  },
  {
    id: '5',
    name: 'nix-devshell-2025-07-05',
    baseName: 'nix-devshell',
    date: '2025-07-05',
    source: 'tries',
    score: 76,
    lastModified: daysAgo(12),
    fileCount: 4,
    sizeKB: 8,
    tags: ['nix'],
  },
  {
    id: '6',
    name: 'grpc-reflection-2025-06-28',
    baseName: 'grpc-reflection',
    date: '2025-06-28',
    source: 'tries',
    score: 70,
    lastModified: daysAgo(19),
    fileCount: 15,
    sizeKB: 60,
    tags: ['go', 'grpc'],
  },
  {
    id: '7',
    name: 'tree-sitter-go-2025-06-20',
    baseName: 'tree-sitter-go',
    date: '2025-06-20',
    source: 'tries',
    score: 62,
    lastModified: daysAgo(27),
    fileCount: 6,
    sizeKB: 25,
    tags: ['go', 'parser'],
  },
  {
    id: '8',
    name: 'wasm-canvas-2025-06-10',
    baseName: 'wasm-canvas',
    date: '2025-06-10',
    source: 'tries',
    score: 54,
    lastModified: daysAgo(37),
    fileCount: 8,
    sizeKB: 31,
    tags: ['wasm', 'web'],
  },
  {
    id: '9',
    name: 'bubbletea-v2-2025-05-30',
    baseName: 'bubbletea-v2',
    date: '2025-05-30',
    source: 'bug',
    score: 91,
    lastModified: daysAgo(2),
    fileCount: 22,
    sizeKB: 88,
    tags: ['go', 'tui'],
  },
  {
    id: '10',
    name: 'lipgloss-theme-2025-05-20',
    baseName: 'lipgloss-theme',
    date: '2025-05-20',
    source: 'bug',
    score: 72,
    lastModified: daysAgo(18),
    fileCount: 11,
    sizeKB: 42,
    tags: ['go', 'tui'],
  },
  {
    id: '11',
    name: 'redis-streams-2025-04-15',
    baseName: 'redis-streams',
    date: '2025-04-15',
    source: 'tries',
    score: 45,
    lastModified: daysAgo(93),
    fileCount: 7,
    sizeKB: 19,
    tags: ['go', 'redis'],
  },
  {
    id: '12',
    name: 'ebpf-tracing-2025-03-01',
    baseName: 'ebpf-tracing',
    date: '2025-03-01',
    source: 'tries',
    score: 32,
    lastModified: daysAgo(138),
    fileCount: 3,
    sizeKB: 11,
    tags: ['linux', 'ebpf'],
  },
]

export const MOCK_FILES: FileEntry[] = [
  {
    id: 'f1',
    name: 'main.go',
    type: 'go',
    sizeKB: 3.2,
    modified: daysAgo(2),
    isDir: false,
  },
  {
    id: 'f2',
    name: 'handler.go',
    type: 'go',
    sizeKB: 5.8,
    modified: daysAgo(2),
    isDir: false,
  },
  {
    id: 'f3',
    name: 'middleware.go',
    type: 'go',
    sizeKB: 4.1,
    modified: daysAgo(3),
    isDir: false,
  },
  {
    id: 'f4',
    name: 'go.mod',
    type: 'unknown',
    sizeKB: 0.5,
    modified: daysAgo(3),
    isDir: false,
  },
  {
    id: 'f5',
    name: 'go.sum',
    type: 'unknown',
    sizeKB: 12.1,
    modified: daysAgo(3),
    isDir: false,
  },
  {
    id: 'f6',
    name: 'README.md',
    type: 'md',
    sizeKB: 2.3,
    modified: daysAgo(2),
    isDir: false,
  },
  {
    id: 'f7',
    name: 'notes',
    type: 'dir',
    sizeKB: 0,
    modified: daysAgo(1),
    isDir: true,
  },
  {
    id: 'f8',
    name: 'testdata',
    type: 'dir',
    sizeKB: 0,
    modified: daysAgo(3),
    isDir: true,
  },
  {
    id: 'f9',
    name: 'design.docx',
    type: 'docx',
    sizeKB: 28.4,
    modified: daysAgo(4),
    isDir: false,
  },
  {
    id: 'f10',
    name: 'config.json',
    type: 'json',
    sizeKB: 1.2,
    modified: daysAgo(2),
    isDir: false,
  },
]

// ─── Fuzzy Match ──────────────────────────────────────

export interface MatchSegment {
  text: string
  matched: boolean
}

export function fuzzyMatch(text: string, query: string): MatchSegment[] | null {
  if (!query) return [{ text, matched: false }]
  const lower = text.toLowerCase()
  const q = query.toLowerCase()

  const matchedIndices = new Set<number>()
  let qi = 0
  for (let i = 0; i < lower.length && qi < q.length; i++) {
    if (lower[i] === q[qi]) {
      matchedIndices.add(i)
      qi++
    }
  }
  if (qi < q.length) return null

  const segments: MatchSegment[] = []
  let i = 0
  while (i < text.length) {
    if (matchedIndices.has(i)) {
      let end = i
      while (end < text.length && matchedIndices.has(end)) end++
      segments.push({ text: text.slice(i, end), matched: true })
      i = end
    } else {
      let end = i
      while (end < text.length && !matchedIndices.has(end)) end++
      segments.push({ text: text.slice(i, end), matched: false })
      i = end
    }
  }
  return segments
}

export function relativeTime(date: Date): string {
  const seconds = Math.floor((Date.now() - date.getTime()) / 1000)
  if (seconds < 60) return '刚刚'
  const minutes = Math.floor(seconds / 60)
  if (minutes < 60) return `${minutes}m 前`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}h 前`
  const days = Math.floor(hours / 24)
  if (days < 30) return `${days}d 前`
  const months = Math.floor(days / 30)
  if (months < 12) return `${months}mo 前`
  return `${Math.floor(months / 12)}y 前`
}

export function formatSize(kb: number): string {
  if (kb < 1) return `${Math.round(kb * 1024)}B`
  if (kb < 1024) return `${kb.toFixed(1)}KB`
  return `${(kb / 1024).toFixed(1)}MB`
}
