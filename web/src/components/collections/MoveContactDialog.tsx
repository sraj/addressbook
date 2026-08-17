import { useEffect, useState } from 'react'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { toast } from '@/hooks/use-toast'
import { showErrorToast } from '@/lib/error'
import { api } from '@/lib/api'
import { useCollectionsStore } from '@/store/collections'
import type { Contact } from '@/types'

interface Props {
  open: boolean
  onOpenChange: (open: boolean) => void
  contact: Contact
  onMoved?: () => void
}

export default function MoveContactDialog({ open, onOpenChange, contact, onMoved }: Props) {
  const { collections, fetchCollections } = useCollectionsStore()
  const [target, setTarget] = useState<string>('0')
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    if (open) {
      setTarget(String(contact.collection_id ?? 0))
      if (collections.length === 0) fetchCollections()
    }
  }, [open, contact.collection_id])

  const handleMove = async () => {
    setSaving(true)
    try {
      const collectionId = Number(target)
      if (collectionId > 0) {
        await api.moveContactToCollection(collectionId, contact.id)
      } else if (contact.collection_id) {
        await api.removeContactFromCollection(contact.collection_id, contact.id)
      }
      toast({ title: 'Contact moved', variant: 'success' })
      onOpenChange(false)
      onMoved?.()
    } catch (err) {
      showErrorToast(err, 'Failed to move contact')
    } finally {
      setSaving(false)
    }
  }

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Move Contact</AlertDialogTitle>
          <AlertDialogDescription>
            Move <strong>{contact.name}</strong> to a collection.
          </AlertDialogDescription>
        </AlertDialogHeader>

        <div className="space-y-2">
          <Label htmlFor="move-target">Collection</Label>
          <Select value={target} onValueChange={setTarget}>
            <SelectTrigger id="move-target" className="w-full">
              <SelectValue placeholder="Select a collection" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="0">No collection</SelectItem>
              {collections.map((c) => (
                <SelectItem key={c.id} value={String(c.id)}>{c.name}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <AlertDialogFooter>
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <AlertDialogAction onClick={handleMove} disabled={saving}>
            {saving ? 'Moving…' : 'Move'}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}