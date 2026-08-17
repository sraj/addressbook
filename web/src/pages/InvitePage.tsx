import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { api } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { showErrorToast } from '@/lib/error'
import { toast } from '@/hooks/use-toast'
import { BookUser, CheckCircle2 } from 'lucide-react'

interface FormState {
  name: string
  email: string
  phone: string
  label: string
  line1: string
  line2: string
  city: string
  state: string
  zip: string
  country: string
}

const initialForm: FormState = {
  name: '',
  email: '',
  phone: '',
  label: 'Home',
  line1: '',
  line2: '',
  city: '',
  state: '',
  zip: '',
  country: '',
}

export default function InvitePage() {
  const { token } = useParams<{ token: string }>()
  const [collectionName, setCollectionName] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [notFound, setNotFound] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [success, setSuccess] = useState<string | null>(null)
  const [form, setForm] = useState<FormState>(initialForm)

  useEffect(() => {
    if (!token) {
      setNotFound(true)
      setLoading(false)
      return
    }
    api
      .getInviteInfo(token)
      .then((res) => setCollectionName(res.name))
      .catch(() => setNotFound(true))
      .finally(() => setLoading(false))
  }, [token])

  const setField = (field: keyof FormState) => (value: string) => {
    setForm((f) => ({ ...f, [field]: value }))
  }

  const handleSubmit = async () => {
    if (!form.name.trim() || !form.line1.trim() || !form.city.trim() || !form.state.trim() || !form.zip.trim() || !form.country.trim()) {
      toast({ title: 'Missing fields', description: 'Please fill in name and complete address', variant: 'destructive' })
      return
    }
    setSubmitting(true)
    try {
      const res = await api.submitInvite(token!, {
        name: form.name.trim(),
        email: form.email.trim() || undefined,
        phone: form.phone.trim() || undefined,
        address: {
          label: form.label || undefined,
          line1: form.line1.trim(),
          line2: form.line2.trim() || undefined,
          city: form.city.trim(),
          state: form.state.trim(),
          zip: form.zip.trim(),
          country: form.country.trim(),
        },
      })
      setSuccess(res.name)
    } catch (err) {
      showErrorToast(err, 'Failed to submit your address')
    } finally {
      setSubmitting(false)
    }
  }

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-muted/30 p-4">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle>Loading…</CardTitle>
          </CardHeader>
        </Card>
      </div>
    )
  }

  if (notFound) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-muted/30 p-4">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle>Invite not found</CardTitle>
            <CardDescription>This invite link is invalid or has been disabled.</CardDescription>
          </CardHeader>
        </Card>
      </div>
    )
  }

  if (success) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-muted/30 p-4">
        <Card className="w-full max-w-md text-center">
          <CardHeader className="items-center">
            <CheckCircle2 className="h-12 w-12 text-emerald-600" />
            <CardTitle>Thanks, {success}!</CardTitle>
            <CardDescription>
              Your address has been added to <span className="font-medium text-foreground">{collectionName}</span>.
            </CardDescription>
          </CardHeader>
        </Card>
      </div>
    )
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-muted/30 p-4">
      <div className="w-full max-w-lg">
        <div className="mb-6 flex items-center justify-center gap-2">
          <BookUser className="h-7 w-7 text-emerald-600" />
          <h1 className="text-xl font-bold">{collectionName}</h1>
        </div>

        <Card>
          <CardHeader>
            <CardTitle>Add your address</CardTitle>
            <CardDescription>Help us keep your address up to date. Please fill in your details below.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="name">Full name</Label>
              <Input id="name" value={form.name} onChange={(e) => setField('name')(e.target.value)} placeholder="Jane Doe" />
            </div>

            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div className="space-y-2">
                <Label htmlFor="email">Email (optional)</Label>
                <Input id="email" type="email" value={form.email} onChange={(e) => setField('email')(e.target.value)} placeholder="jane@example.com" />
              </div>
              <div className="space-y-2">
                <Label htmlFor="phone">Phone (optional)</Label>
                <Input id="phone" value={form.phone} onChange={(e) => setField('phone')(e.target.value)} placeholder="+1 555 000 0000" />
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="label">Address label</Label>
              <Input id="label" value={form.label} onChange={(e) => setField('label')(e.target.value)} placeholder="Home" />
            </div>

            <div className="space-y-2">
              <Label htmlFor="line1">Address line 1</Label>
              <Input id="line1" value={form.line1} onChange={(e) => setField('line1')(e.target.value)} placeholder="123 Main Street" />
            </div>

            <div className="space-y-2">
              <Label htmlFor="line2">Address line 2 (optional)</Label>
              <Input id="line2" value={form.line2} onChange={(e) => setField('line2')(e.target.value)} placeholder="Apt 4B" />
            </div>

            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div className="space-y-2">
                <Label htmlFor="city">City</Label>
                <Input id="city" value={form.city} onChange={(e) => setField('city')(e.target.value)} placeholder="Springfield" />
              </div>
              <div className="space-y-2">
                <Label htmlFor="state">State / Province</Label>
                <Input id="state" value={form.state} onChange={(e) => setField('state')(e.target.value)} placeholder="IL" />
              </div>
              <div className="space-y-2">
                <Label htmlFor="zip">ZIP / Postal code</Label>
                <Input id="zip" value={form.zip} onChange={(e) => setField('zip')(e.target.value)} placeholder="62704" />
              </div>
              <div className="space-y-2">
                <Label htmlFor="country">Country</Label>
                <Input id="country" value={form.country} onChange={(e) => setField('country')(e.target.value)} placeholder="United States" />
              </div>
            </div>

            <Button className="w-full" onClick={handleSubmit} disabled={submitting}>
              {submitting ? 'Submitting…' : 'Submit'}
            </Button>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}