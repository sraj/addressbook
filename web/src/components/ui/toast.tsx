import { forwardRef } from 'react'
import {
  ToastRoot as NToast,
  ToastTitle as NToastTitle,
  ToastDescription as NToastDescription,
  ToastClose as NToastClose,
  type ToastRootProps,
  type ToastTitleProps,
  type ToastDescriptionProps,
  type ToastCloseProps,
} from '@mobentum/nebula-ui'
import { X } from 'lucide-react'
import { cn } from '@/lib/utils'

const Toast = forwardRef<HTMLDivElement, ToastRootProps>(({ className, ...props }, ref) => (
  <NToast ref={ref} className={cn(className)} {...props} />
))
Toast.displayName = 'Toast'

const ToastTitle = forwardRef<HTMLHeadingElement, ToastTitleProps>(({ className, ...props }, ref) => (
  <NToastTitle ref={ref} className={cn('text-sm font-semibold', className)} {...props} />
))
ToastTitle.displayName = 'ToastTitle'

const ToastDescription = forwardRef<HTMLParagraphElement, ToastDescriptionProps>(({ className, ...props }, ref) => (
  <NToastDescription ref={ref} className={cn('text-sm opacity-90', className)} {...props} />
))
ToastDescription.displayName = 'ToastDescription'

const ToastClose = forwardRef<HTMLButtonElement, ToastCloseProps>(({ className, ...props }, ref) => (
  <NToastClose
    ref={ref}
    className={cn(
      'absolute right-1 top-1 rounded-md p-1 text-foreground/50 opacity-0 transition-opacity hover:text-foreground focus:opacity-100 focus:outline-none focus:ring-1 group-hover:opacity-100',
      className,
    )}
    {...props}
  >
    <X className="h-4 w-4" />
  </NToastClose>
))
ToastClose.displayName = 'ToastClose'

export { Toast, ToastTitle, ToastDescription, ToastClose }