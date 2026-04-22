package githubdata

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	gh "github.com/google/go-github/v61/github"
	"golang.org/x/oauth2"

	"sync/internal/config"
	"sync/internal/npm"
	"sync/internal/paths"
)

const githubTokenEnv = "GITHUB_PAT"

func cachePath(now time.Time) string {
	return filepath.Join(paths.GitHubDataDir(), now.Format("2006-01-02")+".json")
}

func FetchRepositoryMetadata(ctx context.Context, fullName string, now time.Time) (RepositoryMetadata, string, error) {
	owner, repo, err := splitRepository(fullName)
	if err != nil {
		return RepositoryMetadata{}, "", err
	}

	cacheFile := cachePath(now)
	cache, err := loadDailyCache(cacheFile, now)
	if err != nil {
		return RepositoryMetadata{}, cacheFile, err
	}
	if cached, ok := cache.Repositories[fullName]; ok {
		return cached, cacheFile, nil
	}

	client, err := newGitHubClient(ctx)
	if err != nil {
		return RepositoryMetadata{}, cacheFile, err
	}

	metadata, err := fetchRepositoryMetadata(ctx, client, owner, repo, now)
	if err != nil {
		return RepositoryMetadata{}, cacheFile, err
	}

	cache.Repositories[metadata.FullName] = metadata
	if err := writeDailyCache(cacheFile, cache); err != nil {
		return RepositoryMetadata{}, cacheFile, err
	}

	return metadata, cacheFile, nil
}

func FetchRepositoriesFromPackages(ctx context.Context, packages []npm.PackageRecord, now time.Time) (BulkFetchResult, error) {
	cacheFile := cachePath(now)
	cache, err := loadDailyCache(cacheFile, now)
	if err != nil {
		return BulkFetchResult{}, err
	}

	resolved := map[string]string{}
	skipped := make([]string, 0)
	for _, pkg := range packages {
		repoURL := strings.TrimSpace(pkg.Links.Repo)
		if repoURL == "" {
			skipped = append(skipped, pkg.Name+": missing repository link")
			continue
		}
		fullName, err := NormalizeGitHubRepository(repoURL)
		if err != nil {
			skipped = append(skipped, fmt.Sprintf("%s: %v (%s)", pkg.Name, err, repoURL))
			continue
		}
		if _, exists := resolved[fullName]; !exists {
			resolved[fullName] = repoURL
		}
	}

	client, err := newGitHubClient(ctx)
	if err != nil {
		return BulkFetchResult{}, err
	}

	fullNames := make([]string, 0, len(resolved))
	for fullName := range resolved {
		fullNames = append(fullNames, fullName)
	}
	sort.Strings(fullNames)

	fetched := make([]RepositoryMetadata, 0, len(fullNames))
	total := len(fullNames)
	fmt.Fprintf(os.Stderr, "github: fetching metadata for %d repositories\n", total)
	for i, fullName := range fullNames {
		fmt.Fprintf(os.Stderr, "github: repo %d/%d %s\n", i+1, total, fullName)
		if cached, ok := cache.Repositories[fullName]; ok {
			fetched = append(fetched, cached)
			continue
		}

		owner, repo, err := splitRepository(fullName)
		if err != nil {
			skipped = append(skipped, fmt.Sprintf("%s: %v", fullName, err))
			continue
		}

		metadata, err := fetchRepositoryMetadata(ctx, client, owner, repo, now)
		if err != nil {
			skipped = append(skipped, fmt.Sprintf("%s: %v", fullName, err))
			continue
		}
		metadata.SourceURL = resolved[fullName]
		cache.Repositories[metadata.FullName] = metadata
		fetched = append(fetched, metadata)

		if err := writeDailyCache(cacheFile, cache); err != nil {
			return BulkFetchResult{}, err
		}
	}

	return BulkFetchResult{
		CacheFile:      cacheFile,
		Requested:      len(packages),
		UniqueResolved: len(fullNames),
		Fetched:        fetched,
		Skipped:        skipped,
	}, nil
}

func NormalizeGitHubRepository(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", fmt.Errorf("empty repository link")
	}

	value = strings.TrimPrefix(value, "git+")
	value = strings.TrimSuffix(value, ".git")
	value = strings.TrimSuffix(value, "/")

	if strings.HasPrefix(value, "git@github.com:") {
		value = strings.TrimPrefix(value, "git@github.com:")
		return cleanFullName(value)
	}
	if strings.HasPrefix(value, "github:") {
		value = strings.TrimPrefix(value, "github:")
		return cleanFullName(value)
	}
	if strings.Count(value, "/") == 1 && !strings.Contains(value, "://") {
		return cleanFullName(value)
	}

	parsed, err := url.Parse(value)
	if err != nil {
		return "", fmt.Errorf("parse repository URL: %w", err)
	}
	if !strings.EqualFold(parsed.Host, "github.com") {
		return "", fmt.Errorf("repository is not hosted on github.com")
	}

	return cleanFullName(strings.TrimPrefix(parsed.Path, "/"))
}

func cleanFullName(value string) (string, error) {
	value = strings.Trim(value, "/")
	parts := strings.Split(value, "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("repository must be in owner/name format")
	}
	return parts[0] + "/" + parts[1], nil
}

func splitRepository(fullName string) (string, string, error) {
	parts := strings.Split(strings.TrimSpace(fullName), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("repository must be in owner/name format")
	}
	return parts[0], parts[1], nil
}

