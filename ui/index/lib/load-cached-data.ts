import "server-only"

import { promises as fs } from "node:fs"
import path from "node:path"

import type {
  PackageTableRow,
  PackagesDashboardData,
  RawAlertsFile,
  RawGitHubCache,
  RawGitHubRepositoryMetadata,
  RawNpmLegacyCache,
  RawNpmLegacyObject,
  RawNpmPackageRecord,
} from "@/lib/types"

const DATA_ROOT = path.join(process.cwd(), "public", "data")
const MANIFEST_FILE = path.join(DATA_ROOT, "manifest.json")

export async function loadPackagesDashboardData(baseUrl?: string): Promise<PackagesDashboardData> {
  const manifest = await readJsonFile<DataManifest>("/data/manifest.json", MANIFEST_FILE, baseUrl)
  const npmCacheMeta = await resolveDataFile(manifest.npm_cache_file, "npm-data", baseUrl)
  const githubCacheMeta = manifest.github_cache_file
    ? await resolveDataFile(manifest.github_cache_file, "github", baseUrl)
    : null
  const alertsMeta = manifest.alerts_file ? await resolveDataFile(manifest.alerts_file, null, baseUrl) : null

  const npmCache = normalizeNpmCache(
    await readJsonFile<RawNpmLegacyCache | RawNpmPackageRecord[]>(
      npmCacheMeta.assetPath,
      npmCacheMeta.filePath,
      baseUrl,
    ),
  )

  const githubCache = githubCacheMeta
    ? ((await readJsonFile<RawGitHubCache>(githubCacheMeta.assetPath, githubCacheMeta.filePath, baseUrl)) ?? null)
    : null

  const alertsCache = alertsMeta
    ? ((await readJsonFile<RawAlertsFile>(alertsMeta.assetPath, alertsMeta.filePath, baseUrl)) ?? null)
    : null

  const rows = npmCache.map((entry, index) => toPackageTableRow(entry, index, githubCache?.repositories ?? {}, alertsCache))

  return {
    npmCacheFile: npmCacheMeta.assetPath,
    githubCacheFile: githubCacheMeta?.assetPath ?? null,
    alertsFile: alertsMeta?.assetPath ?? null,
    rowCount: rows.length,
    githubRepoCount: githubCache ? Object.keys(githubCache.repositories).length : 0,
    rows,
  }
}

type DataManifest = {
  generated_at: string
  npm_cache_file: string
  github_cache_file?: string
  alerts_file?: string
}

type ResolvedDataFile = {
  assetPath: string
  filePath: string
}

async function resolveDataFile(
  assetPath: string,
  fallbackDir: string | null,
  baseUrl?: string,
): Promise<ResolvedDataFile> {
  const directPath = toFilePath(assetPath)
  if (await fileExists(directPath)) {
    return { assetPath, filePath: directPath }
  }

  if (fallbackDir) {
    const latestAssetPath = await findLatestAssetPath(fallbackDir, baseUrl)
    return {
      assetPath: latestAssetPath,
      filePath: toFilePath(latestAssetPath),
    }
  }

  throw new Error(`Data file not found: ${assetPath}`)
}

function toFilePath(assetPath: string): string {
  return path.join(DATA_ROOT, assetPath.replace(/^\/data\/?/, ""))
}

async function fileExists(filePath: string): Promise<boolean> {
  try {
    await fs.access(filePath)
    return true
  } catch {
    return false
  }
}

async function readJsonFile<T>(assetPath: string, filePath: string, baseUrl?: string): Promise<T> {
  if (await fileExists(filePath)) {
    return JSON.parse(await fs.readFile(filePath, "utf8")) as T
  }

  if (!baseUrl) {
    throw new Error(`Could not load data asset ${assetPath} without a base URL`)
  }

  const response = await fetch(new URL(assetPath, baseUrl).toString(), {
    cache: "no-store",
  })

  if (!response.ok) {
    throw new Error(`Failed to fetch data asset ${assetPath}: ${response.status} ${response.statusText}`)
  }

  return (await response.json()) as T
}

