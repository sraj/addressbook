import { useEffect, useState, useCallback, useRef } from 'react'
import { useContactsStore } from '@/store/contacts'
import PageLayout from '@/components/layout/PageLayout'
import PageHeader from '@/components/layout/PageHeader'
import UpgradeBanner from '@/components/ui/UpgradeBanner'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import LoadingSkeleton from '@/components/ui/LoadingSkeleton'
import EmptyState from '@/components/ui/EmptyState'
import DeleteDialog from '@/components/ui/DeleteDialog'
import Pagination from '@/components/ui/Pagination'
import { toast } from '@/hooks/use-toast'
import { Plus, Search, Pencil, Trash2, User } from 'lucide-react'
import ContactFormDialog from '@/components/contacts/ContactFormDialog'
import type { Contact } from '@/types'

export default function ContactsPage() {
  const {
    contacts,
    total,
    page,
    totalPages,
    searchQuery,
    loading,
    fetchContacts,
    deleteContact,
    setSearch,
    setPage,
    quotaError,
  } = useContactsStore()

  const [searchInput, setSearchInput] = useState('')
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const [showForm, setShowForm] = useState(false)
  const [editingContact, setEditingContact] = useState<Contact | null>(null)

  const [deleteTarget, setDeleteTarget] = useState<Contact | null>(null)
  const [deleting, setDeleting] = useState(false)

  useEffect(() => {
    fetchContacts()
    return () => {
      if (debounceRef.current) clearTimeout(debounceRef.current)
    }
  }, [])

  const handleSearch = useCallback((value: string) => {
    setSearchInput(value)
    if (debounceRef.current) clearTimeout(debounceRef.current)
    debounceRef.current = setTimeout(() => {
      setSearch(value)
    }, 300)
  }, [])

  const openCreate = () => {
    setEditingContact(null)
    setShowForm(true)
  }

  const openEdit = (contact: Contact) => {
    setEditingContact(contact)
    setShowForm(true)
  }

  const handleDelete = async () => {
    if (!deleteTarget) return
    setDeleting(true)
    try {
      await deleteContact(deleteTarget.id)
      toast({ title: 'Contact deleted', variant: 'success' })
      setDeleteTarget(null)
    } catch {
      toast({ title: 'Error', description: 'Failed to delete contact', variant: 'destructive' })
    } finally {
      setDeleting(false)
    }
  }

  return (
    <>
    <PageLayout>
      <PageHeader
        title="Contacts"
        description="Manage your contacts"
        action={
          <Button onClick={openCreate}>
            <Plus className="h-4 w-4" />
            Add Contact
          </Button>
        }
      />

        {quotaError && <UpgradeBanner resource="Contacts" />}

        <div className="relative mb-4">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Search contacts…"
            className="pl-9"
            value={searchInput}
            onChange={(e) => handleSearch(e.target.value)}
          />
        </div>

        {loading ? (
          <LoadingSkeleton height="h-12" />
        ) : contacts.length === 0 ? (
          <EmptyState
            icon={<User className="h-12 w-12" />}
            heading={searchQuery ? 'No results found' : 'No contacts yet'}
            description={searchQuery ? `No contacts matching "${searchQuery}"` : 'Add your first contact to get started'}
            action={!searchQuery ? (
              <Button onClick={openCreate}>
                <Plus className="h-4 w-4" />
                Add Contact
              </Button>
            ) : undefined}
          />
        ) : (
          <>
            <div className="rounded-lg border bg-background">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Name</TableHead>
                    <TableHead className="hidden sm:table-cell">Email</TableHead>
                    <TableHead className="hidden md:table-cell">Phone</TableHead>
                    <TableHead className="hidden lg:table-cell">Address</TableHead>
                    <TableHead className="w-[80px] text-right">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {contacts.map((contact) => (
                    <TableRow key={contact.id}>
                      <TableCell className="font-medium">{contact.name}</TableCell>
                      <TableCell className="hidden sm:table-cell text-muted-foreground max-w-[200px] truncate">
                        {contact.emails?.slice(0, 2).map((e, i) => (
                          <span key={i}>{e}<br /></span>
                        ))}
                        {contact.emails?.length > 2 && <span className="text-xs">+{contact.emails.length - 2} more</span>}
                      </TableCell>
                      <TableCell className="hidden md:table-cell text-muted-foreground max-w-[150px] truncate">
                        {contact.phones?.slice(0, 2).map((p, i) => (
                          <span key={i}>{p}<br /></span>
                        ))}
                        {contact.phones?.length > 2 && <span className="text-xs">+{contact.phones.length - 2} more</span>}
                      </TableCell>
                      <TableCell className="hidden lg:table-cell text-muted-foreground max-w-[200px] truncate">
                        {contact.addresses[0]?.line1}, {contact.addresses[0]?.city}
                      </TableCell>
                      <TableCell className="text-right">
                        <div className="flex justify-end gap-1">
                          <Button variant="ghost" size="icon" onClick={() => openEdit(contact)}>
                            <Pencil className="h-4 w-4" />
                          </Button>
                          <Button variant="ghost" size="icon" onClick={() => setDeleteTarget(contact)}>
                            <Trash2 className="h-4 w-4 text-destructive" />
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>

            {totalPages > 1 && (
              <Pagination page={page} totalPages={totalPages} total={total} onPageChange={setPage} />
            )}
          </>
        )}

      <ContactFormDialog
        open={showForm}
        onOpenChange={setShowForm}
        editingContact={editingContact}
      />

      <DeleteDialog
        open={!!deleteTarget}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
        title="Delete Contact"
        label={deleteTarget?.name ?? ''}
        deleting={deleting}
        onConfirm={handleDelete}
      />
    </PageLayout>
    </>
  )
}
