export type RawNpmLegacyCache = {
  objects: RawNpmLegacyObject[]
  total?: number
  time?: string
}

export type RawNpmLegacyObject = {
  downloads?: {
    monthly?: number
    weekly?: number
  }
  updated?: string
  package?: {
    name?: string
    version?: string
    description?: string
    maintainers?: Array<{
      username?: string
      email?: string
    }>
    license?: string
    date?: string
    links?: {
      homepage?: string
      repository?: string
      bugs?: string
      npm?: string
    }
  }
}

export type RawNpmPackageRecord = {
  name?: string
  version?: string
  description?: string
  license?: string
  downloads_monthly?: number
  downloads_weekly?: number
  updated_at?: string
  created_at?: string
  maintainers?: Array<{
    username?: string
    email?: string
  }>
  links?: {
    homepage?: string
    repo?: string
    bugs?: string
    npm?: string
  }
}

export type RawGitHubRepositoryMetadata = {
  full_name: string
  owner: string
  name: string
  source_url?: string
  stars: number
  forks: number
  watches: number
  releases_count: number
  commits_count: number
  age_days: number
  created_at: string
  owner_metadata: {
    login: string
    type: string
    followers: number
    age_days: number
    created_at: string
    public_repos: number
    total_repos: number
  }
  fetched_at: string
}

export type RawGitHubCache = {
  date: string
  repositories: Record<string, RawGitHubRepositoryMetadata>
}

export type PackageTableMaintainer = {
  username: string
  email: string
}

export type AlertSeverity = "low" | "medium" | "high" | "critical"

export type RawAlertDefinition = {
  id: string
  name: string
  description: string
  severity: AlertSeverity
}

export type RawPackageAlerts = {
  alert_ids: string[]
}

export type RawAlertsFile = {
  date: string
  updated_at?: string
  definitions: Record<string, RawAlertDefinition>
  packages: Record<string, RawPackageAlerts>
}

export type PackageAlert = {
  id: string
  name: string
  description: string
  severity: AlertSeverity
}

export type PackageTableRow = {
  id: string
  packageName: string
  version: string
  description: string
  license: string
  downloadsMonthly: number | null
  downloadsWeekly: number | null
  npmUpdatedAt: string
  npmCreatedAt: string
  maintainers: PackageTableMaintainer[]
  maintainersText: string
  homepageUrl: string
  repoUrl: string
  bugsUrl: string
  npmUrl: string
  githubFullName: string
  githubStars: number | null
  githubForks: number | null
  githubWatches: number | null
  githubReleasesCount: number | null
  githubCommitsCount: number | null
  githubRepoAgeDays: number | null
  githubOwnerLogin: string
  githubOwnerType: string
  githubOwnerFollowers: number | null
  githubOwnerAgeDays: number | null
  githubOwnerTotalRepos: number | null
  hasGithubData: boolean
  alerts: PackageAlert[]
  securityScore: number
  searchText: string
}

export type PackagesDashboardData = {
  npmCacheFile: string
  githubCacheFile: string | null
  alertsFile: string | null
  lastUpdatedAt: string | null
  rowCount: number
  githubRepoCount: number
  rows: PackageTableRow[]
}