async function findLatestAssetPath(dir: string, baseUrl?: string): Promise<string> {
  const absoluteDir = path.join(DATA_ROOT, dir)

  if (await fileExists(absoluteDir)) {
    const entries = await fs.readdir(absoluteDir, { withFileTypes: true })
    const filenames = entries
      .filter((entry) => entry.isFile() && entry.name.endsWith(".json"))
      .map((entry) => entry.name)
      .sort()

    const latest = filenames.at(-1)
    if (!latest) {
      throw new Error(`No JSON cache files found in ${absoluteDir}`)
    }

    return `/data/${dir}/${latest}`
  }

  const manifest = await readJsonFile<DataManifest>("/data/manifest.json", MANIFEST_FILE, baseUrl)
  const latest = dir === "npm-data" ? manifest.npm_cache_file : manifest.github_cache_file

  if (!latest) {
    throw new Error(`No JSON cache files found in ${dir}`)
  }

  return latest
}

function normalizeNpmCache(raw: RawNpmLegacyCache | RawNpmPackageRecord[]): RawNpmPackageRecord[] {
  if (Array.isArray(raw)) {
    return raw
  }

  return (raw.objects ?? []).map((entry) => legacyObjectToPackageRecord(entry))
}

function legacyObjectToPackageRecord(entry: RawNpmLegacyObject): RawNpmPackageRecord {
  const pkg = entry.package ?? {}
  const links = pkg.links ?? {}

  return {
    name: pkg.name,
    version: pkg.version,
    description: pkg.description,
    license: pkg.license,
    downloads_monthly: entry.downloads?.monthly,
    downloads_weekly: entry.downloads?.weekly,
    updated_at: entry.updated,
    created_at: pkg.date,
    maintainers: pkg.maintainers,
    links: {
      homepage: links.homepage,
      repo: links.repository,
      bugs: links.bugs,
      npm: links.npm,
    },
  }
}

function toPackageTableRow(
  entry: RawNpmPackageRecord,
  index: number,
  repositories: Record<string, RawGitHubRepositoryMetadata>,
  alertsCache: RawAlertsFile | null
): PackageTableRow {
  const links = entry.links ?? {}
  const repoUrl = links.repo ?? ""
  const githubFullName = normalizeGitHubRepository(repoUrl)
  const github = githubFullName ? repositories[githubFullName] : undefined
  const maintainers = (entry.maintainers ?? []).map((maintainer) => ({
    username: maintainer.username?.trim() ?? "",
    email: maintainer.email?.trim() ?? "",
  }))
  const maintainersText = maintainers
    .map((maintainer) => {
      if (maintainer.username && maintainer.email) return `${maintainer.username} <${maintainer.email}>`
      return maintainer.username || maintainer.email
    })
    .filter(Boolean)
    .join(", ")

  const row: PackageTableRow = {
    id: `${entry.name ?? "package"}-${index}`,
    packageName: entry.name ?? "",
    version: entry.version ?? "",
    description: entry.description ?? "",
    license: entry.license ?? "",
    downloadsMonthly: entry.downloads_monthly ?? null,
    downloadsWeekly: entry.downloads_weekly ?? null,
    npmUpdatedAt: entry.updated_at ?? "",
    npmCreatedAt: entry.created_at ?? "",
    maintainers,
    maintainersText,
    homepageUrl: links.homepage ?? "",
    repoUrl,
    bugsUrl: links.bugs ?? "",
    npmUrl: links.npm ?? "",
    githubFullName: github?.full_name ?? githubFullName,
    githubStars: github?.stars ?? null,
    githubForks: github?.forks ?? null,
    githubWatches: github?.watches ?? null,
    githubReleasesCount: github?.releases_count ?? null,
    githubCommitsCount: github?.commits_count ?? null,
    githubRepoAgeDays: github?.age_days ?? null,
    githubOwnerLogin: github?.owner_metadata.login ?? "",
    githubOwnerType: github?.owner_metadata.type ?? "",
    githubOwnerFollowers: github?.owner_metadata.followers ?? null,
    githubOwnerAgeDays: github?.owner_metadata.age_days ?? null,
    githubOwnerTotalRepos: github?.owner_metadata.total_repos ?? null,
    hasGithubData: Boolean(github),
    alerts: buildPackageAlerts(entry.name ?? "", alertsCache),
    securityScore: computeSecurityScore(buildPackageAlerts(entry.name ?? "", alertsCache)),
    searchText: "",
  }

  row.searchText = buildSearchText(row)
  return row
}

