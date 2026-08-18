import { forwardRef, type HTMLAttributes } from 'react'
import {
  AlertDialogRoot as NAlertDialogRoot,
  AlertDialogPortal as NAlertDialogPortal,
  AlertDialogPopup as NAlertDialogPopup,
  AlertDialogBackdrop as NAlertDialogBackdrop,
  AlertDialogTitle as NAlertDialogTitle,
  AlertDialogDescription as NAlertDialogDescription,
  AlertDialogCancel as NAlertDialogCancel,
  AlertDialogAction as NAlertDialogAction,
  type AlertDialogPopupProps,
  type AlertDialogTitleProps,
  type AlertDialogDescriptionProps,
  type AlertDialogCloseProps,
} from '@mobentum/nebula-ui'
import { cn } from '@/lib/utils'

const AlertDialog = NAlertDialogRoot

const AlertDialogContent = forwardRef<HTMLDivElement, AlertDialogPopupProps>(
  ({ className, children, ...props }, ref) => (
    <NAlertDialogPortal>
      <NAlertDialogBackdrop />
      <NAlertDialogPopup
        ref={ref}
        className={cn(
          'fixed left-1/2 top-1/2 z-50 grid w-full max-w-lg -translate-x-1/2 -translate-y-1/2 gap-4 border bg-background p-6 shadow-lg sm:rounded-lg',
          className,
        )}
        {...props}
      >
        {children}
      </NAlertDialogPopup>
    </NAlertDialogPortal>
  ),
)
AlertDialogContent.displayName = 'AlertDialogContent'

const AlertDialogHeader = ({ className, ...props }: HTMLAttributes<HTMLDivElement>) => (
  <div className={cn('flex flex-col space-y-2 text-center sm:text-left', className)} {...props} />
)
AlertDialogHeader.displayName = 'AlertDialogHeader'

const AlertDialogFooter = ({ className, ...props }: HTMLAttributes<HTMLDivElement>) => (
  <div className={cn('flex flex-col-reverse sm:flex-row sm:justify-end sm:space-x-2', className)} {...props} />
)
AlertDialogFooter.displayName = 'AlertDialogFooter'

const AlertDialogTitle = forwardRef<HTMLHeadingElement, AlertDialogTitleProps>(
  ({ className, ...props }, ref) => (
    <NAlertDialogTitle ref={ref} className={cn('text-lg font-semibold', className)} {...props} />
  ),
)
AlertDialogTitle.displayName = 'AlertDialogTitle'

const AlertDialogDescription = forwardRef<HTMLDivElement, AlertDialogDescriptionProps>(
  ({ className, ...props }, ref) => (
    <NAlertDialogDescription ref={ref} className={cn('text-sm text-muted-foreground', className)} {...props} />
  ),
)
AlertDialogDescription.displayName = 'AlertDialogDescription'

const AlertDialogAction = forwardRef<HTMLButtonElement, AlertDialogCloseProps>(
  ({ className, ...props }, ref) => (
    <NAlertDialogAction ref={ref} className={cn('h-9 px-4 py-2', className)} {...props} />
  ),
)
AlertDialogAction.displayName = 'AlertDialogAction'

const AlertDialogCancel = forwardRef<HTMLButtonElement, AlertDialogCloseProps>(
  ({ className, ...props }, ref) => (
    <NAlertDialogCancel ref={ref} className={cn('mt-2 h-9 px-4 py-2 sm:mt-0', className)} {...props} />
  ),
)
AlertDialogCancel.displayName = 'AlertDialogCancel'

export {
  AlertDialog,
  AlertDialogContent,
  AlertDialogHeader,
  AlertDialogFooter,
  AlertDialogTitle,
  AlertDialogDescription,
  AlertDialogAction,
  AlertDialogCancel,
}