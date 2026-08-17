import { useEffect, useState } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'
import { useAuthStore } from '@/store/auth'
import { Button } from '@/components/ui/button'
import { BookUser, Users, LogOut, Settings, FileText, Bookmark, Shield, User, Folder, Tag } from 'lucide-react'
import { cn } from '@/lib/utils'

export default function AppShell({ children }: { children: React.ReactNode }) {
  const navigate = useNavigate()
  const location = useLocation()
  const user = useAuthStore((s) => s.user)
  const logout = useAuthStore((s) => s.logout)
  const [showGreeting, setShowGreeting] = useState(true)
  const [fading, setFading] = useState(false)

  useEffect(() => {
    const t1 = setTimeout(() => setFading(true), 3500)
    const t2 = setTimeout(() => setShowGreeting(false), 4200)
    return () => {
      clearTimeout(t1)
      clearTimeout(t2)
    }
  }, [])

  const isAdmin = user?.role === 'admin'
  const navItems = [
    { icon: Users, label: 'Contacts', path: '/contacts' },
    { icon: Folder, label: 'Collections', path: '/collections', startsWith: true },
    { icon: Tag, label: 'Labels', path: '/labels' },
    { icon: FileText, label: 'Notes', path: '/notes' },
    { icon: Bookmark, label: 'Bookmarks', path: '/bookmarks' },
    { icon: Settings, label: 'Settings', path: '/settings' },
  ]

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  return (
    <div className="flex min-h-screen">
      <aside className="fixed left-4 top-4 z-30 flex h-[calc(100vh-2rem)] w-[60px] flex-col items-center rounded-xl border bg-background py-3 shadow-sm sm:w-[80px] sm:py-4">
        <div className="mb-6 flex h-9 w-9 items-center justify-center rounded-lg sm:mb-8 sm:h-10 sm:w-10">
          <BookUser className="h-6 w-6 text-emerald-600 sm:h-7 sm:w-7" />
        </div>

        <nav className="flex flex-1 flex-col items-center gap-2">
          {navItems.map((item) => {
            const active = item.path === '/settings'
              ? location.pathname.startsWith('/settings')
              : item.startsWith
                ? location.pathname.startsWith(item.path)
                : location.pathname === item.path
            return (
              <Button
                key={item.label}
                variant="ghost"
                size="icon"
                className={cn(
                  'h-9 w-9 rounded-lg sm:h-10 sm:w-10',
                  active && 'bg-primary/10 text-primary hover:bg-primary/15',
                )}
                onClick={() => navigate(item.path)}
                title={item.label}
              >
                <item.icon className="h-4 w-4 sm:h-5 sm:w-5" />
              </Button>
            )
          })}
        </nav>

        <Button
          variant="ghost"
          size="icon"
          className={cn(
            'h-9 w-9 rounded-lg sm:h-10 sm:w-10',
            location.pathname === '/profile'
              ? 'bg-primary/10 text-primary hover:bg-primary/15'
              : 'text-muted-foreground hover:text-foreground',
          )}
          onClick={() => navigate('/profile')}
          title="Profile"
        >
          <User className="h-4 w-4 sm:h-5 sm:w-5" />
        </Button>

        {isAdmin && (
          <Button
            variant="ghost"
            size="icon"
            className={cn(
              'h-9 w-9 rounded-lg sm:h-10 sm:w-10 mb-1',
              location.pathname === '/admin'
                ? 'bg-purple-500/10 text-purple-600 hover:bg-purple-500/15'
                : 'text-purple-500/70 hover:text-purple-600',
            )}
            onClick={() => navigate('/admin')}
            title="Admin"
          >
            <Shield className="h-4 w-4 sm:h-5 sm:w-5" />
          </Button>
        )}

        <Button
          variant="ghost"
          size="icon"
          className="h-9 w-9 rounded-lg text-muted-foreground hover:text-foreground sm:h-10 sm:w-10"
          onClick={handleLogout}
          title="Sign out"
        >
          <LogOut className="h-4 w-4 sm:h-5 sm:w-5" />
        </Button>
      </aside>

      <main className="ml-[76px] flex-1 sm:ml-24">
        {user?.name && showGreeting && (
          <div className={`mx-auto w-full max-w-5xl px-4 pt-6 sm:px-6 ${fading ? 'greeting-fade-out' : ''}`}>
            <div className="mb-4 rounded-lg border border-primary/20 bg-primary/5 p-3.5">
              <div className="flex items-center gap-2.5">
                <div className="flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-primary/10">
                  <BookUser className="h-4 w-4 text-primary" />
                </div>
                <p className="text-sm text-muted-foreground">
                  Hello, <span className="font-semibold text-foreground">{user.name}</span> — welcome back
                </p>
              </div>
            </div>
          </div>
        )}
        {children}
      </main>
    </div>
  )
}
