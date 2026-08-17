import { useState, useEffect } from 'react'
import FormSheet from '@/components/ui/FormSheet'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { toast } from '@/hooks/use-toast'
import { handleFieldErrors, showErrorToast } from '@/lib/error'
import { api } from '@/lib/api'
import { Mail, Phone, MapPin, X } from 'lucide-react'
import { contactSchema } from '@/lib/schemas'
import { useContactsStore } from '@/store/contacts'
import { useCollectionsStore } from '@/store/collections'
import type { Contact } from '@/types'
import type { z } from 'zod'

type FormData = z.infer<typeof contactSchema>

const emptyForm = (): FormData => ({
  name: '',
  emails: [''],
  phones: [''],
  addresses: [{ label: 'Home', line1: '', line2: '', city: '', state: '', zip: '', country: '' }],
  notes: '',
})

function contactToForm(contact: Contact): FormData {
  const emails = contact.emails ?? []
  const phones = contact.phones ?? []
  const addresses = contact.addresses ?? []
  return {
    name: contact.name,
    emails: emails.length > 0 ? emails : [''],
    phones: phones.length > 0 ? phones : [''],
    addresses: addresses.length > 0
      ? addresses.map((a) => ({
          label: (['Home', 'Office', 'Other'].includes(a.label) ? a.label : 'Home') as 'Home' | 'Office' | 'Other',
          line1: a.line1,
          line2: a.line2 || '',
          city: a.city,
          state: a.state,
          zip: a.zip,
          country: a.country,
        }))
      : [{ label: 'Home' as const, line1: '', line2: '', city: '', state: '', zip: '', country: '' }],
    notes: contact.notes,
  }
}

function flattenZodErrors(error: z.ZodError): Record<string, string> {
  const result: Record<string, string> = {}
  for (const issue of error.issues) {
    const path = issue.path.join('.')
    if (!result[path]) {
      result[path] = issue.message
    }
  }
  return result
}

function serverFieldToPath(field: string): string {
  return field
    .replace(/\[(\d+)\]/g, '.$1')
    .replace(/\.(\w)/g, (_, c) => `.${c.toLowerCase()}`)
    .replace(/^(\w)/, (c) => c.toLowerCase())
}

interface Props {
  open: boolean
  onOpenChange: (open: boolean) => void
  editingContact: Contact | null
  collectionId?: number
}

