import type { Contact, Note, Bookmark, User, UserPreferences } from '@/types'
import { AppError } from '@/lib/error'

const BASE = '/api/v1'

async function request<T>(url: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${url}`, {
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    ...options,
  })

  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }))
    throw new AppError(
      body.error || res.statusText,
      res.status,
      body.fields,
      body.request_id,
    )
  }

  if (res.status === 204) {
    return undefined as T
  }

  return res.json()
}

export const api = {
  login: (email: string, password: string) =>
    request<{ user: User }>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    }),

  register: (email: string, password: string) =>
    request<{ user: User }>('/auth/register', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    }),

  me: () => request<{ user: User }>('/auth/me'),

  logout: () => request<void>('/auth/logout', { method: 'POST' }),

  forgotPassword: (email: string) =>
    request<{ message: string }>('/auth/forgot-password', {
      method: 'POST',
      body: JSON.stringify({ email }),
    }),

  resetPassword: (token: string, password: string) =>
    request<{ message: string }>('/auth/reset-password', {
      method: 'POST',
      body: JSON.stringify({ token, password }),
    }),

  changePassword: (currentPassword: string, newPassword: string) =>
    request<{ message: string }>('/auth/change-password', {
      method: 'POST',
      body: JSON.stringify({ current_password: currentPassword, new_password: newPassword }),
    }),

  updateProfile: (name: string, preferences?: Partial<UserPreferences>) =>
    request<{ user: User }>('/auth/profile', {
      method: 'PUT',
      body: JSON.stringify({ name, preferences }),
    }),

  getContacts: (params: { q?: string; page?: number; size?: number }) => {
    const sp = new URLSearchParams()
    if (params.q) sp.set('q', params.q)
    sp.set('page', String(params.page ?? 1))
    sp.set('size', String(params.size ?? 20))
    return request<{ data: Contact[]; total: number; page: number; size: number; total_pages: number }>(
      `/contacts?${sp.toString()}`
    )
  },

  createContact: (data: Omit<Contact, 'id' | 'user_id' | 'created_at' | 'updated_at'>) =>
    request<Contact>('/contacts', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  updateContact: (id: number, data: Omit<Contact, 'id' | 'user_id' | 'created_at' | 'updated_at'>) =>
    request<Contact>(`/contacts/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  deleteContact: (id: number) =>
    request<void>(`/contacts/${id}`, { method: 'DELETE' }),

  getNotes: (params: { page?: number; size?: number }) => {
    const sp = new URLSearchParams()
    sp.set('page', String(params.page ?? 1))
    sp.set('size', String(params.size ?? 20))
    return request<{ data: Note[]; total: number; page: number; size: number; total_pages: number }>(
      `/notes?${sp.toString()}`
    )
  },

  getNote: (id: number) =>
    request<Note>(`/notes/${id}`),

  createNote: (data: { title: string; content: string }) =>
    request<Note>('/notes', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  updateNote: (id: number, data: { title: string; content: string }) =>
    request<Note>(`/notes/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  deleteNote: (id: number) =>
    request<void>(`/notes/${id}`, { method: 'DELETE' }),

  getBillingUsage: () =>
    request<{ usage: Record<string, number>; plan: { id: number; name: string; price_monthly: number }; limits: Record<string, number> }>('/billing/usage'),

  getPlans: () =>
    request<{ plans: Array<{ id: number; name: string; price_monthly: number; stripe_price_id: string }> }>('/billing/plans'),

  createCheckoutSession: (plan?: string) =>
    request<{ url: string }>(`/billing/checkout${plan ? `?plan=${plan}` : ''}`, { method: 'POST' }),

  createPortalSession: () =>
    request<{ url: string }>('/billing/portal', { method: 'POST' }),

  getAdminPlans: () =>
    request<{ plans: Array<{ id: number; name: string; price_monthly: number; stripe_price_id: string }> }>('/admin/plans'),

  updatePlanPriceID: (id: number, stripe_price_id: string) =>
    request<{ status: string }>(`/admin/plans/${id}/price-id`, {
      method: 'PUT',
      body: JSON.stringify({ stripe_price_id }),
    }),

  syncPlansFromStripe: () =>
    request<{ status: string }>('/admin/plans/sync-prices', { method: 'POST' }),

  cancelSubscription: () =>
    request<{ status: string }>('/billing/cancel', { method: 'POST' }),

  changePlan: (plan: string) =>
    request<{ status: string; plan: string }>(`/billing/change-plan?plan=${plan}`, { method: 'POST' }),

  getInvoices: () =>
    request<Array<{ id: string; amount_paid: number; currency: string; status: string; created: number; period_start: number; period_end: number; hosted_invoice_url: string; invoice_pdf: string; number: string }>>('/billing/invoices'),

  getAdminUsers: () =>
    request<{ users: Array<{ id: number; email: string; role: string; status: string; created_at: string; plan_name: string; subscription_status: string; subscription_end: string }> }>('/admin/users'),

  updateUserStatus: (id: number, status: string) =>
    request<{ status: string }>(`/admin/users/${id}/status?status=${status}`, { method: 'PUT' }),

  getBookmarks: (params: { category?: string; page?: number; size?: number }) => {
    const sp = new URLSearchParams()
    if (params.category) sp.set('category', params.category)
    sp.set('page', String(params.page ?? 1))
    sp.set('size', String(params.size ?? 20))
    return request<{ data: Bookmark[]; total: number; page: number; size: number; total_pages: number }>(
      `/bookmarks?${sp.toString()}`
    )
  },

  getBookmark: (id: number) =>
    request<Bookmark>(`/bookmarks/${id}`),

  createBookmark: (data: { url: string; title: string; description?: string; favicon_url?: string; category?: string }) =>
    request<Bookmark>('/bookmarks', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  updateBookmark: (id: number, data: { url: string; title: string; description?: string; favicon_url?: string; category?: string }) =>
    request<Bookmark>(`/bookmarks/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  deleteBookmark: (id: number) =>
    request<void>(`/bookmarks/${id}`, { method: 'DELETE' }),

  importBookmarks: (bookmarks: Array<{ url: string; title: string; description?: string; favicon_url?: string; category?: string }>) =>
    request<{ imported: number; skipped: number }>('/bookmarks/import', {
      method: 'POST',
      body: JSON.stringify({ bookmarks }),
    }),

  importBookmarkHTML: (html: string) =>
    request<{ imported: number; skipped: number }>('/bookmarks/import-html', {
      method: 'POST',
      body: JSON.stringify({ html }),
    }),

  getBookmarkCategories: () =>
    request<{ categories: string[] }>('/bookmarks/categories'),
}
