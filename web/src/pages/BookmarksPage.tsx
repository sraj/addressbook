import { useEffect, useState } from 'react'
import { useBookmarksStore } from '@/store/bookmarks'
import PageLayout from '@/components/layout/PageLayout'
import PageHeader from '@/components/layout/PageHeader'
import UpgradeBanner from '@/components/ui/UpgradeBanner'
import { Button } from '@/components/ui/button'
import LoadingSkeleton from '@/components/ui/LoadingSkeleton'
import EmptyState from '@/components/ui/EmptyState'
import DeleteDialog from '@/components/ui/DeleteDialog'
import Pagination from '@/components/ui/Pagination'
import BookmarkCard from '@/components/bookmarks/BookmarkCard'
import BookmarkDialog from '@/components/bookmarks/BookmarkDialog'
import { SheetMode } from '@/types'
import BookmarkImportDialog from '@/components/bookmarks/BookmarkImportDialog'
import CategoryFilter from '@/components/bookmarks/CategoryFilter'
import { toast } from '@/hooks/use-toast'
import { Plus, Upload, Bookmark as BookmarkIcon } from 'lucide-react'
import type { Bookmark } from '@/types'

export default function BookmarksPage() {
  const {
    bookmarks,
    total,
    page,
    totalPages,
    category,
    categories,
    loading,
    fetchBookmarks,
    fetchCategories,
    createBookmark,
    updateBookmark,
    deleteBookmark,
    importBookmarks,
    setCategory,
    setPage,
    quotaError,
  } = useBookmarksStore()

  const [sheetMode, setSheetMode] = useState<SheetMode>(SheetMode.Closed)
  const [activeBookmark, setActiveBookmark] = useState<Bookmark | null>(null)
  const [showImport, setShowImport] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<Bookmark | null>(null)
  const [deleting, setDeleting] = useState(false)

  useEffect(() => {
    fetchBookmarks()
    fetchCategories()
  }, [])

  const openView = (bookmark: Bookmark) => {
    setActiveBookmark(bookmark)
    setSheetMode(SheetMode.View)
  }

  const openEdit = (bookmark?: Bookmark | null) => {
    setActiveBookmark(bookmark ?? null)
    setSheetMode(SheetMode.Edit)
  }

  const closeDialog = () => {
    setSheetMode(SheetMode.Closed)
    setActiveBookmark(null)
  }

  const switchToEdit = () => {
    if (!activeBookmark) return
    setSheetMode(SheetMode.Edit)
  }

  const handleSave = async (data: { url: string; title: string; description?: string; favicon_url?: string; category?: string }) => {
    if (activeBookmark) {
      await updateBookmark(activeBookmark.id, data)
      toast({ title: 'Bookmark updated', variant: 'success' })
    } else {
      await createBookmark(data)
      toast({ title: 'Bookmark created', variant: 'success' })
    }
    closeDialog()
  }

  const handleImport = async (items: Array<{ url: string; title: string; description?: string; favicon_url?: string; category?: string }>) => {
    await importBookmarks(items)
  }

  const handleDelete = async () => {
    if (!deleteTarget) return
    setDeleting(true)
    try {
      await deleteBookmark(deleteTarget.id)
      toast({ title: 'Bookmark deleted', variant: 'success' })
      setDeleteTarget(null)
    } catch {
      toast({ title: 'Error', description: 'Failed to delete bookmark', variant: 'destructive' })
    } finally {
      setDeleting(false)
    }
  }

  return (
    <PageLayout>
      <PageHeader
        title="Bookmarks"
        description="Save and organize your bookmarks"
        action={
          <div className="flex gap-2">
            <Button variant="outline" onClick={() => setShowImport(true)}>
              <Upload className="h-4 w-4" />
              Import
            </Button>
            <Button onClick={() => openEdit(null)}>
              <Plus className="h-4 w-4" />
              Add Bookmark
            </Button>
          </div>
        }
      />

      {quotaError && <UpgradeBanner resource="Bookmarks" />}

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-[220px_1fr]">
        <div className="space-y-1 lg:border-r lg:pr-4">
          <CategoryFilter
            categories={categories}
            selected={category}
            onSelect={setCategory}
          />
        </div>

        <div>
          {loading ? (
            <LoadingSkeleton />
          ) : bookmarks.length === 0 ? (
            <EmptyState
              icon={<BookmarkIcon className="h-12 w-12" />}
              heading={category ? 'No bookmarks in this category' : 'No bookmarks yet'}
              description={category ? `No bookmarks in "${category}"` : 'Add your first bookmark to get started'}
              action={!category ? (
                <div className="flex gap-2">
                  <Button variant="outline" onClick={() => setShowImport(true)}>
                    <Upload className="h-4 w-4" />
                    Import
                  </Button>
                  <Button onClick={() => openEdit(null)}>
                    <Plus className="h-4 w-4" />
                    Add Bookmark
                  </Button>
                </div>
              ) : undefined}
            />
          ) : (
            <>
              <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                {bookmarks.map((bookmark) => (
                  <BookmarkCard
                    key={bookmark.id}
                    bookmark={bookmark}
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
        </div>
      </div>

      <BookmarkDialog
        mode={sheetMode}
        bookmark={activeBookmark}
        categories={categories}
        onClose={closeDialog}
        onSwitchToEdit={switchToEdit}
        onSave={handleSave}
      />

      <BookmarkImportDialog
        open={showImport}
        onOpenChange={setShowImport}
        onImport={handleImport}
      />

      <DeleteDialog
        open={!!deleteTarget}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
        title="Delete Bookmark"
        label={deleteTarget?.title ?? ''}
        deleting={deleting}
        onConfirm={handleDelete}
      />
    </PageLayout>
  )
}