export default function ContactFormDialog({ open, onOpenChange, editingContact, collectionId }: Props) {
  const { createContact, updateContact } = useContactsStore()
  const fetchCollections = useCollectionsStore((s) => s.fetchCollections)
  const [form, setForm] = useState<FormData>(emptyForm())
  const [saving, setSaving] = useState(false)
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})

  useEffect(() => {
    if (open) {
      setForm(editingContact ? contactToForm(editingContact) : emptyForm())
      setFieldErrors({})
      setSaving(false)
    }
  }, [open, editingContact])

  const clearFieldError = (path: string) => {
    setFieldErrors((prev) => {
      const next = { ...prev }
      const parts = path.split('.')
      for (let i = parts.length; i > 0; i--) {
        delete next[parts.slice(0, i).join('.')]
      }
      return next
    })
  }

  const addField = (field: 'emails' | 'phones' | 'addresses') => {
    setForm((prev) => ({
      ...prev,
      [field]:
        field === 'addresses'
          ? [...prev.addresses, { label: '', line1: '', line2: '', city: '', state: '', zip: '', country: '' }]
          : [...prev[field], ''],
    }))
  }

  const removeField = (field: 'emails' | 'phones', index: number) => {
    setForm((prev) => ({
      ...prev,
      [field]: prev[field].filter((_, i) => i !== index),
    }))
  }

  const removeAddress = (index: number) => {
    setForm((prev) => ({
      ...prev,
      addresses: prev.addresses.filter((_, i) => i !== index),
    }))
  }

  const updateArrayField = (field: 'emails' | 'phones', index: number, value: string) => {
    clearFieldError(`${field}.${index}`)
    setForm((prev) => {
      const updated = [...prev[field]]
      updated[index] = value
      return { ...prev, [field]: updated }
    })
  }

  const updateAddress = (index: number, key: string, value: string) => {
    clearFieldError(`addresses.${index}.${key}`)
    setForm((prev) => {
      const updated = [...prev.addresses]
      updated[index] = { ...updated[index], [key]: value }
      return { ...prev, addresses: updated }
    })
  }

  const handleSave = async () => {
    const result = contactSchema.safeParse(form)
    if (!result.success) {
      setFieldErrors(flattenZodErrors(result.error))
      return
    }
    setFieldErrors({})

    setSaving(true)
    try {
      const cleanedEmails = form.emails.filter((e) => e.trim())
      const cleanedPhones = form.phones.filter((p) => p.trim())
      const cleanedAddresses = form.addresses.filter((a) => a.line1.trim())
      const payload = {
        name: form.name.trim(),
        emails: cleanedEmails,
        phones: cleanedPhones,
        addresses: cleanedAddresses.map((a) => ({
          ...a,
          line2: a.line2 || undefined,
        })),
        notes: (form.notes || '').trim(),
      }
      if (editingContact) {
        await updateContact(editingContact.id, payload)
        toast({ title: 'Contact updated', variant: 'success' })
      } else if (collectionId) {
        await api.addCollectionContact(collectionId, payload)
        await fetchCollections()
        toast({ title: 'Contact added to collection', variant: 'success' })
      } else {
        await createContact(payload)
        toast({ title: 'Contact created', variant: 'success' })
      }
      onOpenChange(false)
    } catch (err) {
      if (!handleFieldErrors(err, setFieldErrors, serverFieldToPath)) {
        showErrorToast(err, 'Failed to save contact')
      }
    } finally {
      setSaving(false)
    }
  }

  const err = (path: string) => fieldErrors[path]

  return (
    <FormSheet
      open={open}
      onOpenChange={onOpenChange}
      title={editingContact ? 'Edit Contact' : 'Add Contact'}
      description={editingContact ? 'Update the contact details below' : 'Fill in the details to add a new contact'}
      footer={
        <>
          <Button variant="outline" onClick={() => onOpenChange(false)}>Cancel</Button>
          <Button onClick={handleSave} disabled={saving}>
            {saving ? 'Saving…' : editingContact ? 'Update Contact' : 'Create Contact'}
          </Button>
        </>
      }
    >
      <div className="space-y-5">
        <div className="space-y-2">
          <Label htmlFor="name">Name *</Label>
            <Input
              id="name"
              placeholder="John Doe"
              value={form.name}
              onChange={(e) => {
                clearFieldError('name')
                setForm((p) => ({ ...p, name: e.target.value }))
              }}
            />
            {err('name') && <p className="text-sm text-destructive">{err('name')}</p>}
          </div>

          <div className="space-y-2">
            <Label>Email addresses *</Label>
            {form.emails.map((email, i) => (
              <div key={i}>
                <div className="flex items-center gap-2">
                  <Mail className="h-4 w-4 shrink-0 text-muted-foreground" />
                  <Input
                    placeholder="email@example.com"
                    value={email}
                    onChange={(e) => updateArrayField('emails', i, e.target.value)}
                  />
                  {form.emails.length > 1 && (
                    <Button variant="ghost" size="icon" onClick={() => removeField('emails', i)}>
                      <X className="h-4 w-4" />
                    </Button>
                  )}
                </div>
                {err(`emails.${i}`) && <p className="mt-1 text-sm text-destructive">{err(`emails.${i}`)}</p>}
              </div>
            ))}
            {err('emails') && <p className="text-sm text-destructive">{err('emails')}</p>}
            <Button variant="link" size="sm" className="h-auto p-0" onClick={() => addField('emails')}>
              + Add email
            </Button>
          </div>

          <div className="space-y-2">
            <Label>Phone numbers *</Label>
            {form.phones.map((phone, i) => (
              <div key={i}>
                <div className="flex items-center gap-2">
                  <Phone className="h-4 w-4 shrink-0 text-muted-foreground" />
                  <Input
                    placeholder="+1-555-0100"
                    value={phone}
                    onChange={(e) => updateArrayField('phones', i, e.target.value)}
                  />
                  {form.phones.length > 1 && (
                    <Button variant="ghost" size="icon" onClick={() => removeField('phones', i)}>
                      <X className="h-4 w-4" />
                    </Button>
                  )}
                </div>
                {err(`phones.${i}`) && <p className="mt-1 text-sm text-destructive">{err(`phones.${i}`)}</p>}
              </div>
            ))}
            {err('phones') && <p className="text-sm text-destructive">{err('phones')}</p>}
            <Button variant="link" size="sm" className="h-auto p-0" onClick={() => addField('phones')}>
              + Add phone
            </Button>
          </div>

          <div className="space-y-3">
            <Label>Addresses *</Label>
            {form.addresses.map((addr, i) => (
              <div key={i} className="rounded-lg border p-3 space-y-2">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                    <MapPin className="h-4 w-4" />
                    Address {i + 1}
                  </div>
                  {form.addresses.length > 1 && (
                    <Button variant="ghost" size="icon" onClick={() => removeAddress(i)}>
                      <X className="h-4 w-4" />
                    </Button>
                  )}
                </div>
                <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
                  <div className="col-span-2">
                    <Select value={addr.label} onValueChange={(value) => updateAddress(i, 'label', value)}>
                      <SelectTrigger>
                        <SelectValue placeholder="Select label" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="Home">Home</SelectItem>
                        <SelectItem value="Office">Office</SelectItem>
                        <SelectItem value="Other">Other</SelectItem>
                      </SelectContent>
                    </Select>
                    {err(`addresses.${i}.label`) && <p className="mt-1 text-sm text-destructive">{err(`addresses.${i}.label`)}</p>}
                  </div>
                  <div className="col-span-2">
                    <Input
                      placeholder="Street address"
                      value={addr.line1}
                      onChange={(e) => updateAddress(i, 'line1', e.target.value)}
                    />
                    {err(`addresses.${i}.line1`) && <p className="mt-1 text-sm text-destructive">{err(`addresses.${i}.line1`)}</p>}
                  </div>
                  <div className="col-span-2">
                    <Input
                      placeholder="Apt, suite (optional)"
                      value={addr.line2}
                      onChange={(e) => updateAddress(i, 'line2', e.target.value)}
                    />
                  </div>
                  <div className="space-y-1">
                    <Input
                      placeholder="City"
                      value={addr.city}
                      onChange={(e) => updateAddress(i, 'city', e.target.value)}
                    />
                    {err(`addresses.${i}.city`) && <p className="text-sm text-destructive">{err(`addresses.${i}.city`)}</p>}
                  </div>
                  <div className="space-y-1">
                    <Input
                      placeholder="State"
                      value={addr.state}
                      onChange={(e) => updateAddress(i, 'state', e.target.value)}
                    />
                    {err(`addresses.${i}.state`) && <p className="text-sm text-destructive">{err(`addresses.${i}.state`)}</p>}
                  </div>
                  <div className="space-y-1">
                    <Input
                      placeholder="ZIP"
                      value={addr.zip}
                      onChange={(e) => updateAddress(i, 'zip', e.target.value)}
                    />
                    {err(`addresses.${i}.zip`) && <p className="text-sm text-destructive">{err(`addresses.${i}.zip`)}</p>}
                  </div>
                  <div className="space-y-1">
                    <Input
                      placeholder="Country"
                      value={addr.country}
                      onChange={(e) => updateAddress(i, 'country', e.target.value)}
                    />
                    {err(`addresses.${i}.country`) && <p className="text-sm text-destructive">{err(`addresses.${i}.country`)}</p>}
                  </div>
                </div>
              </div>
            ))}
            {err('addresses') && <p className="text-sm text-destructive">{err('addresses')}</p>}
            <Button variant="link" size="sm" className="h-auto p-0" onClick={() => addField('addresses')}>
              + Add address
            </Button>
          </div>

          <div className="space-y-2">
            <Label htmlFor="notes">Notes</Label>
            <textarea
              id="notes"
              className="flex min-h-[80px] w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
              placeholder="Any additional notes…"
              value={form.notes}
              onChange={(e) => {
                clearFieldError('notes')
                setForm((p) => ({ ...p, notes: e.target.value }))
              }}
            />
            {err('notes') && <p className="text-sm text-destructive">{err('notes')}</p>}
          </div>
      </div>
    </FormSheet>
  )
}
