import { useState, useEffect } from 'react'
import Select from 'react-select'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { toast } from '@/hooks/use-toast'
import { showErrorToast } from '@/lib/error'
import { api } from '@/lib/api'
import { useAuthStore } from '@/store/auth'
import type { UserPreferences } from '@/types'
import {
  Palette,
  Download,
  Info,
  Moon,
  Sun,
  FileDown,
  LayoutDashboard,
  List,
  Maximize2,
  Bell,
} from 'lucide-react'

const pageOptions = [
  { value: '/contacts', label: 'Contacts' },
  { value: '/notes', label: 'Notes' },
  { value: '/bookmarks', label: 'Bookmarks' },
]

const sizeOptions = [
  { value: 10, label: '10' },
  { value: 20, label: '20' },
  { value: 50, label: '50' },
  { value: 100, label: '100' },
]

const selectStyles = {
  control: (base: any) => ({ ...base, minHeight: 36 }),
  valueContainer: (base: any) => ({ ...base, padding: '2px 8px' }),
  indicatorSeparator: () => ({ display: 'none' }),
}

export default function GeneralSettings() {
  const user = useAuthStore((s) => s.user)
  const prefs = user?.preferences
  const [theme, setTheme] = useState<'light' | 'dark'>(
    () => (localStorage.getItem('theme') as 'light' | 'dark') || 'light',
  )
  const [defaultPage, setDefaultPage] = useState(prefs?.default_page || '/contacts')
  const [pageSize, setPageSize] = useState(prefs?.page_size || 20)
  const [compact, setCompact] = useState(prefs?.compact || false)

  const savePrefs = async (update: Partial<UserPreferences>) => {
    try {
      const res = await api.updateProfile(user?.name || '', {
        ...prefs,
        ...update,
      })
      useAuthStore.setState({ user: res.user })
    } catch (err) {
      showErrorToast(err, 'Failed to save settings')
    }
  }

  const handleExport = async () => {
    try {
      const res = await fetch('/api/contacts?size=10000', { credentials: 'include' })
      const data = await res.json()
      const blob = new Blob([JSON.stringify(data.data, null, 2)], { type: 'application/json' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `contacts-${new Date().toISOString().slice(0, 10)}.json`
      a.click()
      URL.revokeObjectURL(url)
      toast({ title: 'Export complete', description: `${data.total} contacts exported`, variant: 'success' })
    } catch {
      toast({ title: 'Export failed', variant: 'destructive' })
    }
  }

  useEffect(() => {
    const root = document.documentElement
    if (theme === 'dark') {
      root.classList.add('dark')
    } else {
      root.classList.remove('dark')
    }
    localStorage.setItem('theme', theme)
  }, [theme])

  const sections = [
    {
      id: 'theme',
      icon: Palette,
      title: 'Appearance',
      description: 'Switch between light and dark mode',
      content: (
        <div className="flex items-center gap-3">
          <Button variant={theme === 'light' ? 'default' : 'outline'} size="sm" className="gap-2" onClick={() => setTheme('light')}>
            <Sun className="h-4 w-4" />
            Light
          </Button>
          <Button variant={theme === 'dark' ? 'default' : 'outline'} size="sm" className="gap-2" onClick={() => setTheme('dark')}>
            <Moon className="h-4 w-4" />
            Dark
          </Button>
        </div>
      ),
    },
    {
      id: 'default-page',
      icon: LayoutDashboard,
      title: 'Default page',
      description: 'Choose your landing page after login',
      content: (
        <Select
          instanceId="default-page"
          options={pageOptions}
          value={pageOptions.find((o) => o.value === defaultPage)}
          onChange={(opt) => {
            if (opt) {
              setDefaultPage(opt.value)
              savePrefs({ default_page: opt.value })
            }
          }}
          styles={selectStyles}
        />
      ),
    },
    {
      id: 'page-size',
      icon: List,
      title: 'Items per page',
      description: 'Number of items shown in lists',
      content: (
        <Select
          instanceId="page-size"
          options={sizeOptions}
          value={sizeOptions.find((o) => o.value === pageSize)}
          onChange={(opt) => {
            if (opt) {
              setPageSize(opt.value)
              savePrefs({ page_size: opt.value })
            }
          }}
          styles={selectStyles}
        />
      ),
    },
    {
      id: 'compact',
      icon: Maximize2,
      title: 'Compact mode',
      description: 'Show lists in compact view',
      content: (
        <div className="flex items-center gap-3">
          <Button
            variant={!compact ? 'default' : 'outline'}
            size="sm"
            onClick={() => {
              setCompact(false)
              savePrefs({ compact: false })
            }}
          >
            Comfortable
          </Button>
          <Button
            variant={compact ? 'default' : 'outline'}
            size="sm"
            onClick={() => {
              setCompact(true)
              savePrefs({ compact: true })
            }}
          >
            Compact
          </Button>
        </div>
      ),
    },
    {
      id: 'data',
      icon: FileDown,
      title: 'Data',
      description: 'Export your contacts',
      content: (
        <div className="space-y-2">
          <p className="text-sm text-muted-foreground">Download all your contacts as a JSON file.</p>
          <Button variant="outline" size="sm" className="gap-2" onClick={handleExport}>
            <Download className="h-4 w-4" />
            Export contacts
          </Button>
        </div>
      ),
    },
    {
      id: 'notifications',
      icon: Bell,
      title: 'Notifications',
      description: 'Manage email notifications',
      content: (
        <div className="space-y-3">
          <label className="flex items-center gap-3 text-sm">
            <input
              type="checkbox"
              className="h-4 w-4 rounded border-gray-300"
              checked={prefs?.notify_billing ?? true}
              onChange={(e) => savePrefs({ notify_billing: e.target.checked })}
            />
            Billing updates (invoices, payment failures)
          </label>
          <label className="flex items-center gap-3 text-sm">
            <input
              type="checkbox"
              className="h-4 w-4 rounded border-gray-300"
              checked={prefs?.notify_marketing ?? true}
              onChange={(e) => savePrefs({ notify_marketing: e.target.checked })}
            />
            Product updates and tips
          </label>
        </div>
      ),
    },
    {
      id: 'about',
      icon: Info,
      title: 'About',
      description: 'Application information',
      content: (
        <div className="space-y-2 text-sm">
          <div className="flex justify-between">
            <span className="text-muted-foreground">Version</span>
            <span className="font-medium">1.0.0</span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted-foreground">Framework</span>
            <span className="font-medium">Kern + xdb</span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted-foreground">Database</span>
            <span className="font-medium">SQLite (FTS5)</span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted-foreground">Frontend</span>
            <span className="font-medium">React + shadcn/ui</span>
          </div>
          <div className="pt-2 border-t mt-2">
            <p className="text-muted-foreground text-xs">
              Built by{' '}
              <a href="https://www.linkedin.com/in/srajvenkat/" target="_blank" rel="noopener noreferrer" className="font-medium text-primary hover:underline">
                Suman Raj Venkatesan
              </a>
              — author of{' '}
              <a href="https://github.com/mobentum/kern" target="_blank" rel="noopener noreferrer" className="font-medium text-primary hover:underline">Kern</a>
              {' & '}
              <a href="https://github.com/mobentum/xdb" target="_blank" rel="noopener noreferrer" className="font-medium text-primary hover:underline">xdb</a>.
            </p>
          </div>
        </div>
      ),
    },
  ]

  return (
    <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
      {sections.map((section) => (
        <Card key={section.id}>
          <CardHeader className="pb-3">
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
  )
}
