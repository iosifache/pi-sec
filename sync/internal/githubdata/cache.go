package githubdata

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

func LoadLatestDailyCache(dir string) (DailyCache, string, error) {
	matches, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return DailyCache{}, "", fmt.Errorf("glob github cache files in %s: %w", dir, err)
	}
	if len(matches) == 0 {
		return DailyCache{Repositories: map[string]RepositoryMetadata{}}, "", nil
	}

	sort.Strings(matches)
	path := matches[len(matches)-1]
	data, err := os.ReadFile(path)
	if err != nil {
		return DailyCache{}, path, fmt.Errorf("read github cache %s: %w", path, err)
	}

	var cache DailyCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return DailyCache{}, path, fmt.Errorf("unmarshal github cache %s: %w", path, err)
	}
	if cache.Repositories == nil {
		cache.Repositories = map[string]RepositoryMetadata{}
	}
	return cache, path, nil
}
