import type { Contact, Note, Bookmark, User, UserPreferences, Collection, LabelOrder, LabelFormat } from '@/types'
import { AppError } from '@/lib/error'

const BASE = '/api/v1'

const SAFE_METHODS = new Set(['GET', 'HEAD', 'OPTIONS', 'TRACE'])

let csrfToken: string | null = null

async function getCSRFToken(): Promise<string> {
  if (csrfToken) return csrfToken
  const res = await fetch(`${BASE}/csrf`, { credentials: 'include' })
  if (res.ok) {
    const body = await res.json().catch(() => ({}))
    csrfToken = body.csrf_token || null
  }
  return csrfToken ?? ''
}

async function csrfHeader(): Promise<Record<string, string>> {
  const token = await getCSRFToken()
  return token ? { 'X-CSRF-Token': token } : {}
}

async function request<T>(url: string, options?: RequestInit): Promise<T> {
  const method = (options?.method ?? 'GET').toUpperCase()
  const headers: Record<string, string> = { 'Content-Type': 'application/json' }
  if (!SAFE_METHODS.has(method)) {
    Object.assign(headers, await csrfHeader())
  }
  const res = await fetch(`${BASE}${url}`, {
    credentials: 'include',
    headers,
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

async function upload<T>(url: string, formData: FormData): Promise<T> {
  const headers = await csrfHeader()
  const res = await fetch(`${BASE}${url}`, {
    method: 'POST',
    credentials: 'include',
    headers,
    body: formData,
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
  return res.json()
}

async function download(url: string): Promise<void> {
  const res = await fetch(`${BASE}${url}`, { credentials: 'include' })
  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }))
    throw new AppError(body.error || res.statusText, res.status, body.fields, body.request_id)
  }
  const blob = await res.blob()
  const disposition = res.headers.get('Content-Disposition') || ''
  let filename = 'contacts.csv'
  const match = disposition.match(/filename="?([^";]+)"?/)
  if (match) filename = match[1]
  const link = document.createElement('a')
  link.href = URL.createObjectURL(blob)
  link.download = filename
  document.body.appendChild(link)
  link.click()
  link.remove()
  URL.revokeObjectURL(link.href)
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

  importContacts: (file: File, format: 'csv' | 'xlsx', collectionId?: number) => {
    const fd = new FormData()
    fd.append('file', file)
    return upload<{ imported: number; skipped: number }>(
      `/contacts/import?format=${format}${collectionId ? `&collection_id=${collectionId}` : ''}`,
      fd,
    )
  },

  exportContacts: (format: 'csv' | 'xlsx', collectionId?: number) =>
    download(`/contacts/export?format=${format}${collectionId ? `&collection_id=${collectionId}` : ''}`),

  getCollections: () =>
    request<{ collections: Collection[] }>('/collections'),

  createCollection: (name: string) =>
    request<Collection>('/collections', {
      method: 'POST',
      body: JSON.stringify({ name }),
    }),

  getCollection: (id: number) =>
    request<Collection>(`/collections/${id}`),

  renameCollection: (id: number, name: string) =>
    request<Collection>(`/collections/${id}`, {
      method: 'PUT',
      body: JSON.stringify({ name }),
    }),

  deleteCollection: (id: number) =>
    request<void>(`/collections/${id}`, { method: 'DELETE' }),

  regenerateCollectionToken: (id: number) =>
    request<Collection>(`/collections/${id}/regenerate-token`, { method: 'POST' }),

  getCollectionContacts: (id: number, params: { page?: number; size?: number }) => {
    const sp = new URLSearchParams()
    sp.set('page', String(params.page ?? 1))
    sp.set('size', String(params.size ?? 20))
    return request<{ data: Contact[]; total: number; page: number; size: number; total_pages: number }>(
      `/collections/${id}/contacts?${sp.toString()}`
    )
  },

  addCollectionContact: (id: number, data: Omit<Contact, 'id' | 'user_id' | 'collection_id' | 'created_at' | 'updated_at'>) =>
    request<Contact>(`/collections/${id}/contacts`, {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  moveContactToCollection: (collectionId: number, contactId: number) =>
    request<void>(`/collections/${collectionId}/contacts/${contactId}`, { method: 'PUT' }),

  removeContactFromCollection: (collectionId: number, contactId: number) =>
    request<void>(`/collections/${collectionId}/contacts/${contactId}`, { method: 'DELETE' }),

  getInviteInfo: (token: string) =>
    request<{ name: string; token: string }>(`/invites/${token}`),

  submitInvite: (token: string, data: { name: string; email?: string; phone?: string; address: { label?: string; line1: string; line2?: string; city: string; state: string; zip: string; country: string } }) =>
    request<{ status: string; name: string }>(`/invites/${token}`, {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  getLabelOrders: () =>
    request<{ orders: LabelOrder[] }>('/labels/orders'),

  getLabelFormats: () =>
    request<{ formats: LabelFormat[] }>('/labels/formats'),

  createLabelOrder: (collectionId?: number, email?: string, format?: string) =>
    request<{ order: LabelOrder; url: string }>('/labels/order', {
      method: 'POST',
      body: JSON.stringify({ collection_id: collectionId ?? 0, email: email ?? '', format: format ?? '' }),
    }),

  confirmLabelOrder: (sessionId: string) =>
    request<LabelOrder>(`/labels/confirm?session_id=${encodeURIComponent(sessionId)}`, { method: 'POST' }),

  labelSheetUrl: (collectionId?: number, format?: string) =>
    `/labels/sheet${collectionId ? `?collection_id=${collectionId}` : ''}${format ? `${collectionId ? '&' : '?'}format=${format}` : ''}`,

  downloadLabelSheet: async (collectionId?: number, format?: string) => {
    const res = await fetch(`${BASE}/labels/sheet${collectionId ? `?collection_id=${collectionId}` : ''}${format ? `${collectionId ? '&' : '?'}format=${format}` : ''}`, {
      credentials: 'include',
    })
    if (!res.ok) {
      const body = await res.json().catch(() => ({ error: res.statusText }))
      throw new AppError(body.error || res.statusText, res.status, body.fields, body.request_id)
    }
    const html = await res.text()
    const blob = new Blob([html], { type: 'text/html;charset=utf-8' })
    const filename = `address-labels-${format || '5160'}.html`
    const link = document.createElement('a')
    link.href = URL.createObjectURL(blob)
    link.download = filename
    document.body.appendChild(link)
    link.click()
    link.remove()
    URL.revokeObjectURL(link.href)
  },

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
