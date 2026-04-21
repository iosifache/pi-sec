"use client"

import * as React from "react"
import Image from "next/image"
import Link from "next/link"
import {
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getSortedRowModel,
  useReactTable,
  type ColumnDef,
  type FilterFn,
  type SortingState,
} from "@tanstack/react-table"
import {
  Activity,
  AlertTriangle,
  ArrowDown,
  ArrowUp,
  ArrowUpDown,
  Bug,
  CircleCheck,
  Clock3,
  ExternalLink,
  FileText,
  FileWarning,
  GitBranch,
  GitCommitHorizontal,
  GitFork,
  Home,
  Link2Off,
  Package,
  Scale,
  Search,
  ShieldAlert,
  Star,
  TrendingDown,
  TrendingUp,
  Users,
} from "lucide-react"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { ScrollArea } from "@/components/ui/scroll-area"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip"
import { cn } from "@/lib/utils"
import type { AlertSeverity, PackageTableRow, PackagesDashboardData } from "@/lib/types"

const globalRowFilter: FilterFn<PackageTableRow> = (row, _columnId, value) => {
  const query = String(value ?? "").trim().toLowerCase()
  if (!query) return true
  return row.original.searchText.includes(query)
}

type ConcernFilterId =
  | "missingGithub"
  | "noLicense"
  | "maintainerOrOwnerRisk"
  | "ageOrFreshnessRisk"
  | "maintenanceRisk"
  | "lowStars"
  | "lowForks"
  | "lowWatches"
  | "lowAdoption"
  | "downloadSpike"
  | "missingTrustLinks"
  | "missingDescription"
  | "otherAlert"

type ConcernLegendItem = {
  id: ConcernFilterId
  label: string
  icon: React.ComponentType<{ className?: string }>
  severity: AlertSeverity
  matches: (alertId: string) => boolean
}

const alertLegendItems: ConcernLegendItem[] = [
  {
    id: "missingGithub",
    label: "Missing GitHub metadata",
    icon: GitBranch,
    severity: "critical",
    matches: (alertId) => alertId.includes("missing_github") || alertId.includes("github_repository"),
  },
  {
    id: "noLicense",
    label: "No license",
    icon: Scale,
    severity: "high",
    matches: (alertId) => alertId.includes("no_license"),
  },
  {
    id: "maintainerOrOwnerRisk",
    label: "Maintainer or owner risk",
    icon: Users,
    severity: "high",
    matches: (alertId) => alertId.includes("maintainer") || alertId.includes("owner_track_record") || alertId.includes("owner_reputation"),
  },
  {
    id: "ageOrFreshnessRisk",
    label: "Age or freshness risk",
    icon: Clock3,
    severity: "medium",
    matches: (alertId) => alertId.includes("very_new") || alertId.includes("new_owner_account") || alertId.includes("stale") || alertId.includes("age_mismatch"),
  },
  {
    id: "maintenanceRisk",
    label: "Commit, release, or maintenance risk",
    icon: GitCommitHorizontal,
    severity: "medium",
    matches: (alertId) => alertId.includes("commit_history") || alertId.includes("no_releases") || alertId.includes("low_maintenance"),
  },
  {
    id: "lowStars",
    label: "Low stars",
    icon: Star,
    severity: "medium",
    matches: (alertId) => alertId.includes("low_stars"),
  },
  {
    id: "lowForks",
    label: "Low forks",
    icon: GitFork,
    severity: "low",
    matches: (alertId) => alertId.includes("low_forks"),
  },
  {
    id: "lowWatches",
    label: "Low watches",
    icon: Activity,
    severity: "low",
    matches: (alertId) => alertId.includes("low_watches"),
  },
  {
    id: "lowAdoption",
    label: "Low adoption",
    icon: TrendingDown,
    severity: "medium",
    matches: (alertId) => alertId.includes("low_weekly_downloads") || alertId.includes("low_monthly_downloads") || alertId.includes("weak_consistency"),
  },
  {
    id: "downloadSpike",
    label: "Download spike",
    icon: TrendingUp,
    severity: "high",
    matches: (alertId) => alertId.includes("download_spike"),
  },
  {
    id: "missingTrustLinks",
    label: "Missing trust links",
    icon: Link2Off,
    severity: "critical",
    matches: (alertId) => alertId.includes("broken_trust_link_surface"),
  },
  {
    id: "missingDescription",
    label: "Missing description",
    icon: FileText,
    severity: "medium",
    matches: (alertId) => alertId.includes("missing_description"),
  },
  {
    id: "otherAlert",
    label: "Other alert",
    icon: AlertTriangle,
    severity: "medium",
    matches: (alertId) => {
      const trimmed = alertId.trim()
      return Boolean(trimmed) && !alertLegendItems.slice(0, -1).some((item) => item.matches(trimmed))
    },
  },
]


