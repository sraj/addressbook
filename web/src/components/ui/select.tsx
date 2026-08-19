import { Children, forwardRef, isValidElement, type ReactNode } from 'react'
import {
  SelectRoot as NSelect,
  SelectGroup as NSelectGroup,
  SelectValue as NSelectValue,
  SelectTrigger as NSelectTrigger,
  SelectPortal as NSelectPortal,
  SelectPositioner as NSelectPositioner,
  SelectPopup as NSelectPopup,
  SelectItem as NSelectItem,
  SelectItemIndicator as NSelectItemIndicator,
  type SelectRootProps,
  type SelectTriggerProps,
  type SelectValueProps,
  type SelectPopupProps,
  type SelectItemProps,
} from '@mobentum/nebula-ui'
import { Check, ChevronDown } from 'lucide-react'
import { cn } from '@/lib/utils'

function collectItemLabels(children: ReactNode, out: Record<string, ReactNode> = {}): Record<string, ReactNode> {
  Children.forEach(children, (child) => {
    if (!isValidElement(child)) return
    const childProps = child.props as { value?: string; children?: ReactNode }
    if (typeof childProps.value === 'string') {
      out[childProps.value] = childProps.children
    }
    if (childProps.children) {
      collectItemLabels(childProps.children, out)
    }
  })
  return out
}

type SelectProps = Omit<SelectRootProps, 'onValueChange'> & {
  onValueChange?: (value: string) => void
}

const Select = ({ onValueChange, children, ...props }: SelectProps) => (
  <NSelect
    items={collectItemLabels(children)}
    onValueChange={(value) => {
      if (value !== null) onValueChange?.(String(value))
    }}
    {...props}
  >
    {children}
  </NSelect>
)
Select.displayName = 'Select'

const SelectGroup = NSelectGroup

const SelectValue = forwardRef<HTMLSpanElement, SelectValueProps>(
  ({ className, ...props }, ref) => (
    <NSelectValue ref={ref} className={cn('placeholder:text-muted-foreground', className)} {...props} />
  ),
)
SelectValue.displayName = 'SelectValue'

const SelectTrigger = forwardRef<HTMLButtonElement, SelectTriggerProps>(
  ({ className, children, ...props }, ref) => (
    <NSelectTrigger
      ref={ref}
      className={cn(
        'flex h-9 w-full items-center justify-between whitespace-nowrap rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring disabled:cursor-not-allowed disabled:opacity-50 [&>span]:line-clamp-1',
        className,
      )}
      {...props}
    >
      {children}
      <ChevronDown className="h-4 w-4 opacity-50" />
    </NSelectTrigger>
  ),
)
SelectTrigger.displayName = 'SelectTrigger'

const SelectContent = forwardRef<HTMLDivElement, SelectPopupProps>(
  ({ className, children, ...props }, ref) => (
    <NSelectPortal>
      <NSelectPositioner>
        <NSelectPopup
          ref={ref}
          className={cn(
            'z-50 max-h-96 min-w-[8rem] overflow-hidden rounded-md border border-border bg-popover p-1 text-popover-foreground shadow-md',
            className,
          )}
          {...props}
        >
          {children}
        </NSelectPopup>
      </NSelectPositioner>
    </NSelectPortal>
  ),
)
SelectContent.displayName = 'SelectContent'

const SelectItem = forwardRef<HTMLDivElement, SelectItemProps>(
  ({ className, children, ...props }, ref) => (
    <NSelectItem
      ref={ref}
      className={cn(
        'relative flex w-full cursor-default select-none items-center rounded-sm py-1.5 pl-2 pr-8 text-sm outline-none focus:bg-accent focus:text-accent-foreground data-highlighted:bg-accent data-highlighted:text-accent-foreground data-disabled:pointer-events-none data-disabled:opacity-50',
        className,
      )}
      {...props}
    >
      <span className="absolute right-2 flex h-3.5 w-3.5 items-center justify-center">
        <NSelectItemIndicator>
          <Check className="h-4 w-4" />
        </NSelectItemIndicator>
      </span>
      {children}
    </NSelectItem>
  ),
)
SelectItem.displayName = 'SelectItem'

export {
  Select,
  SelectGroup,
  SelectValue,
  SelectTrigger,
  SelectContent,
  SelectItem,
}