func newGitHubClient(ctx context.Context) (*gh.Client, error) {
	token, err := config.LookupEnvOrDotEnv(githubTokenEnv)
	if err != nil {
		return nil, err
	}
	return newClient(ctx, token), nil
}

func newClient(ctx context.Context, token string) *gh.Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	httpClient := oauth2.NewClient(ctx, ts)
	httpClient.Timeout = 30 * time.Second
	return gh.NewClient(httpClient)
}

func fetchRepositoryMetadata(ctx context.Context, client *gh.Client, owner, repo string, now time.Time) (RepositoryMetadata, error) {
	repository, _, err := client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return RepositoryMetadata{}, fmt.Errorf("get repository %s/%s: %w", owner, repo, err)
	}

	releasesCount, err := fetchReleasesCount(ctx, client, owner, repo)
	if err != nil {
		return RepositoryMetadata{}, err
	}

	commitsCount, err := fetchCommitsCount(ctx, client, owner, repo, repository.GetDefaultBranch())
	if err != nil {
		return RepositoryMetadata{}, err
	}

	ownerMetadata, err := fetchOwnerMetadata(ctx, client, repository.GetOwner().GetLogin(), repository.GetOwner().GetType(), now)
	if err != nil {
		return RepositoryMetadata{}, err
	}

	return RepositoryMetadata{
		FullName:      repository.GetFullName(),
		Owner:         repository.GetOwner().GetLogin(),
		Name:          repository.GetName(),
		Stars:         repository.GetStargazersCount(),
		Forks:         repository.GetForksCount(),
		Watches:       repository.GetSubscribersCount(),
		ReleasesCount: releasesCount,
		CommitsCount:  commitsCount,
		AgeDays:       ageDays(repository.GetCreatedAt().Time, now),
		CreatedAt:     repository.GetCreatedAt().Time,
		OwnerMetadata: ownerMetadata,
		FetchedAt:     now.UTC(),
	}, nil
}

func fetchReleasesCount(ctx context.Context, client *gh.Client, owner, repo string) (int, error) {
	_, resp, err := client.Repositories.ListReleases(ctx, owner, repo, &gh.ListOptions{Page: 1, PerPage: 1})
	if err != nil {
		return 0, fmt.Errorf("list releases for %s/%s: %w", owner, repo, err)
	}
	if resp.LastPage > 0 {
		return resp.LastPage, nil
	}
	return 0, nil
}

func fetchCommitsCount(ctx context.Context, client *gh.Client, owner, repo, branch string) (int, error) {
	_, resp, err := client.Repositories.ListCommits(ctx, owner, repo, &gh.CommitsListOptions{
		SHA:         branch,
		ListOptions: gh.ListOptions{Page: 1, PerPage: 1},
	})
	if err != nil {
		return 0, fmt.Errorf("list commits for %s/%s: %w", owner, repo, err)
	}
	if resp.LastPage > 0 {
		return resp.LastPage, nil
	}
	if len(resp.NextPageToken) > 0 {
		return 1, nil
	}
	return 0, nil
}

func fetchOwnerMetadata(ctx context.Context, client *gh.Client, login, ownerType string, now time.Time) (OwnerMetadata, error) {
	switch strings.ToLower(ownerType) {
	case "organization":
		org, _, err := client.Organizations.Get(ctx, login)
		if err != nil {
			return OwnerMetadata{}, fmt.Errorf("get organization %s: %w", login, err)
		}
		publicRepos := org.GetPublicRepos()
		return OwnerMetadata{
			Login:       org.GetLogin(),
			Type:        "organization",
			Followers:   org.GetFollowers(),
			AgeDays:     ageDays(org.GetCreatedAt().Time, now),
			CreatedAt:   org.GetCreatedAt().Time,
			PublicRepos: publicRepos,
			TotalRepos:  publicRepos,
		}, nil
	default:
		user, _, err := client.Users.Get(ctx, login)
		if err != nil {
			return OwnerMetadata{}, fmt.Errorf("get user %s: %w", login, err)
		}
		publicRepos := user.GetPublicRepos()
		return OwnerMetadata{
			Login:       user.GetLogin(),
			Type:        "user",
			Followers:   user.GetFollowers(),
			AgeDays:     ageDays(user.GetCreatedAt().Time, now),
			CreatedAt:   user.GetCreatedAt().Time,
			PublicRepos: publicRepos,
			TotalRepos:  publicRepos,
		}, nil
	}
}

func ageDays(createdAt, now time.Time) int {
	if createdAt.IsZero() {
		return 0
	}
	if now.IsZero() {
		now = time.Now()
	}
	return int(now.UTC().Sub(createdAt.UTC()).Hours() / 24)
}

func loadDailyCache(path string, now time.Time) (DailyCache, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DailyCache{
				Date:         now.Format("2006-01-02"),
				Repositories: map[string]RepositoryMetadata{},
			}, nil
		}
		return DailyCache{}, fmt.Errorf("read github cache %s: %w", path, err)
	}

	var cache DailyCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return DailyCache{}, fmt.Errorf("unmarshal github cache %s: %w", path, err)
	}
	if cache.Repositories == nil {
		cache.Repositories = map[string]RepositoryMetadata{}
	}
	if cache.Date == "" {
		cache.Date = now.Format("2006-01-02")
	}
	return cache, nil
}

func writeDailyCache(path string, cache DailyCache) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create github cache dir: %w", err)
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal github cache: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write github cache %s: %w", path, err)
	}
	return nil
}
