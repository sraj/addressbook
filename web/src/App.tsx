import { useEffect } from 'react'
import { Routes, Route, Navigate, Outlet, useLocation } from 'react-router-dom'
import { useAuthStore } from '@/store/auth'
import { useBillingStore } from '@/store/billing'
import AppShell from '@/components/layout/AppShell'
import LoginPage from '@/pages/LoginPage'
import RegisterPage from '@/pages/RegisterPage'
import ForgotPasswordPage from '@/pages/ForgotPasswordPage'
import ResetPasswordPage from '@/pages/ResetPasswordPage'
import ContactsPage from '@/pages/ContactsPage'
import NotesPage from '@/pages/NotesPage'
import BookmarksPage from '@/pages/BookmarksPage'
import CollectionsPage from '@/pages/CollectionsPage'
import CollectionDetailPage from '@/pages/CollectionDetailPage'
import InvitePage from '@/pages/InvitePage'
import LabelsPage from '@/pages/LabelsPage'
import AdminPage from '@/pages/AdminPage'
import ProfilePage from '@/pages/ProfilePage'
import SettingsPage from '@/pages/SettingsPage'
import GeneralSettings from '@/pages/settings/GeneralSettings'
import BillingSettings from '@/pages/settings/BillingSettings'
import NotFoundPage from '@/pages/NotFoundPage'

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated)
  if (!isAuthenticated) return <Navigate to="/login" replace />
  return <AppShell>{children}</AppShell>
}

function PublicRoute({ children }: { children: React.ReactNode }) {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated)
  if (isAuthenticated) return <Navigate to="/contacts" replace />
  return <>{children}</>
}

export default function App() {
  const location = useLocation()
  const initialized = useAuthStore((s) => s.initialized)
  const checkAuth = useAuthStore((s) => s.checkAuth)

  const publicPaths = ['/login', '/register', '/forgot-password', '/reset-password']
  const isPublicRoute = publicPaths.includes(location.pathname) || location.pathname.startsWith('/invite/')

  useEffect(() => {
    if (!initialized && !isPublicRoute) {
      checkAuth().then(() => {
        useBillingStore.getState().fetchUsage()
      })
    } else if (!initialized && isPublicRoute) {
      useAuthStore.setState({ initialized: true })
    }
  }, [])

  if (!initialized) return null

  return (
    <Routes>
      <Route element={<PublicRoute><Outlet /></PublicRoute>}>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />
        <Route path="/forgot-password" element={<ForgotPasswordPage />} />
        <Route path="/reset-password" element={<ResetPasswordPage />} />
      </Route>
      <Route element={<ProtectedRoute><Outlet /></ProtectedRoute>}>
        <Route path="/contacts" element={<ContactsPage />} />
        <Route path="/collections" element={<CollectionsPage />} />
        <Route path="/collections/:id" element={<CollectionDetailPage />} />
        <Route path="/labels" element={<LabelsPage />} />
        <Route path="/notes" element={<NotesPage />} />
        <Route path="/bookmarks" element={<BookmarksPage />} />
        <Route path="/profile" element={<ProfilePage />} />
        <Route path="/admin" element={<AdminPage />} />
        <Route path="/settings" element={<SettingsPage />}>
          <Route index element={<GeneralSettings />} />
          <Route path="billing" element={<BillingSettings />} />
        </Route>
      </Route>
      <Route path="/invite/:token" element={<InvitePage />} />
      <Route path="/" element={<Navigate to="/contacts" replace />} />
      <Route path="*" element={<NotFoundPage />} />
    </Routes>
  )
}
