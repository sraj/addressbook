import { Button } from '@/components/ui/button'
import { Pencil, Trash2, ExternalLink } from 'lucide-react'
import type { Bookmark } from '@/types'

interface BookmarkCardProps {
  bookmark: Bookmark
  onView: (bookmark: Bookmark) => void
  onEdit: (bookmark: Bookmark) => void
  onDelete: (bookmark: Bookmark) => void
}

export default function BookmarkCard({ bookmark, onView, onEdit, onDelete }: BookmarkCardProps) {
  return (
    <div
      className="rounded-lg border bg-background p-4 hover:border-primary/30 transition-colors cursor-pointer"
      onClick={() => onView(bookmark)}
    >
      <div className="flex items-start justify-between gap-4">
        <div className="flex items-start gap-3 min-w-0 flex-1">
          {bookmark.favicon_url ? (
            <img
              src={bookmark.favicon_url}
              alt=""
              className="h-6 w-6 shrink-0 rounded mt-0.5"
              onError={(e) => { (e.target as HTMLImageElement).style.display = 'none' }}
            />
          ) : (
            <div className="h-6 w-6 shrink-0 rounded bg-muted flex items-center justify-center text-xs text-muted-foreground">
              *
            </div>
          )}
          <div className="min-w-0">
            <h3 className="font-semibold truncate">{bookmark.title}</h3>
            <p className="mt-0.5 text-sm text-muted-foreground truncate max-w-md">
              {bookmark.url}
            </p>
            {bookmark.description && (
              <p className="mt-1 text-sm text-muted-foreground/70 line-clamp-1">
                {bookmark.description}
              </p>
            )}
            <div className="mt-1.5 flex items-center gap-2">
              {bookmark.category && (
                <span className="inline-flex items-center rounded-full bg-primary/10 px-2 py-0.5 text-xs font-medium text-primary">
                  {bookmark.category}
                </span>
              )}
            </div>
          </div>
        </div>
        <div className="flex shrink-0 gap-1" onClick={(e) => e.stopPropagation()}>
          <Button variant="ghost" size="icon" asChild>
            <a href={bookmark.url} target="_blank" rel="noopener noreferrer" onClick={(e) => e.stopPropagation()}>
              <ExternalLink className="h-4 w-4" />
            </a>
          </Button>
          <Button variant="ghost" size="icon" onClick={() => onEdit(bookmark)}>
            <Pencil className="h-4 w-4" />
          </Button>
          <Button variant="ghost" size="icon" onClick={() => onDelete(bookmark)}>
            <Trash2 className="h-4 w-4 text-destructive" />
          </Button>
        </div>
      </div>
    </div>
  )
}
