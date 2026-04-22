"use client"

import * as React from "react"
import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight } from "lucide-react"

import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"

export interface PaginationProps extends React.ComponentPropsWithoutRef<"nav"> {
  page: number
  pageCount: number
  onPageChange: (page: number) => void
  siblingCount?: number
}

function Pagination({
  page,
  pageCount,
  onPageChange,
  siblingCount = 1,
  className,
  ...props
}: PaginationProps) {
  const range = React.useMemo(() => {
    return buildPageRange(page, pageCount, siblingCount)
  }, [page, pageCount, siblingCount])

  if (pageCount <= 1) return null

  return (
    <nav
      data-slot="pagination"
      className={cn("flex items-center gap-1 font-sans", className)}
      aria-label="Pagination"
      {...props}
    >
      <Button
        variant="ghost"
        size="icon-xs"
        onClick={() => onPageChange(1)}
        disabled={page <= 1}
        aria-label="First page"
      >
        <ChevronsLeft className="size-3" />
      </Button>
      <Button
        variant="ghost"
        size="icon-xs"
        onClick={() => onPageChange(page - 1)}
        disabled={page <= 1}
        aria-label="Previous page"
      >
        <ChevronLeft className="size-3" />
      </Button>

      {range.map((item, index) => {
        if (item === "ellipsis") {
          return (
            <span key={`ellipsis-${index}`} className="flex size-6 items-center justify-center text-xs text-muted-foreground">
              &hellip;
            </span>
          )
        }
        const pageNum = item as number
        return (
          <Button
            key={pageNum}
            variant={page === pageNum ? "secondary" : "ghost"}
            size="icon-xs"
            onClick={() => onPageChange(pageNum)}
            aria-label={`Page ${pageNum}`}
            aria-current={page === pageNum ? "page" : undefined}
          >
            {pageNum}
          </Button>
        )
      })}

      <Button
        variant="ghost"
        size="icon-xs"
        onClick={() => onPageChange(page + 1)}
        disabled={page >= pageCount}
        aria-label="Next page"
      >
        <ChevronRight className="size-3" />
      </Button>
      <Button
        variant="ghost"
        size="icon-xs"
        onClick={() => onPageChange(pageCount)}
        disabled={page >= pageCount}
        aria-label="Last page"
      >
        <ChevronsRight className="size-3" />
      </Button>
    </nav>
  )
}

function buildPageRange(current: number, total: number, siblingCount: number): (number | "ellipsis")[] {
  const delta = siblingCount * 2 + 5

  if (total <= delta + 2) {
    return Array.from({ length: total }, (_, i) => i + 1)
  }

  const pages: (number | "ellipsis")[] = []

  pages.push(1)

  if (current - siblingCount > 2) {
    pages.push("ellipsis")
  }

  const leftBound = Math.max(2, current - siblingCount)
  const rightBound = Math.min(total - 1, current + siblingCount)

  for (let i = leftBound; i <= rightBound; i++) {
    pages.push(i)
  }

  if (current + siblingCount < total - 1) {
    pages.push("ellipsis")
  }

  pages.push(total)

  return pages
}

export { Pagination }
