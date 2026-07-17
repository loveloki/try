'use client'

import { FileType } from '@/lib/try-types'
import { cn } from '@/lib/utils'
import {
  Folder,
  FileCode,
  FileText,
  FileJson,
  File,
  FileArchive,
  Image,
} from 'lucide-react'

interface FileIconProps {
  type: FileType
  isDir?: boolean
  size?: number
  className?: string
}

const TYPE_CONFIG: Record<FileType, { icon: React.ElementType; color: string }> = {
  dir:     { icon: Folder,      color: 'text-[var(--try-blue)]' },
  go:      { icon: FileCode,    color: 'text-[#00ADD8]' },
  ts:      { icon: FileCode,    color: 'text-[#3178c6]' },
  js:      { icon: FileCode,    color: 'text-yellow-400' },
  md:      { icon: FileText,    color: 'text-[var(--try-match)]' },
  json:    { icon: FileJson,    color: 'text-[var(--try-accent)]' },
  txt:     { icon: FileText,    color: 'text-muted-foreground' },
  docx:    { icon: FileText,    color: 'text-blue-400' },
  zip:     { icon: FileArchive, color: 'text-purple-400' },
  image:   { icon: Image,       color: 'text-pink-400' },
  unknown: { icon: File,        color: 'text-muted-foreground' },
}

export function FileIcon({ type, size = 15, className }: FileIconProps) {
  const config = TYPE_CONFIG[type] ?? TYPE_CONFIG.unknown
  const Icon = config.icon
  return <Icon size={size} className={cn(config.color, className)} />
}
