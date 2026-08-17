import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { api } from '@/lib/api'
import { useCollectionsStore } from '@/store/collections'
import PageLayout from '@/components/layout/PageLayout'
import PageHeader from '@/components/layout/PageHeader'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import LoadingSkeleton from '@/components/ui/LoadingSkeleton'
import EmptyState from '@/components/ui/EmptyState'
import Pagination from '@/components/ui/Pagination'
import { toast } from '@/hooks/use-toast'
import { showErrorToast } from '@/lib/error'
import { ArrowLeft, Link as LinkIcon, Copy, RefreshCw, Upload, Download, Users, Plus, FolderInput, FolderMinus } from 'lucide-react'
import ContactFormDialog from '@/components/contacts/ContactFormDialog'
import MoveContactDialog from '@/components/collections/MoveContactDialog'
import DeleteDialog from '@/components/ui/DeleteDialog'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'
import LabelPrintOrderCard from '@/components/labels/LabelPrintOrderCard'
import type { Contact } from '@/types'

export default function CollectionDetailPage() {
  const { id } = useParams<{ id: string }>()
  const collectionId = Number(id)
  const navigate = useNavigate()

  const { collections, fetchCollections } = useCollectionsStore()
  const collection = collections.find((c) => c.id === collectionId)

  const [contacts, setContacts] = useState<Contact[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [totalPages, setTotalPages] = useState(0)
  const [loading, setLoading] = useState(true)

  const [importing, setImporting] = useState(false)
  const [exporting, setExporting] = useState(false)
  const [importResult, setImportResult] = useState<string | null>(null)

  const [showAddForm, setShowAddForm] = useState(false)
  const [moveTarget, setMoveTarget] = useState<Contact | null>(null)
  const [removeTarget, setRemoveTarget] = useState<Contact | null>(null)
  const [removing, setRemoving] = useState(false)

  const inviteUrl = collection ? `${window.location.origin}/invite/${collection.invite_token}` : ''

  const loadContacts = async (p = page) => {
    setLoading(true)
    try {
      const res = await api.getCollectionContacts(collectionId, { page: p, size: 20 })
      setContacts(res.data)
      setTotal(res.total)
      setPage(res.page)
      setTotalPages(res.total_pages)
    } catch (err) {
      showErrorToast(err, 'Failed to load collection contacts')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchCollections()
  }, [])

  useEffect(() => {
    if (!Number.isNaN(collectionId)) loadContacts(1)
  }, [collectionId])

  const copyInvite = async () => {
    try {
      await navigator.clipboard.writeText(inviteUrl)
      toast({ title: 'Invite link copied', variant: 'success' })
    } catch {
      toast({ title: 'Could not copy', description: 'Select and copy the link manually', variant: 'destructive' })
    }
  }

  const handleRegenerate = async () => {
    if (!window.confirm('Regenerate the invite link? The old link will stop working.')) return
    try {
      await useCollectionsStore.getState().regenerateToken(collectionId)
      toast({ title: 'Invite link regenerated', variant: 'success' })
    } catch (err) {
      showErrorToast(err, 'Failed to regenerate invite link')
    }
  }

  const handleImport = async (file: File, format: 'csv' | 'xlsx') => {
    setImporting(true)
    setImportResult(null)
    try {
      const res = await api.importContacts(file, format, collectionId)
      setImportResult(`Imported ${res.imported} contacts${res.skipped ? ` (${res.skipped} skipped)` : ''}`)
      await loadContacts(1)
      await fetchCollections()
    } catch (err) {
      showErrorToast(err, 'Import failed')
    } finally {
      setImporting(false)
    }
  }

  const handleExport = async (format: 'csv' | 'xlsx') => {
    setExporting(true)
    try {
      await api.exportContacts(format, collectionId)
    } catch (err) {
      showErrorToast(err, 'Export failed')
    } finally {
      setExporting(false)
    }
  }

  const handleRemoveFromCollection = async () => {
    if (!removeTarget) return
    setRemoving(true)
    try {
      await api.removeContactFromCollection(collectionId, removeTarget.id)
      toast({ title: 'Removed from collection', variant: 'success' })
      setRemoveTarget(null)
      await Promise.all([loadContacts(page), fetchCollections()])
    } catch (err) {
      showErrorToast(err, 'Failed to remove contact')
    } finally {
      setRemoving(false)
    }
  }

  const handleContactChanged = async () => {
    await Promise.all([loadContacts(page), fetchCollections()])
  }

  return (
    <PageLayout>
      <Button variant="ghost" size="sm" className="mb-2 -ml-2" onClick={() => navigate('/collections')}>
        <ArrowLeft className="h-4 w-4" />
        Back to Collections
      </Button>

      <PageHeader
        title={collection?.name ?? 'Collection'}
        description={collection ? `${total} contact${total === 1 ? '' : 's'}` : undefined}
        action={
          <Button onClick={() => setShowAddForm(true)}>
            <Plus className="h-4 w-4" />
            Add Contact
          </Button>
        }
      />

      <Tabs defaultValue="contacts">
        <TabsList>
          <TabsTrigger value="contacts">Contacts</TabsTrigger>
          <TabsTrigger value="labels">Order / Download Labels</TabsTrigger>
        </TabsList>

        <TabsContent value="contacts">
          <div className="mb-4 grid grid-cols-1 gap-4 lg:grid-cols-2">
            <Card>
              <CardHeader className="pb-3">
                <CardTitle className="text-base flex items-center gap-2">
                  <LinkIcon className="h-4 w-4" />
                  Invite Link
                </CardTitle>
                <CardDescription>Share this link so others can submit their address</CardDescription>
              </CardHeader>
              <CardContent className="space-y-3">
                <div className="flex items-center gap-2">
                  <Input readOnly value={inviteUrl} className="flex-1 min-w-0 font-mono text-xs" onFocus={(e) => e.target.select()} />
                  <Button size="icon" variant="outline" onClick={copyInvite} title="Copy link" className="shrink-0">
                    <Copy className="h-4 w-4" />
                  </Button>
                  <Button size="icon" variant="ghost" onClick={handleRegenerate} title="Regenerate link" className="shrink-0">
                    <RefreshCw className="h-4 w-4" />
                  </Button>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="pb-3">
                <CardTitle className="text-base flex items-center gap-2">
                  <Users className="h-4 w-4" />
                  Import / Export
                </CardTitle>
                <CardDescription>Bulk add contacts to this collection or export them</CardDescription>
              </CardHeader>
              <CardContent className="space-y-3">
                <div className="flex flex-wrap gap-2">
                  <label className="inline-flex">
                    <Button variant="outline" size="sm" asChild disabled={importing}>
                      <span>
                        <Upload className="h-4 w-4" />
                        {importing ? 'Importing…' : 'Import CSV'}
                      </span>
                    </Button>
                    <input
                      type="file"
                      accept=".csv"
                      className="hidden"
                      onChange={(e) => {
                        const f = e.target.files?.[0]
                        if (f) handleImport(f, 'csv')
                        e.target.value = ''
                      }}
                    />
                  </label>
                  <label className="inline-flex">
                    <Button variant="outline" size="sm" asChild disabled={importing}>
                      <span>
                        <Upload className="h-4 w-4" />
                        Import XLSX
                      </span>
                    </Button>
                    <input
                      type="file"
                      accept=".xlsx"
                      className="hidden"
                      onChange={(e) => {
                        const f = e.target.files?.[0]
                        if (f) handleImport(f, 'xlsx')
                        e.target.value = ''
                      }}
                    />
                  </label>
                  <Button variant="outline" size="sm" onClick={() => handleExport('csv')} disabled={exporting}>
                    <Download className="h-4 w-4" />
                    Export CSV
                  </Button>
                  <Button variant="outline" size="sm" onClick={() => handleExport('xlsx')} disabled={exporting}>
                    <Download className="h-4 w-4" />
                    Export XLSX
                  </Button>
                </div>
                {importResult && <p className="text-sm text-success">{importResult}</p>}
              </CardContent>
            </Card>
          </div>

          {loading ? (
            <LoadingSkeleton height="h-12" />
          ) : contacts.length === 0 ? (
            <EmptyState
              icon={<Users className="h-12 w-12" />}
              heading="No contacts in this collection"
              description="Share the invite link to collect addresses"
            />
          ) : (
            <>
              <div className="rounded-lg border bg-background">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Name</TableHead>
                      <TableHead className="hidden sm:table-cell">Email</TableHead>
                      <TableHead className="hidden md:table-cell">Address</TableHead>
                      <TableHead className="w-[80px] text-right">Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {contacts.map((contact) => (
                      <TableRow key={contact.id}>
                        <TableCell className="font-medium">{contact.name}</TableCell>
                        <TableCell className="hidden sm:table-cell text-muted-foreground">{contact.emails?.[0]}</TableCell>
                        <TableCell className="hidden md:table-cell text-muted-foreground">
                          {contact.addresses?.[0]?.line1}, {contact.addresses?.[0]?.city}
                        </TableCell>
                        <TableCell className="text-right">
                          <div className="flex justify-end gap-1">
                            <Button variant="ghost" size="icon" title="Move to another collection" onClick={() => setMoveTarget(contact)}>
                              <FolderInput className="h-4 w-4" />
                            </Button>
                            <Button variant="ghost" size="icon" title="Remove from this collection" onClick={() => setRemoveTarget(contact)}>
                              <FolderMinus className="h-4 w-4 text-destructive" />
                            </Button>
                          </div>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
              {totalPages > 1 && (
                <Pagination page={page} totalPages={totalPages} total={total} onPageChange={loadContacts} />
              )}
            </>
          )}
        </TabsContent>

        <TabsContent value="labels">
          <LabelPrintOrderCard collectionId={collectionId} />
        </TabsContent>
      </Tabs>
    <ContactFormDialog
        open={showAddForm}
        onOpenChange={setShowAddForm}
        editingContact={null}
        collectionId={collectionId}
      />

      <MoveContactDialog
        open={!!moveTarget}
        onOpenChange={(open) => !open && setMoveTarget(null)}
        contact={moveTarget ?? ({} as Contact)}
        onMoved={handleContactChanged}
      />

      <DeleteDialog
        open={!!removeTarget}
        onOpenChange={(open) => !open && setRemoveTarget(null)}
        title="Remove from Collection"
        label={removeTarget?.name ?? ''}
        deleting={removing}
        description={
          <>
            Remove <strong>{removeTarget?.name}</strong> from this collection? The contact stays in your address book.
          </>
        }
        confirmLabel="Remove"
        onConfirm={handleRemoveFromCollection}
      />
    </PageLayout>
  )
}