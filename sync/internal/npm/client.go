package npm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"sync/internal/paths"
)

const (
	searchBaseURL  = "https://registry.npmjs.org/-/v1/search"
	searchPageSize = 250
	searchLimit    = 500
)

func cachePath(now time.Time) string {
	return filepath.Join(paths.NPMDataDir(), now.Format("2006-01-02")+".json")
}

func LoadOrFetchPackages(ctx context.Context, client *http.Client, now time.Time) ([]PackageRecord, string, error) {
	path := cachePath(now)

	packages, found, err := loadCachedPackages(path)
	if err != nil {
		return nil, path, err
	}
	if found {
		return packages, path, nil
	}

	packages, err = fetchPackages(ctx, client)
	if err != nil {
		return nil, path, err
	}

	if err := writePackagesCache(path, packages); err != nil {
		return nil, path, err
	}

	return packages, path, nil
}

func loadCachedPackages(path string) ([]PackageRecord, bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("read cache %s: %w", path, err)
	}

	var packages []PackageRecord
	if err := json.Unmarshal(data, &packages); err == nil {
		return packages, true, nil
	}

	fmt.Fprintf(os.Stderr, "ignoring legacy npm cache %s and refreshing\n", path)
	return nil, false, nil
}

func writePackagesCache(path string, packages []PackageRecord) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create cache dir: %w", err)
	}

	data, err := json.MarshalIndent(packages, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal npm package cache: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write cache %s: %w", path, err)
	}

	return nil
}

func fetchPackages(ctx context.Context, client *http.Client) ([]PackageRecord, error) {
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}

	searchResponse, err := fetchSearchResponse(ctx, client)
	if err != nil {
		return nil, err
	}

	return ExtractPackages(ctx, searchResponse)
}

func fetchSearchResponse(ctx context.Context, client *http.Client) (SearchResponse, error) {
	combined := SearchResponse{
		Objects: make([]SearchObject, 0, searchLimit),
	}

	for offset := 0; offset < searchLimit; offset += searchPageSize {
		pageNumber := offset/searchPageSize + 1
		fmt.Fprintf(os.Stderr, "npm search: fetching page %d (offset %d)\n", pageNumber, offset)

		page, err := fetchSearchPage(ctx, client, offset)
		if err != nil {
			return SearchResponse{}, err
		}

		combined.Total = page.Total
		combined.Time = page.Time
		combined.Objects = append(combined.Objects, page.Objects...)

		if len(page.Objects) < searchPageSize || len(combined.Objects) >= searchLimit || len(combined.Objects) >= page.Total {
			break
		}
	}

	if len(combined.Objects) > searchLimit {
		combined.Objects = combined.Objects[:searchLimit]
	}

	return combined, nil
}

func fetchSearchPage(ctx context.Context, client *http.Client, offset int) (SearchResponse, error) {
	requestURL, err := buildSearchURL(offset)
	if err != nil {
		return SearchResponse{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return SearchResponse{}, fmt.Errorf("create search request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return SearchResponse{}, fmt.Errorf("fetch npm search page at offset %d: %w", offset, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return SearchResponse{}, fmt.Errorf("unexpected npm search status %d for offset %d: %s", resp.StatusCode, offset, string(body))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return SearchResponse{}, fmt.Errorf("read npm search response body for offset %d: %w", offset, err)
	}

	var page SearchResponse
	if err := json.Unmarshal(data, &page); err != nil {
		return SearchResponse{}, fmt.Errorf("unmarshal npm search response for offset %d: %w", offset, err)
	}

	return page, nil
}

func buildSearchURL(offset int) (string, error) {
	parsed, err := url.Parse(searchBaseURL)
	if err != nil {
		return "", fmt.Errorf("parse npm search url: %w", err)
	}

	query := parsed.Query()
	query.Set("text", "keywords:pi-package")
	query.Set("size", strconv.Itoa(searchPageSize))
	query.Set("from", strconv.Itoa(offset))
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}
