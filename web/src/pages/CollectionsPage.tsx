import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useCollectionsStore } from '@/store/collections'
import PageLayout from '@/components/layout/PageLayout'
import PageHeader from '@/components/layout/PageHeader'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import FormSheet from '@/components/ui/FormSheet'
import EmptyState from '@/components/ui/EmptyState'
import DeleteDialog from '@/components/ui/DeleteDialog'
import LoadingSkeleton from '@/components/ui/LoadingSkeleton'
import { toast } from '@/hooks/use-toast'
import { showErrorToast } from '@/lib/error'
import { Folder, Plus, Pencil, Trash2, Users, ArrowRight, RefreshCw } from 'lucide-react'
import type { Collection } from '@/types'

export default function CollectionsPage() {
  const navigate = useNavigate()
  const { collections, loading, fetchCollections, createCollection, renameCollection, deleteCollection, regenerateToken } = useCollectionsStore()

  const [showForm, setShowForm] = useState(false)
  const [editing, setEditing] = useState<Collection | null>(null)
  const [name, setName] = useState('')
  const [saving, setSaving] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<Collection | null>(null)
  const [deleting, setDeleting] = useState(false)

  useEffect(() => {
    fetchCollections()
  }, [])

  const openCreate = () => {
    setEditing(null)
    setName('')
    setShowForm(true)
  }

  const openEdit = (collection: Collection) => {
    setEditing(collection)
    setName(collection.name)
    setShowForm(true)
  }

  const handleSave = async () => {
    if (!name.trim()) return
    setSaving(true)
    try {
      if (editing) {
        await renameCollection(editing.id, name.trim())
        toast({ title: 'Collection renamed', variant: 'success' })
      } else {
        await createCollection(name.trim())
        toast({ title: 'Collection created', variant: 'success' })
      }
      setShowForm(false)
    } catch (err) {
      showErrorToast(err, 'Failed to save collection')
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async () => {
    if (!deleteTarget) return
    setDeleting(true)
    try {
      await deleteCollection(deleteTarget.id)
      toast({ title: 'Collection deleted', description: 'Contacts were kept in your address book', variant: 'success' })
      setDeleteTarget(null)
    } catch (err) {
      showErrorToast(err, 'Failed to delete collection')
    } finally {
      setDeleting(false)
    }
  }

  const handleRegenerate = async (collection: Collection) => {
    if (!window.confirm('Regenerate the invite link? The old link will stop working.')) return
    try {
      await regenerateToken(collection.id)
      toast({ title: 'Invite link regenerated', variant: 'success' })
    } catch (err) {
      showErrorToast(err, 'Failed to regenerate invite link')
    }
  }

  return (
    <PageLayout>
      <PageHeader
        title="Collections"
        description="Group contacts and share invite links to collect addresses"
        action={
          <Button onClick={openCreate}>
            <Plus className="h-4 w-4" />
            New Collection
          </Button>
        }
      />

      {loading ? (
        <LoadingSkeleton height="h-20" />
      ) : collections.length === 0 ? (
        <EmptyState
          icon={<Folder className="h-12 w-12" />}
          heading="No collections yet"
          description="Create a collection, then share its invite link to collect addresses"
          action={
            <Button onClick={openCreate}>
              <Plus className="h-4 w-4" />
              New Collection
            </Button>
          }
        />
      ) : (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          {collections.map((collection) => (
            <Card key={collection.id} className="cursor-pointer transition-shadow hover:shadow-md" onClick={() => navigate(`/collections/${collection.id}`)}>
              <CardHeader className="pb-2">
                <div className="flex items-start justify-between gap-2">
                  <div className="flex items-center gap-3">
                    <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-primary/10">
                      <Folder className="h-4 w-4 text-primary" />
                    </div>
                    <div>
                      <CardTitle className="text-base">{collection.name}</CardTitle>
                      <CardDescription className="flex items-center gap-1">
                        <Users className="h-3 w-3" />
                        {collection.contact_count ?? 0} contacts
                      </CardDescription>
                    </div>
                  </div>
                </div>
              </CardHeader>
              <CardContent>
                <div className="flex items-center justify-between">
                  <Button size="sm" variant="outline" onClick={(e) => { e.stopPropagation(); navigate(`/collections/${collection.id}`) }}>
                    Manage
                    <ArrowRight className="ml-1 h-3 w-3" />
                  </Button>
                  <div className="flex gap-1" onClick={(e) => e.stopPropagation()}>
                    <Button variant="ghost" size="icon" title="Rename" onClick={() => openEdit(collection)}>
                      <Pencil className="h-4 w-4" />
                    </Button>
                    <Button variant="ghost" size="icon" title="Regenerate invite link" onClick={() => handleRegenerate(collection)}>
                      <RefreshCw className="h-4 w-4" />
                    </Button>
                    <Button variant="ghost" size="icon" title="Delete collection" onClick={() => setDeleteTarget(collection)}>
                      <Trash2 className="h-4 w-4 text-destructive" />
                    </Button>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      <FormSheet open={showForm} onOpenChange={setShowForm} title={editing ? 'Rename Collection' : 'New Collection'} description={editing ? 'Update the collection name' : 'Create a collection to group contacts'}>
        <div className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Name</label>
            <Input
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g. Holiday Cards"
              autoFocus
              onKeyDown={(e) => e.key === 'Enter' && handleSave()}
            />
          </div>
          <div className="flex justify-end gap-2">
            <Button variant="outline" onClick={() => setShowForm(false)}>Cancel</Button>
            <Button onClick={handleSave} disabled={saving || !name.trim()}>
              {saving ? 'Saving…' : editing ? 'Save' : 'Create'}
            </Button>
          </div>
        </div>
      </FormSheet>

      <DeleteDialog
        open={!!deleteTarget}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
        title="Delete Collection"
        label={deleteTarget?.name ?? ''}
        deleting={deleting}
        onConfirm={handleDelete}
      />
    </PageLayout>
  )
}