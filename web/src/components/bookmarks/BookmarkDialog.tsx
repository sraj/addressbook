import { useState, useEffect } from 'react'
import FormSheet from '@/components/ui/FormSheet'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { showErrorToast } from '@/lib/error'
import { faviconURL, extractDomain } from '@/lib/bookmark-import'
import { Pencil, ExternalLink } from 'lucide-react'
import { type Bookmark, SheetMode } from '@/types'

interface BookmarkDialogProps {
  mode: SheetMode
  bookmark: Bookmark | null
  categories: string[]
  onClose: () => void
  onSwitchToEdit: () => void
  onSave: (data: { url: string; title: string; description?: string; favicon_url?: string; category?: string }) => Promise<void>
}

export default function BookmarkDialog({ mode, bookmark, categories, onClose, onSwitchToEdit, onSave }: BookmarkDialogProps) {
  const [formURL, setFormURL] = useState(bookmark?.url ?? '')
  const [formTitle, setFormTitle] = useState(bookmark?.title ?? '')
  const [formDescription, setFormDescription] = useState(bookmark?.description ?? '')
  const [formCategory, setFormCategory] = useState(bookmark?.category ?? '')
  const [saving, setSaving] = useState(false)
  const [newCategory, setNewCategory] = useState('')

  const isOpen = mode !== SheetMode.Closed
  const isEdit = mode === SheetMode.Edit

  useEffect(() => {
    if (mode === SheetMode.Edit) {
      setFormURL(bookmark?.url ?? '')
      setFormTitle(bookmark?.title ?? '')
      setFormDescription(bookmark?.description ?? '')
      setFormCategory(bookmark?.category ?? '')
      setNewCategory('')
    }
  }, [mode])

  const handleURLChange = (url: string) => {
    setFormURL(url)
    const domain = extractDomain(url)
    if (domain && !formTitle) {
      setFormTitle(domain)
    }
  }

  const handleSave = async () => {
    if (!formURL.trim() || !formTitle.trim() || saving) return
    const category = newCategory || formCategory
    setSaving(true)
    try {
      await onSave({
        url: formURL,
        title: formTitle,
        description: formDescription,
        favicon_url: faviconURL(formURL),
        category,
      })
    } catch (err) {
      showErrorToast(err, 'Failed to save bookmark')
    } finally {
      setSaving(false)
    }
  }

  const title = isEdit ? (bookmark ? 'Edit Bookmark' : 'New Bookmark') : 'View Bookmark'

  return (
    <FormSheet
      open={isOpen}
      onOpenChange={(open) => { if (!open) onClose() }}
      title={title}
      footer={
        mode === SheetMode.View && bookmark ? (
          <>
            <Button variant="outline" onClick={onClose}>Close</Button>
            <Button onClick={onSwitchToEdit}>
              <Pencil className="h-4 w-4" />
              Edit
            </Button>
          </>
        ) : mode === SheetMode.Edit ? (
          <>
            <Button variant="outline" onClick={onClose}>Cancel</Button>
            <Button onClick={handleSave} disabled={saving || !formURL.trim() || !formTitle.trim()}>
              {saving ? 'Saving…' : bookmark ? 'Update' : 'Create'}
            </Button>
          </>
        ) : undefined
      }
    >
      {mode === SheetMode.View && bookmark && (
        <div className="space-y-4">
          <div className="flex items-center gap-3">
            {bookmark.favicon_url ? (
              <img src={bookmark.favicon_url} alt="" className="h-8 w-8 rounded" />
            ) : (
              <div className="h-8 w-8 rounded bg-muted flex items-center justify-center text-sm text-muted-foreground">*</div>
            )}
            <div>
              <h2 className="text-xl font-semibold">{bookmark.title}</h2>
            </div>
          </div>
          <a
            href={bookmark.url}
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center gap-1 text-sm text-primary hover:underline"
          >
            <ExternalLink className="h-4 w-4" />
            {bookmark.url}
          </a>
          {bookmark.description && (
            <p className="text-sm text-muted-foreground">{bookmark.description}</p>
          )}
          {bookmark.category && (
            <span className="inline-flex self-start items-center rounded-full bg-primary/10 px-3 py-1 text-sm font-medium text-primary">
              {bookmark.category}
            </span>
          )}
        </div>
      )}

      {mode === SheetMode.Edit && (
        <div className="space-y-4">
          {formURL && (
            <div className="flex items-center gap-2 text-xs text-muted-foreground">
              <img
                src={faviconURL(formURL)}
                alt=""
                className="h-4 w-4 rounded"
                onError={(e) => { (e.target as HTMLImageElement).style.display = 'none' }}
              />
              {extractDomain(formURL) && <span>Favicon will be auto-detected</span>}
            </div>
          )}
          <div className="space-y-2">
            <label className="text-sm font-medium">URL *</label>
            <Input
              placeholder="https://example.com"
              value={formURL}
              onChange={(e) => handleURLChange(e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Title *</label>
            <Input
              placeholder="My Bookmark"
              value={formTitle}
              onChange={(e) => setFormTitle(e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Description</label>
            <textarea
              className="flex min-h-[80px] w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
              placeholder="Optional description…"
              value={formDescription}
              onChange={(e) => setFormDescription(e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Category</label>
            <Select
              value={formCategory || (categories.length > 0 ? '' : '')}
              onValueChange={(value) => { setFormCategory(value); setNewCategory('') }}
            >
              <SelectTrigger>
                <SelectValue placeholder="None" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="">None</SelectItem>
                {categories.map((c) => (
                  <SelectItem key={c} value={c}>{c}</SelectItem>
                ))}
                <SelectItem value="__new__">+ New category</SelectItem>
              </SelectContent>
            </Select>
            {formCategory === '__new__' && (
              <Input
                placeholder="Enter new category name"
                value={newCategory}
                onChange={(e) => setNewCategory(e.target.value)}
                className="mt-1"
              />
            )}
          </div>
        </div>
      )}
    </FormSheet>
  )
}
