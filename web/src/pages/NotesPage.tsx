import { useEffect, useState } from 'react'
import { useNotesStore } from '@/store/notes'
import PageLayout from '@/components/layout/PageLayout'
import PageHeader from '@/components/layout/PageHeader'
import UpgradeBanner from '@/components/ui/UpgradeBanner'
import { Button } from '@/components/ui/button'
import LoadingSkeleton from '@/components/ui/LoadingSkeleton'
import EmptyState from '@/components/ui/EmptyState'
import DeleteDialog from '@/components/ui/DeleteDialog'
import Pagination from '@/components/ui/Pagination'
import NoteCard from '@/components/notes/NoteCard'
import NoteDialog from '@/components/notes/NoteDialog'
import { SheetMode } from '@/types'
import { toast } from '@/hooks/use-toast'
import { Plus, FileText } from 'lucide-react'
import type { Note } from '@/types'

export default function NotesPage() {
  const {
    notes,
    total,
    page,
    totalPages,
    loading,
    fetchNotes,
    createNote,
    updateNote,
    deleteNote,
    setPage,
    quotaError,
  } = useNotesStore()

  const [sheetMode, setSheetMode] = useState<SheetMode>(SheetMode.Closed)
  const [activeNote, setActiveNote] = useState<Note | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<Note | null>(null)
  const [deleting, setDeleting] = useState(false)

  useEffect(() => {
    fetchNotes()
  }, [])

  const openView = (note: Note) => {
    setActiveNote(note)
    setSheetMode(SheetMode.View)
  }

  const openEdit = (note?: Note | null) => {
    setActiveNote(note ?? null)
    setSheetMode(SheetMode.Edit)
  }

  const closeDialog = () => {
    setSheetMode(SheetMode.Closed)
    setActiveNote(null)
  }

  const switchToEdit = () => {
    if (!activeNote) return
    setSheetMode(SheetMode.Edit)
  }

  const handleSave = async (title: string, content: string) => {
    if (activeNote) {
      await updateNote(activeNote.id, { title, content })
      toast({ title: 'Note updated', variant: 'success' })
    } else {
      await createNote({ title, content })
      toast({ title: 'Note created', variant: 'success' })
    }
    closeDialog()
  }

  const handleDelete = async () => {
    if (!deleteTarget) return
    setDeleting(true)
    try {
      await deleteNote(deleteTarget.id)
      toast({ title: 'Note deleted', variant: 'success' })
      setDeleteTarget(null)
    } catch {
      toast({ title: 'Error', description: 'Failed to delete note', variant: 'destructive' })
    } finally {
      setDeleting(false)
    }
  }

  return (
    <>
    <PageLayout>
      <PageHeader
        title="Notes"
        description="Write and manage your notes in Markdown"
        action={
          <Button onClick={() => openEdit(null)}>
            <Plus className="h-4 w-4" />
            New Note
          </Button>
        }
      />

        {quotaError && <UpgradeBanner resource="Notes" />}

        {loading ? (
          <LoadingSkeleton />
        ) : notes.length === 0 ? (
          <EmptyState
            icon={<FileText className="h-12 w-12" />}
            heading="No notes yet"
            description="Create your first note to get started"
            action={
              <Button onClick={() => openEdit(null)}>
                <Plus className="h-4 w-4" />
                New Note
              </Button>
            }
          />
        ) : (
          <>
            <div className="space-y-3">
              {notes.map((note) => (
                <NoteCard
                  key={note.id}
                  note={note}
                  onView={openView}
                  onEdit={openEdit}
                  onDelete={setDeleteTarget}
                />
              ))}
            </div>
            {totalPages > 1 && (
              <Pagination page={page} totalPages={totalPages} total={total} onPageChange={setPage} />
            )}
          </>
        )}

      <NoteDialog
        mode={sheetMode}
        note={activeNote}
        onClose={closeDialog}
        onSwitchToEdit={switchToEdit}
        onSave={handleSave}
      />

      <DeleteDialog
        open={!!deleteTarget}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
        title="Delete Note"
        label={deleteTarget?.title ?? ''}
        deleting={deleting}
        onConfirm={handleDelete}
      />
    </PageLayout>
    </>
  )
}
