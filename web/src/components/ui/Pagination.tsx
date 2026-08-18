import { PaginationRoot, PaginationList, PaginationItem, PaginationPrevious, PaginationNext, PaginationPage } from '@mobentum/nebula-ui'
import { ChevronLeft, ChevronRight } from 'lucide-react'

interface PaginationProps {
  page: number
  totalPages: number
  total: number
  onPageChange: (page: number) => void
}

export default function Pagination({
  page,
  totalPages,
  total,
  onPageChange,
}: PaginationProps) {
  return (
    <div className="mt-4 flex flex-col items-center gap-2 sm:flex-row sm:justify-between text-sm text-muted-foreground">
      <span className="hidden sm:inline">
        Page {page} of {totalPages} ({total} total)
      </span>
      <PaginationRoot className="mx-0 w-auto justify-end">
        <PaginationList>
          <PaginationItem>
            <PaginationPrevious
              disabled={page <= 1}
              onClick={() => onPageChange(page - 1)}
            >
              <ChevronLeft className="h-4 w-4" />
              <span className="hidden sm:inline">Previous</span>
            </PaginationPrevious>
          </PaginationItem>
          {Array.from({ length: Math.min(totalPages, 5) }, (_, i) => {
            const start = Math.max(1, page - 2)
            const p = start + i
            if (p > totalPages) return null
            return (
              <PaginationItem key={p}>
                <PaginationPage
                  className="min-w-[32px]"
                  aria-current={p === page ? 'page' : undefined}
                  onClick={() => onPageChange(p)}
                >
                  {p}
                </PaginationPage>
              </PaginationItem>
            )
          })}
          <PaginationItem>
            <PaginationNext
              disabled={page >= totalPages}
              onClick={() => onPageChange(page + 1)}
            >
              <span className="hidden sm:inline">Next</span>
              <ChevronRight className="h-4 w-4" />
            </PaginationNext>
          </PaginationItem>
        </PaginationList>
      </PaginationRoot>
    </div>
  )
}