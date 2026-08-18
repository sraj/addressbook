import { forwardRef } from 'react'
import {
  TableRoot as NTable,
  TableHeader as NTableHeader,
  TableBody as NTableBody,
  TableRow as NTableRow,
  TableHead as NTableHead,
  TableCell as NTableCell,
  type TableRootProps,
  type TableHeaderProps,
  type TableBodyProps,
  type TableRowProps,
  type TableHeadProps,
  type TableCellProps,
} from '@mobentum/nebula-ui'
import { cn } from '@/lib/utils'

const Table = forwardRef<HTMLTableElement, TableRootProps>(({ className, ...props }, ref) => (
  <NTable ref={ref} className={className} {...props} />
))
Table.displayName = 'Table'

const TableHeader = forwardRef<HTMLTableSectionElement, TableHeaderProps>(({ className, ...props }, ref) => (
  <NTableHeader ref={ref} className={className} {...props} />
))
TableHeader.displayName = 'TableHeader'

const TableBody = forwardRef<HTMLTableSectionElement, TableBodyProps>(({ className, ...props }, ref) => (
  <NTableBody ref={ref} className={className} {...props} />
))
TableBody.displayName = 'TableBody'

const TableRow = forwardRef<HTMLTableRowElement, TableRowProps>(({ className, ...props }, ref) => (
  <NTableRow ref={ref} className={cn('hover:bg-muted/50', className)} {...props} />
))
TableRow.displayName = 'TableRow'

const TableHead = forwardRef<HTMLTableCellElement, TableHeadProps>(({ className, ...props }, ref) => (
  <NTableHead ref={ref} className={cn('h-10 px-2', className)} {...props} />
))
TableHead.displayName = 'TableHead'

const TableCell = forwardRef<HTMLTableCellElement, TableCellProps>(({ className, ...props }, ref) => (
  <NTableCell ref={ref} className={cn('p-2', className)} {...props} />
))
TableCell.displayName = 'TableCell'

export { Table, TableHeader, TableBody, TableRow, TableHead, TableCell }