function buildSearchText(row: PackageTableRow): string {
  return [
    row.packageName,
    row.version,
    row.description,
    row.license,
    row.maintainersText,
    row.homepageUrl,
    row.repoUrl,
    row.bugsUrl,
    row.npmUrl,
    row.githubFullName,
    row.githubOwnerLogin,
    row.githubOwnerType,
    stringifyNullable(row.downloadsMonthly),
    stringifyNullable(row.downloadsWeekly),
    stringifyNullable(row.githubStars),
    stringifyNullable(row.githubForks),
    stringifyNullable(row.githubWatches),
    stringifyNullable(row.githubReleasesCount),
    stringifyNullable(row.githubCommitsCount),
    stringifyNullable(row.githubRepoAgeDays),
    stringifyNullable(row.githubOwnerFollowers),
    stringifyNullable(row.githubOwnerAgeDays),
    stringifyNullable(row.githubOwnerTotalRepos),
    row.npmUpdatedAt,
    row.npmCreatedAt,
    ...row.alerts.flatMap((alert) => [alert.name, alert.description, alert.severity]),
  ]
    .filter(Boolean)
    .join(" ")
    .toLowerCase()
}

function stringifyNullable(value: number | null): string {
  return value == null ? "" : String(value)
}

function buildPackageAlerts(packageName: string, alertsCache: RawAlertsFile | null): PackageTableRow["alerts"] {
  if (!packageName || !alertsCache) return []

  const alertIds = alertsCache.packages[packageName]?.alert_ids ?? []
  return alertIds
    .map((alertId) => {
      const definition = alertsCache.definitions[alertId]
      if (!definition) return null
      return {
        id: definition.id,
        name: definition.name,
        description: definition.description,
        severity: definition.severity,
      }
    })
    .filter((alert): alert is NonNullable<typeof alert> => Boolean(alert))
}

function severityScoreWeight(severity: PackageTableRow["alerts"][number]["severity"]): number {
  switch (severity) {
    case "critical":
      return 8
    case "high":
      return 4
    case "medium":
      return 2
    case "low":
    default:
      return 1
  }
}

function computeSecurityScore(alerts: PackageTableRow["alerts"]): number {
  return alerts.reduce((score, alert) => score + severityScoreWeight(alert.severity), 0)
}

function normalizeGitHubRepository(value: string): string {
  const trimmed = value.trim()
  if (!trimmed) return ""

  let normalized = trimmed.replace(/^git\+/, "").replace(/\.git$/, "").replace(/\/$/, "")

  if (normalized.startsWith("git@github.com:")) {
    normalized = normalized.replace("git@github.com:", "")
    return cleanRepositoryPath(normalized)
  }

  if (normalized.startsWith("github:")) {
    normalized = normalized.replace("github:", "")
    return cleanRepositoryPath(normalized)
  }

  if (!normalized.includes("://")) {
    return cleanRepositoryPath(normalized)
  }

  try {
    const parsed = new URL(normalized)
    if (parsed.hostname.toLowerCase() !== "github.com") {
      return ""
    }
    return cleanRepositoryPath(parsed.pathname)
  } catch {
    return ""
  }
}

function cleanRepositoryPath(value: string): string {
  const parts = value.replace(/^\//, "").split("/").filter(Boolean)
  if (parts.length < 2) return ""
  return `${parts[0]}/${parts[1]}`
}
