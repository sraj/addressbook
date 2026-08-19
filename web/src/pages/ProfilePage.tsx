import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { toast } from '@/hooks/use-toast'
import { showErrorToast } from '@/lib/error'
import { api } from '@/lib/api'
import { useAuthStore } from '@/store/auth'
import { User, ShieldCheck } from 'lucide-react'
import PageLayout from '@/components/layout/PageLayout'
import PageHeader from '@/components/layout/PageHeader'

export default function ProfilePage() {
  const user = useAuthStore((s) => s.user)
  const [name, setName] = useState(user?.name ?? '')
  const [saving, setSaving] = useState(false)

  const handleSaveName = async () => {
    if (!name.trim()) return
    setSaving(true)
    try {
      const res = await api.updateProfile(name.trim())
      useAuthStore.setState({ user: res.user })
      toast({ title: 'Name updated', variant: 'success' })
    } catch (err) {
      showErrorToast(err, 'Failed to update name')
    } finally {
      setSaving(false)
    }
  }

  const sections = [
    {
      id: 'profile',
      icon: User,
      title: 'Profile',
      description: 'Your account information',
      content: (
        <div className="space-y-3">
          <div className="space-y-1">
            <Label>Email address</Label>
            <Input value={user?.email ?? ''} readOnly className="bg-muted/50" />
            <p className="text-xs text-muted-foreground">Email cannot be changed</p>
          </div>
          <div className="space-y-1">
            <Label htmlFor="name">Name</Label>
            <div className="flex gap-2">
              <Input id="name" value={name} onChange={(e) => setName(e.target.value)} placeholder="Your name" />
              <Button size="sm" onClick={handleSaveName} disabled={saving || !name.trim() || name === user?.name}>
                {saving ? 'Saving…' : 'Save'}
              </Button>
            </div>
          </div>
        </div>
      ),
    },
    {
      id: 'security',
      icon: ShieldCheck,
      title: 'Security',
      description: 'Change your password',
      content: (
        <form
          className="space-y-3"
          onSubmit={async (e) => {
            e.preventDefault()
            const form = e.currentTarget
            const formData = new FormData(form)
            const current = formData.get('currentPassword') as string
            const newPass = formData.get('newPassword') as string
            const confirm = formData.get('confirmPassword') as string
            if (!current || !newPass || !confirm) {
              toast({ title: 'All fields are required', variant: 'destructive' })
              return
            }
            if (newPass.length < 8) {
              toast({ title: 'Password must be at least 8 characters', variant: 'destructive' })
              return
            }
            if (newPass !== confirm) {
              toast({ title: 'Passwords do not match', variant: 'destructive' })
              return
            }
            try {
              await api.changePassword(current, newPass)
              toast({ title: 'Password changed successfully', variant: 'success' })
              form.reset()
            } catch (err) {
              showErrorToast(err, 'Password change failed')
            }
          }}
        >
          <div className="space-y-1">
            <Label htmlFor="currentPassword">Current password</Label>
            <Input id="currentPassword" name="currentPassword" type="password" placeholder="••••••••" />
          </div>
          <div className="space-y-1">
            <Label htmlFor="newPassword">New password</Label>
            <Input id="newPassword" name="newPassword" type="password" placeholder="••••••••" />
          </div>
          <div className="space-y-1">
            <Label htmlFor="confirmPassword">Confirm new password</Label>
            <Input id="confirmPassword" name="confirmPassword" type="password" placeholder="••••••••" />
          </div>
          <Button type="submit" size="sm">Update password</Button>
        </form>
      ),
    },
  ]

  return (
    <PageLayout>
      <PageHeader title="Profile" description="Manage your account and security" />
      <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
        {sections.map((section) => (
          <Card key={section.id}>
            <CardHeader>
              <div className="flex items-center gap-2">
                <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary/10">
                  <section.icon className="h-4 w-4 text-primary" />
                </div>
                <div>
                  <CardTitle className="text-base">{section.title}</CardTitle>
                  <CardDescription>{section.description}</CardDescription>
                </div>
              </div>
            </CardHeader>
            <CardContent>{section.content}</CardContent>
          </Card>
        ))}
      </div>
    </PageLayout>
  )
}
