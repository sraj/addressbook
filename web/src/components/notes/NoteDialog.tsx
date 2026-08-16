import { useState, useEffect } from 'react'
import FormSheet from '@/components/ui/FormSheet'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import MDXEditor from '@/components/notes/MDXEditor'
import { showErrorToast } from '@/lib/error'
import { Pencil } from 'lucide-react'
import { type Note, SheetMode } from '@/types'

interface NoteDialogProps {
  mode: SheetMode
  note: Note | null
  onClose: () => void
  onSwitchToEdit: () => void
  onSave: (title: string, content: string) => Promise<void>
}

export default function NoteDialog({
  mode,
  note,
  onClose,
  onSwitchToEdit,
  onSave,
}: NoteDialogProps) {
  const [formTitle, setFormTitle] = useState(note?.title ?? '')
  const [formContent, setFormContent] = useState(note?.content ?? '')
  const [saving, setSaving] = useState(false)

  const isOpen = mode !== SheetMode.Closed
  const isEdit = mode === SheetMode.Edit

  useEffect(() => {
    if (mode === SheetMode.Edit) {
      setFormTitle(note?.title ?? '')
      setFormContent(note?.content ?? '')
    }
  }, [mode])

  const handleSave = async () => {
    if (!formTitle.trim() || saving) return
    setSaving(true)
    try {
      await onSave(formTitle, formContent)
    } catch (err) {
      showErrorToast(err, 'Failed to save note')
    } finally {
      setSaving(false)
    }
  }

  const title = isEdit ? (note ? 'Edit Note' : 'New Note') : 'View Note'

  return (
    <FormSheet
      open={isOpen}
      onOpenChange={(open) => { if (!open) onClose() }}
      title={title}
      wide
      footer={
        mode === SheetMode.View && note ? (
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
            <Button onClick={handleSave} disabled={saving || !formTitle.trim()}>
              {saving ? 'Saving…' : note ? 'Update' : 'Create'}
            </Button>
          </>
        ) : undefined
      }
    >
      <div className="flex flex-col flex-1 min-h-0">
        {mode === SheetMode.View && note && (
          <div className="space-y-4 flex flex-col flex-1">
            <MDXEditor
              value={note.content}
              onChange={() => {}}
              editable={false}
              title={note.title}
            />
          </div>
        )}

        {mode === SheetMode.Edit && (
          <div className="space-y-4 flex flex-col flex-1">
            <Input
              placeholder="Note title"
              className="shrink-0"
              value={formTitle}
              onChange={(e) => setFormTitle(e.target.value)}
            />
            <MDXEditor
              value={formContent}
              onChange={setFormContent}
              editable
            />
          </div>
        )}
      </div>
    </FormSheet>
  )
}
