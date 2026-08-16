import { toast } from '@/hooks/use-toast'
import type { ApiError } from '@/types'

export class AppError extends Error {
  status: number
  fields?: ApiError['fields']
  requestId?: string

  constructor(message: string, status: number, fields?: ApiError['fields'], requestId?: string) {
    super(message)
    this.name = 'AppError'
    this.status = status
    this.fields = fields
    this.requestId = requestId
  }
}

export function parseApiError(err: unknown): AppError {
  if (err instanceof AppError) return err

  const apiErr = err as { status?: number; error?: string; fields?: ApiError['fields']; request_id?: string }
  const status = apiErr.status ?? 500
  const message = apiErr.error ?? 'An unexpected error occurred'
  return new AppError(message, status, apiErr.fields, apiErr.request_id)
}

export function showErrorToast(err: unknown, title = 'Error') {
  const parsed = parseApiError(err)
  const description = parsed.requestId
    ? `${parsed.message} (ref: ${parsed.requestId})`
    : parsed.message
  toast({ title, description, variant: 'destructive' })
}

export function handleFieldErrors(
  err: unknown,
  setFieldErrors: (errors: Record<string, string>) => void,
  serverFieldToPath: (field: string) => string,
) {
  const parsed = parseApiError(err)
  if (parsed.fields && parsed.fields.length > 0) {
    const serverErrors: Record<string, string> = {}
    for (const f of parsed.fields) {
      const path = serverFieldToPath(f.field)
      serverErrors[path] = f.message || `${f.tag} validation failed`
    }
    setFieldErrors(serverErrors)
    return true
  }
  return false
}
