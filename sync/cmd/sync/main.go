package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"sync/internal/alerts"
	"sync/internal/githubdata"
	"sync/internal/npm"
	"sync/internal/paths"
)

func main() {
	var githubRepo string
	flag.StringVar(&githubRepo, "github-repo", "", "repository to fetch in owner/name format")
	flag.Parse()

	ctx := context.Background()
	now := time.Now()

	if githubRepo != "" {
		metadata, cacheFile, err := githubdata.FetchRepositoryMetadata(ctx, githubRepo, now)
		if err != nil {
			log.Fatalf("sync github repository: %v", err)
		}
		writeJSON(metadata)
		fmt.Fprintf(os.Stderr, "fetched github metadata for %s using cache file %s\n", githubRepo, cacheFile)
		return
	}

	result, err := runDailySync(ctx, now)
	if err != nil {
		log.Fatalf("run daily sync: %v", err)
	}
	writeJSON(result)
	fmt.Fprintf(os.Stderr, "backed up %d files, fetched %d npm packages into %s, cached %d GitHub repositories in %s, recreated %s\n", len(result.Backups), result.NPMPackageCount, result.NPMCacheFile, len(result.GitHub.Fetched), result.GitHub.CacheFile, result.AlertsPath)
}

type DailySyncResult struct {
	Date                 string                     `json:"date"`
	Backups              []string                   `json:"backups"`
	NPMCacheFile         string                     `json:"npm_cache_file"`
	NPMPackageCount      int                        `json:"npm_package_count"`
	GitHub               githubdata.BulkFetchResult `json:"github"`
	AlertsPath           string                     `json:"alerts_path"`
	AlertedPackageCount  int                        `json:"alerted_package_count"`
	AlertDefinitionCount int                        `json:"alert_definition_count"`
}

func runDailySync(ctx context.Context, now time.Time) (DailySyncResult, error) {
	npmPath := filepath.Join(paths.NPMDataDir(), now.Format("2006-01-02")+".json")
	githubPath := filepath.Join(paths.GitHubDataDir(), now.Format("2006-01-02")+".json")

	backups := make([]string, 0, 2)
	for _, path := range []string{npmPath, githubPath} {
		backupPath, err := backupAndRemove(path, now)
		if err != nil {
			return DailySyncResult{}, err
		}
		if backupPath != "" {
			backups = append(backups, backupPath)
		}
	}

	packages, npmCacheFile, err := npm.FetchAndExtractPackages(ctx, nil, now)
	if err != nil {
		return DailySyncResult{}, fmt.Errorf("refresh npm packages: %w", err)
	}
	if err := syncLatestCopy(npmCacheFile); err != nil {
		return DailySyncResult{}, fmt.Errorf("sync latest npm cache: %w", err)
	}

	githubResult, err := githubdata.FetchRepositoriesFromPackages(ctx, packages, now)
	if err != nil {
		return DailySyncResult{}, fmt.Errorf("refresh github repositories from npm packages: %w", err)
	}
	if err := syncLatestCopy(githubResult.CacheFile); err != nil {
		return DailySyncResult{}, fmt.Errorf("sync latest github cache: %w", err)
	}

	alertsFile, _, _, githubCacheFile, alertsPath, err := buildAlertsFile(ctx, now)
	if err != nil {
		return DailySyncResult{}, err
	}
	githubResult.CacheFile = githubCacheFile

	return DailySyncResult{
		Date:                 now.Format("2006-01-02"),
		Backups:              backups,
		NPMCacheFile:         npmCacheFile,
		NPMPackageCount:      len(packages),
		GitHub:               githubResult,
		AlertsPath:           alertsPath,
		AlertedPackageCount:  len(alertsFile.Packages),
		AlertDefinitionCount: len(alertsFile.Definitions),
	}, nil
}

func buildAlertsFile(ctx context.Context, now time.Time) (alerts.AlertsFile, []npm.PackageRecord, string, string, string, error) {
	packages, npmCacheFile, err := npm.FetchAndExtractPackages(ctx, nil, now)
	if err != nil {
		return alerts.AlertsFile{}, nil, "", "", "", fmt.Errorf("load npm packages for alerts: %w", err)
	}
	githubCache, githubCacheFile, err := githubdata.LoadLatestDailyCache(paths.GitHubDataDir())
	if err != nil {
		return alerts.AlertsFile{}, nil, "", "", "", fmt.Errorf("load cached github metadata for alerts: %w", err)
	}
	alertsFile := alerts.Build(packages, githubCache.Repositories, now)
	alertsPath := paths.AlertsFile(now.Format("2006-01-02"))
	backupPath, err := backupAndRemove(alertsPath, now)
	if err != nil {
		return alerts.AlertsFile{}, nil, "", "", "", fmt.Errorf("backup existing alerts file: %w", err)
	}
	if err := alerts.WriteFile(alertsPath, alertsFile); err != nil {
		return alerts.AlertsFile{}, nil, "", "", "", fmt.Errorf("write alerts file: %w", err)
	}
	if err := syncLatestCopy(alertsPath); err != nil {
		return alerts.AlertsFile{}, nil, "", "", "", fmt.Errorf("sync latest alerts file: %w", err)
	}
	_ = backupPath
	return alertsFile, packages, npmCacheFile, githubCacheFile, alertsPath, nil
}

func backupAndRemove(path string, now time.Time) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("read file for backup %s: %w", path, err)
	}

	backupDir := filepath.Join(filepath.Dir(path), "backups")
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	ext := filepath.Ext(path)
	backupPath := filepath.Join(backupDir, fmt.Sprintf("%s.%s%s", base, now.Format("150405"), ext))
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		return "", fmt.Errorf("create backup dir %s: %w", backupDir, err)
	}
	if err := os.WriteFile(backupPath, data, 0o644); err != nil {
		return "", fmt.Errorf("write backup %s: %w", backupPath, err)
	}
	if err := os.Remove(path); err != nil {
		return "", fmt.Errorf("remove original file %s after backup: %w", path, err)
	}
	return backupPath, nil
}

func syncLatestCopy(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read file for latest copy %s: %w", path, err)
	}

	latestPath := filepath.Join(filepath.Dir(path), "latest.json")
	if err := os.WriteFile(latestPath, data, 0o644); err != nil {
		return fmt.Errorf("write latest copy %s: %w", latestPath, err)
	}

	return nil
}

func writeJSON(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		log.Fatalf("encode output: %v", err)
	}
}
