import { Button } from '@/components/ui/button'
import { Pencil, Trash2 } from 'lucide-react'
import type { Note } from '@/types'

interface NoteCardProps {
  note: Note
  onView: (note: Note) => void
  onEdit: (note: Note) => void
  onDelete: (note: Note) => void
}

function stripMarkdown(md: string) {
  return md
    .replace(/^#+\s+/gm, '')
    .replace(/[*_~`]/g, '')
    .replace(/\[([^\]]+)\]\([^)]+\)/g, '$1')
    .substring(0, 150)
}

export default function NoteCard({
  note,
  onView,
  onEdit,
  onDelete,
}: NoteCardProps) {
  return (
    <div
      className="rounded-lg border bg-background p-4 hover:border-primary/30 transition-colors cursor-pointer"
      onClick={() => onView(note)}
    >
      <div className="flex items-start justify-between gap-4">
        <div className="min-w-0 flex-1">
          <h3 className="font-semibold truncate">{note.title}</h3>
          <p className="mt-1 text-sm text-muted-foreground line-clamp-2">
            {stripMarkdown(note.content) || (
              <span className="italic">Empty note</span>
            )}
          </p>
          <p className="mt-2 text-xs text-muted-foreground">
            Updated {note.updated_at}
          </p>
        </div>
        <div
          className="flex shrink-0 gap-1"
          onClick={(e) => e.stopPropagation()}
        >
          <Button variant="ghost" size="icon" onClick={() => onEdit(note)}>
            <Pencil className="h-4 w-4" />
          </Button>
          <Button variant="ghost" size="icon" onClick={() => onDelete(note)}>
            <Trash2 className="h-4 w-4 text-destructive" />
          </Button>
        </div>
      </div>
    </div>
  )
}