export function PackagesTable({ data }: { data: PackagesDashboardData }) {
  const [sorting, setSorting] = React.useState<SortingState>([
    { id: "downloadsMonthly", desc: true },
  ])
  const [globalFilter, setGlobalFilter] = React.useState("")
  const [activeRow, setActiveRow] = React.useState<PackageTableRow | null>(null)
  const [hiddenConcernFilters, setHiddenConcernFilters] = React.useState<ConcernFilterId[]>([])

  const matchedGithubCount = React.useMemo(
    () => data.rows.filter((row) => row.hasGithubData).length,
    [data.rows]
  )

  const filteredSourceRows = React.useMemo(() => {
    return data.rows.filter((row) => {
      const hiddenByConcern = hiddenConcernFilters.some((filterId) => {
        const concern = alertLegendItems.find((item) => item.id === filterId)
        return concern ? row.alerts.some((alert) => concern.matches(alert.id)) : false
      })

      return !hiddenByConcern
    })
  }, [data.rows, hiddenConcernFilters])

  const columns = React.useMemo<ColumnDef<PackageTableRow>[]>(
    () => [
      {
        accessorKey: "packageName",
        header: ({ column }) => sortableHeader(column, "Package"),
        cell: ({ row }) => {
          const packageRow = row.original

          return (
            <div className="flex min-w-52 items-start justify-between gap-3">
              <div className="min-w-0 space-y-1">
                <div className="flex items-center gap-2 font-semibold text-foreground">
                  <span className="truncate">{packageRow.packageName || "—"}</span>
                  {packageRow.npmUrl ? (
                    <Link
                      href={packageRow.npmUrl}
                      target="_blank"
                      rel="noreferrer"
                      className="text-muted-foreground transition-colors hover:text-foreground"
                      aria-label={`Open ${packageRow.packageName} on npm`}
                      title="Open on npm"
                      onClick={(event) => event.stopPropagation()}
                    >
                      <ExternalLink className="size-3.5" />
                    </Link>
                  ) : null}
                </div>
              </div>
            </div>
          )
        },
      },
      {
        id: "concerns",
        header: "Concerns",
        enableSorting: false,
        cell: ({ row }) => {
          const packageRow = row.original
          const sortedAlerts = [...packageRow.alerts].sort((a, b) => severityRank(b.severity) - severityRank(a.severity))

          return (
            <div className="flex min-w-20 flex-wrap items-center gap-2 text-red-600 dark:text-red-400">
              {sortedAlerts.length ? (
                sortedAlerts.map((alert) => {
                  const Icon = iconForAlert(alert.id)
                  return (
                    <InlineStatusIcon
                      key={alert.id}
                      label={alert.name}
                      icon={Icon}
                      className={severityIconClass(alert.severity)}
                    />
                  )
                })
              ) : (
                <InlineStatusIcon
                  label="No concerns found"
                  icon={CircleCheck}
                  className="text-emerald-600 hover:text-emerald-700 focus-visible:text-emerald-700 dark:text-emerald-400 dark:hover:text-emerald-300 dark:focus-visible:text-emerald-300"
                />
              )}
            </div>
          )
        },
      },
      {
        accessorKey: "description",
        header: ({ column }) => sortableHeader(column, "Description"),
        cell: ({ row }) => (
          <div className="min-w-72 max-w-[36rem] whitespace-normal font-sans leading-5 text-muted-foreground">
            {display(row.original.description)}
          </div>
        ),
      },
      {
        accessorKey: "version",
        header: ({ column }) => sortableHeader(column, "Version"),
        cell: ({ row }) => <div className="min-w-24 text-muted-foreground">{display(row.original.version)}</div>,
      },
      {
        accessorKey: "downloadsMonthly",
        header: ({ column }) => sortableHeader(column, "Monthly npm downloads"),
        cell: ({ row }) => (
          <div className="min-w-32 font-medium text-foreground">{displayNumber(row.original.downloadsMonthly)}</div>
        ),
      },
      {
        accessorKey: "githubStars",
        header: ({ column }) => sortableHeader(column, "GitHub stars"),
        cell: ({ row }) => <div className="min-w-28 text-muted-foreground">{displayNumber(row.original.githubStars)}</div>,
      },
      {
        id: "details",
        header: "Details",
        enableSorting: false,
        cell: ({ row }) => (
          <Button
            variant="ghost"
            size="sm"
            className="font-sans"
            onClick={(event) => {
              event.stopPropagation()
              setActiveRow(row.original)
            }}
          >
            Open
          </Button>
        ),
      },
    ],
    []
  )

  const table = useReactTable({
    data: filteredSourceRows,
    columns,
    state: { sorting, globalFilter },
    onSortingChange: setSorting,
    onGlobalFilterChange: setGlobalFilter,
    globalFilterFn: globalRowFilter,
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getSortedRowModel: getSortedRowModel(),
  })

  const visibleRows = table.getRowModel().rows
  const githubCoverage = data.rowCount ? Math.round((matchedGithubCount / data.rowCount) * 100) : 0
  const npmFreshness = formatCacheFileAge(data.npmCacheFile)
  const githubFreshness = data.githubCacheFile ? formatCacheFileAge(data.githubCacheFile) : "Unavailable"

  function toggleConcernFilter(filterId: ConcernFilterId) {
    setHiddenConcernFilters((current) =>
      current.includes(filterId)
        ? current.filter((value) => value !== filterId)
        : [...current, filterId]
    )
  }

  return (
    <>
      <div className="flex h-svh flex-col overflow-hidden bg-background text-foreground">
        <header className="sticky top-0 z-20 bg-background/95 backdrop-blur">
          <div className="p-4">
            <div className="border bg-background p-4">
              <div className="flex flex-wrap items-center justify-between gap-4">
                <div className="min-w-0 flex-1">
                  <Image
                    src="/logo.png"
                    alt="Spigot"
                    width={240}
                    height={32}
                    className="h-8 w-auto"
                    priority
                  />
                  <p className="font-sans text-sm text-muted-foreground">
                    Discover and assess <a href="https://pi.dev/" target="_blank" rel="noreferrer" className="underline underline-offset-4 hover:text-foreground">Pi Agent</a> packages.
                  </p>
                </div>

                <div className="w-full self-center md:w-96">
                  <div className="relative">
                    <Search className="pointer-events-none absolute top-1/2 left-2.5 size-3.5 -translate-y-1/2 text-muted-foreground" />
                    <Input
                      value={globalFilter}
                      onChange={(event) => setGlobalFilter(event.target.value)}
                      placeholder="Search packages, descriptions, maintainers, repos, owners..."
                      className="pl-8 font-sans"
                    />
                  </div>
                </div>
              </div>

              <div className="mt-4 flex flex-wrap items-center justify-between gap-3 font-sans text-sm text-muted-foreground">
                <div>
                  Community project <span aria-hidden="true">·</span>{" "}
                  <Link
                    href="https://github.com/iosifache/pi-sec"
                    target="_blank"
                    rel="noreferrer"
                    className="underline underline-offset-4 hover:text-foreground"
                  >
                    Source code
                  </Link>
                </div>

                <div>
                  {data.rowCount} total <span aria-hidden="true">·</span> {visibleRows.length} shown <span aria-hidden="true">·</span> {githubCoverage}% GitHub coverage <span aria-hidden="true">·</span> {githubFreshness} GitHub cache <span aria-hidden="true">·</span> {npmFreshness} npm cache
                </div>
              </div>
            </div>
          </div>

          <div className="px-4 pb-4">
            <div className="overflow-x-scroll border bg-background px-4 py-3">
              <div className="flex min-w-max items-center gap-5">
                <span className="font-sans text-xs font-semibold text-muted-foreground">Concerns filters</span>
                {alertLegendItems.map((item) => {
                  const Icon = item.icon
                  const hidden = hiddenConcernFilters.includes(item.id)
                  return (
                    <button
                      key={item.id}
                      type="button"
                      className={cn(
                        "inline-flex items-center gap-2 font-sans text-xs outline-none transition-colors hover:text-foreground focus-visible:text-foreground",
                        hidden ? "text-muted-foreground" : "text-foreground"
                      )}
                      onClick={() => toggleConcernFilter(item.id)}
                      aria-pressed={hidden}
                      title={hidden ? `Show packages with ${item.label.toLowerCase()}` : `Hide packages with ${item.label.toLowerCase()}`}
                    >
                      <Icon className={cn("size-3.5", hidden ? "text-muted-foreground/50" : severityIconClass(item.severity))} />
                      <span>{item.label}</span>
                    </button>
                  )
                })}
              </div>
            </div>
          </div>
        </header>

        <main className="flex min-h-0 flex-1 flex-col p-4 pb-6">
          <div className="min-h-0 flex-1 overflow-hidden border bg-background">
            <ScrollArea className="h-full w-full">
              <Table>
                <TableHeader className="sticky top-0 z-10 bg-background">
                  {table.getHeaderGroups().map((headerGroup) => (
                    <TableRow key={headerGroup.id}>
                      {headerGroup.headers.map((header) => (
                        <TableHead key={header.id}>
                          {header.isPlaceholder
                            ? null
                            : flexRender(header.column.columnDef.header, header.getContext())}
                        </TableHead>
                      ))}
                    </TableRow>
                  ))}
                </TableHeader>
                <TableBody>
                  {visibleRows.length ? (
                    visibleRows.map((row) => (
                      <TableRow
                        key={row.id}
                        className="cursor-pointer"
                        onClick={() => setActiveRow(row.original)}
                      >
                        {row.getVisibleCells().map((cell) => (
                          <TableCell key={cell.id}>
                            {flexRender(cell.column.columnDef.cell, cell.getContext())}
                          </TableCell>
                        ))}
                      </TableRow>
                    ))
                  ) : (
                    <TableRow>
                      <TableCell colSpan={columns.length} className="h-24 text-center font-sans text-muted-foreground">
                        No packages match your search. Try a package name, maintainer, repo, owner, or license.
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </ScrollArea>
          </div>
        </main>
      </div>

      <DetailsDialog row={activeRow} onClose={() => setActiveRow(null)} />
    </>
  )
}

function DetailsDialog({
  row,
  onClose,
}: {
  row: PackageTableRow | null
  onClose: () => void
}) {
  return (
    <Dialog open={Boolean(row)} onOpenChange={(open) => !open && onClose()}>
      {row ? (
        <DialogContent>
          <DialogHeader className="border-b px-5 py-4 pr-12">
            <DialogTitle>{row.packageName || "Package details"}</DialogTitle>
            <DialogDescription className="font-sans">
              Core package metadata, npm metrics, and GitHub repository details.
            </DialogDescription>
          </DialogHeader>

          <DialogBody>
            <ScrollArea className="h-[min(70vh,42rem)] w-full">
              <div className="space-y-5 px-5 py-5">
                <DetailsSection
                  title="Metadata"
                  items={[
                    { label: "Package", value: display(row.packageName) },
                    { label: "Description", value: display(row.description), wide: true, prose: true },
                    { label: "License", value: display(row.license), missing: isMissingText(row.license) },
                    { label: "Maintainers", value: formatMaintainers(row), wide: true, prose: true },
                    {
                      label: "Homepage",
                      value: renderLink(row.homepageUrl, "Open homepage", Home),
                      missing: isMissingText(row.homepageUrl),
                    },
                    {
                      label: "Repository",
                      value: renderLink(normalizeExternalUrl(row.repoUrl), "Open repository", GitBranch),
                      missing: isMissingText(row.repoUrl),
                    },
                    {
                      label: "Issue tracker",
                      value: renderLink(row.bugsUrl, "Open issue tracker", Bug),
                      missing: isMissingText(row.bugsUrl),
                    },
                    {
                      label: "npm page",
                      value: renderLink(row.npmUrl, "Open npm page", Package),
                      missing: isMissingText(row.npmUrl),
                    },
                  ]}
                />

                <DetailsSection
                  title="NPM"
                  items={[
                    { label: "Version", value: display(row.version) },
                    {
                      label: "Monthly downloads",
                      value: displayNumber(row.downloadsMonthly),
                      missing: isMissingNumber(row.downloadsMonthly),
                    },
                    {
                      label: "Weekly downloads",
                      value: displayNumber(row.downloadsWeekly),
                      missing: isMissingNumber(row.downloadsWeekly),
                    },
                    { label: "Updated", value: formatDate(row.npmUpdatedAt), missing: isMissingText(row.npmUpdatedAt) },
                    { label: "Created", value: formatDate(row.npmCreatedAt), missing: isMissingText(row.npmCreatedAt) },
                  ]}
                />

                <DetailsSection
                  title="GitHub repository"
                  items={[
                    {
                      label: "GitHub data",
                      value: row.hasGithubData ? "Available" : "Missing",
                      missing: !row.hasGithubData,
                    },
                    {
                      label: "Repository name",
                      value: display(row.githubFullName),
                      missing: isMissingText(row.githubFullName),
                    },
                    { label: "Stars", value: displayNumber(row.githubStars), missing: isMissingNumber(row.githubStars) },
                    { label: "Forks", value: displayNumber(row.githubForks), missing: isMissingNumber(row.githubForks) },
                    { label: "Watches", value: displayNumber(row.githubWatches), missing: isMissingNumber(row.githubWatches) },
                    {
                      label: "Releases",
                      value: displayNumber(row.githubReleasesCount),
                      missing: isMissingNumber(row.githubReleasesCount),
                    },
                    {
                      label: "Commits",
                      value: displayNumber(row.githubCommitsCount),
                      missing: isMissingNumber(row.githubCommitsCount),
                    },
                    {
                      label: "Repository age",
                      value: displayDays(row.githubRepoAgeDays),
                      missing: isMissingNumber(row.githubRepoAgeDays),
                    },
                  ]}
                />

                <DetailsSection
                  title="GitHub owner"
                  items={[
                    { label: "Owner", value: display(row.githubOwnerLogin), missing: isMissingText(row.githubOwnerLogin) },
                    { label: "Owner type", value: display(row.githubOwnerType), missing: isMissingText(row.githubOwnerType) },
                    {
                      label: "Followers",
                      value: displayNumber(row.githubOwnerFollowers),
                      missing: isMissingNumber(row.githubOwnerFollowers),
                    },
                    {
                      label: "Owner age",
                      value: displayDays(row.githubOwnerAgeDays),
                      missing: isMissingNumber(row.githubOwnerAgeDays),
                    },
                    {
                      label: "Total repositories",
                      value: displayNumber(row.githubOwnerTotalRepos),
                      missing: isMissingNumber(row.githubOwnerTotalRepos),
                    },
                  ]}
                />

                <AlertsSection row={row} />
              </div>
            </ScrollArea>
          </DialogBody>

          <DialogFooter className="border-t px-5 py-4">
            <Button variant="outline" onClick={onClose} className="font-sans">
              Close
            </Button>
          </DialogFooter>
        </DialogContent>
      ) : null}
    </Dialog>
  )
}

function AlertsSection({ row }: { row: PackageTableRow }) {
  return (
    <section className="space-y-3 border p-4">
      <div className="flex items-center justify-between gap-3">
        <h3 className="font-sans text-xs font-semibold uppercase tracking-[0.12em] text-muted-foreground">
          Alerts
        </h3>
        <span className="font-sans text-xs text-muted-foreground">
          {row.alerts.length ? `${row.alerts.length} triggered` : "No alerts triggered"}
        </span>
      </div>

      {row.alerts.length ? (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Description</TableHead>
              <TableHead>Severity</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {row.alerts.map((alert) => (
              <TableRow key={alert.id}>
                <TableCell className="align-top font-medium whitespace-normal">{alert.name}</TableCell>
                <TableCell className="max-w-0 align-top whitespace-normal font-sans leading-6 text-muted-foreground">
                  {alert.description}
                </TableCell>
                <TableCell className="align-top">
                  <SeverityBadge severity={alert.severity} />
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      ) : (
        <div className="border border-dashed p-4 font-sans text-sm text-muted-foreground">
          This package did not trigger any stored risk alerts.
        </div>
      )}
    </section>
  )
}

function SeverityBadge({ severity }: { severity: AlertSeverity }) {
  const variant =
    severity === "critical" || severity === "high"
      ? "destructive"
      : severity === "medium"
        ? "secondary"
        : "outline"

  return (
    <Badge variant={variant} className="uppercase">
      {severity}
    </Badge>
  )
}

function DetailsSection({
  title,
  items,
}: {
  title: string
  items: Array<{ label: string; value: React.ReactNode; missing?: boolean; wide?: boolean; prose?: boolean }>
}) {
  return (
    <section className="space-y-3 border p-4">
      <h3 className="font-sans text-xs font-semibold uppercase tracking-[0.12em] text-muted-foreground">
        {title}
      </h3>

      <div className="grid gap-3 md:grid-cols-2">
        {items.map((item) => (
          <DetailItem
            key={`${title}-${item.label}`}
            label={item.label}
            value={item.value}
            missing={item.missing}
            wide={item.wide}
            prose={item.prose}
          />
        ))}
      </div>
    </section>
  )
}

function DetailItem({
  label,
  value,
  missing = false,
  wide = false,
  prose = false,
}: {
  label: string
  value: React.ReactNode
  missing?: boolean
  wide?: boolean
  prose?: boolean
}) {
  return (
    <div
      className={cn(
        "space-y-2 border p-3 transition-colors",
        wide && "md:col-span-2",
        missing
          ? "border-red-200/70 bg-linear-to-br from-red-500/8 via-rose-500/6 to-transparent dark:border-red-500/30 dark:from-red-500/12 dark:via-rose-500/10"
          : "bg-background"
      )}
    >
      <div className="font-sans text-[11px] uppercase tracking-[0.12em] text-muted-foreground">{label}</div>
      <div
        className={cn(
          "break-words text-sm text-foreground",
          prose && "font-sans leading-6",
          missing && "text-red-700 dark:text-red-300"
        )}
      >
        {value}
      </div>
    </div>
  )
}

function severityRank(severity: AlertSeverity) {
  switch (severity) {
    case "critical":
      return 4
    case "high":
      return 3
    case "medium":
      return 2
    case "low":
    default:
      return 1
  }
}

function severityIconClass(severity: AlertSeverity) {
  switch (severity) {
    case "critical":
      return "text-fuchsia-700 hover:text-fuchsia-800 focus-visible:text-fuchsia-800 dark:text-fuchsia-400 dark:hover:text-fuchsia-300 dark:focus-visible:text-fuchsia-300"
    case "high":
      return "text-red-600 hover:text-red-700 focus-visible:text-red-700 dark:text-red-400 dark:hover:text-red-300 dark:focus-visible:text-red-300"
    case "medium":
      return "text-amber-600 hover:text-amber-700 focus-visible:text-amber-700 dark:text-amber-400 dark:hover:text-amber-300 dark:focus-visible:text-amber-300"
    case "low":
    default:
      return "text-sky-600 hover:text-sky-700 focus-visible:text-sky-700 dark:text-sky-400 dark:hover:text-sky-300 dark:focus-visible:text-sky-300"
  }
}

function iconForAlert(alertId: string): React.ComponentType<{ className?: string }> {
  if (alertId.includes("missing_github") || alertId.includes("github_repository")) return GitBranch
  if (alertId.includes("no_license")) return Scale
  if (alertId.includes("maintainer") || alertId.includes("owner_track_record") || alertId.includes("owner_reputation")) return Users
  if (alertId.includes("very_new") || alertId.includes("new_owner_account") || alertId.includes("stale") || alertId.includes("age_mismatch")) {
    return Clock3
  }
  if (alertId.includes("commit_history") || alertId.includes("no_releases") || alertId.includes("low_maintenance")) {
    return GitCommitHorizontal
  }
  if (alertId.includes("low_stars")) return Star
  if (alertId.includes("low_forks")) return GitFork
  if (alertId.includes("low_watches")) return Activity
  if (alertId.includes("low_weekly_downloads") || alertId.includes("low_monthly_downloads") || alertId.includes("weak_consistency")) {
    return TrendingDown
  }
  if (alertId.includes("download_spike")) return TrendingUp
  if (alertId.includes("broken_trust_link_surface")) return Link2Off
  if (alertId.includes("missing_description")) return FileText
  if (alertId.includes("sparse") || alertId.includes("trust")) return ShieldAlert
  if (alertId.includes("missing")) return FileWarning
  return AlertTriangle
}

function InlineStatusIcon({
  label,
  icon: Icon,
  className,
}: {
  label: string
  icon: React.ComponentType<{ className?: string }>
  className?: string
}) {
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <button
          type="button"
          className={cn(
            "inline-flex items-center text-red-600 outline-none transition-colors hover:text-red-700 focus-visible:text-red-700 dark:text-red-400 dark:hover:text-red-300 dark:focus-visible:text-red-300",
            className
          )}
          aria-label={label}
          onClick={(event) => event.stopPropagation()}
        >
          <Icon className="size-3.5" />
        </button>
      </TooltipTrigger>
      <TooltipContent sideOffset={6} className="font-sans">
        {label}
      </TooltipContent>
    </Tooltip>
  )
}

function sortableHeader(
  column: { toggleSorting: (desc?: boolean) => void; getIsSorted: () => false | "asc" | "desc" },
  label: string
) {
  const state = column.getIsSorted()
  const Icon = state === "asc" ? ArrowUp : state === "desc" ? ArrowDown : ArrowUpDown

  return (
    <Button
      variant="ghost"
      size="sm"
      className={cn("-ml-2 h-8 font-sans", state && "text-foreground")}
      onClick={() => column.toggleSorting(state === "asc")}
    >
      {label}
      <Icon className="size-3.5" />
    </Button>
  )
}

function renderLink(value: string, label: string, Icon: React.ComponentType<{ className?: string }>) {
  if (!value) {
    return "—"
  }

  return (
    <a
      href={value}
      target="_blank"
      rel="noreferrer"
      className="inline-flex items-center gap-2 font-sans text-foreground underline underline-offset-4 hover:text-muted-foreground"
    >
      <Icon className="size-3.5" />
      <span>{label}</span>
    </a>
  )
}

function formatMaintainers(row: PackageTableRow) {
  if (!row.maintainers.length) {
    return display(row.maintainersText)
  }

  return row.maintainers
    .map((maintainer) => {
      if (maintainer.username && maintainer.email) {
        return `${maintainer.username} <${maintainer.email}>`
      }

      return maintainer.username || maintainer.email || "—"
    })
    .join(", ")
}

function display(value: string) {
  return value || "—"
}

function isMissingText(value: string) {
  return !value.trim()
}

function isMissingNumber(value: number | null) {
  return value == null
}

function normalizeExternalUrl(value: string) {
  if (!value) return ""
  return value.replace(/^git\+/, "").replace(/\.git$/, "")
}

function displayNumber(value: number | null) {
  return value == null ? "—" : new Intl.NumberFormat("en-US").format(value)
}

function displayDays(value: number | null) {
  return value == null ? "—" : `${displayNumber(value)} days`
}

function formatDate(value: string) {
  if (!value) return "—"
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return new Intl.DateTimeFormat("en-US", {
    year: "numeric",
    month: "short",
    day: "2-digit",
  }).format(date)
}

function formatCacheFileAge(value: string) {
  const filename = value.split("/").pop()?.replace(/\.json$/, "") ?? ""
  if (!filename) return "—"

  const cacheDate = new Date(`${filename}T00:00:00`)
  if (Number.isNaN(cacheDate.getTime())) return filename

  const now = new Date()
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate())
  const cacheDay = new Date(cacheDate.getFullYear(), cacheDate.getMonth(), cacheDate.getDate())
  const diffMs = today.getTime() - cacheDay.getTime()
  const diffDays = Math.max(0, Math.floor(diffMs / 86_400_000))

  return diffDays === 1 ? "1-day old" : `${diffDays}-days old`
}

