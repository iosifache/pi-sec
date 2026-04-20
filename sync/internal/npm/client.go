package npm

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const SearchURL = "https://registry.npmjs.org/-/v1/search?text=keywords:pi-package&size=250&from=1000"

func cachePath(now time.Time) string {
	return filepath.Join("data", "npm-data", now.Format("2006-01-02")+".json")
}

func LoadOrFetchRawJSON(ctx context.Context, client *http.Client, now time.Time) ([]byte, string, error) {
	path := cachePath(now)
	if data, err := os.ReadFile(path); err == nil {
		return data, path, nil
	} else if !os.IsNotExist(err) {
		return nil, path, fmt.Errorf("read cache %s: %w", path, err)
	}

	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, SearchURL, nil)
	if err != nil {
		return nil, path, fmt.Errorf("create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, path, fmt.Errorf("fetch npm data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, path, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, path, fmt.Errorf("read response body: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, path, fmt.Errorf("create cache dir: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return nil, path, fmt.Errorf("write cache %s: %w", path, err)
	}

	return data, path, nil
}
