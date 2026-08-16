import { type ReactNode } from 'react'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { BookUser } from 'lucide-react'

interface AuthLayoutProps {
  title: string
  description: string
  children?: ReactNode
  footer?: ReactNode
}

export default function AuthLayout({ title, description, children, footer }: AuthLayoutProps) {
  return (
    <div className="flex min-h-screen items-center justify-center p-4">
      <Card className="w-full max-w-sm">
        <CardHeader className="space-y-1 text-center">
          <div className="mx-auto mb-2 flex h-12 w-12 items-center justify-center rounded-full">
            <BookUser className="h-9 w-9 text-emerald-600" />
          </div>
          <CardTitle className="text-xl">{title}</CardTitle>
          <CardDescription>{description}</CardDescription>
        </CardHeader>
        <CardContent>{children}</CardContent>
        {footer && (
          <div className="flex justify-center pb-6">
            {footer}
          </div>
        )}
      </Card>
    </div>
  )
}
