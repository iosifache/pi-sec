package githubdata

import "time"

type OwnerMetadata struct {
	Login       string    `json:"login"`
	Type        string    `json:"type"`
	Followers   int       `json:"followers"`
	AgeDays     int       `json:"age_days"`
	CreatedAt   time.Time `json:"created_at"`
	PublicRepos int       `json:"public_repos"`
	TotalRepos  int       `json:"total_repos"`
}

type RepositoryMetadata struct {
	FullName      string        `json:"full_name"`
	Owner         string        `json:"owner"`
	Name          string        `json:"name"`
	SourceURL     string        `json:"source_url,omitempty"`
	Stars         int           `json:"stars"`
	Forks         int           `json:"forks"`
	Watches       int           `json:"watches"`
	ReleasesCount int           `json:"releases_count"`
	CommitsCount  int           `json:"commits_count"`
	AgeDays       int           `json:"age_days"`
	CreatedAt     time.Time     `json:"created_at"`
	OwnerMetadata OwnerMetadata `json:"owner_metadata"`
	FetchedAt     time.Time     `json:"fetched_at"`
}

type DailyCache struct {
	Date         string                        `json:"date"`
	Repositories map[string]RepositoryMetadata `json:"repositories"`
}

type BulkFetchResult struct {
	CacheFile      string               `json:"cache_file"`
	Requested      int                  `json:"requested"`
	UniqueResolved int                  `json:"unique_resolved"`
	Fetched        []RepositoryMetadata `json:"fetched"`
	Skipped        []string             `json:"skipped"`
}
