package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"sync/internal/githubdata"
	"sync/internal/npm"
)

func main() {
	var githubRepo string
	var githubAllFromNPM bool
	flag.StringVar(&githubRepo, "github-repo", "", "repository to fetch in owner/name format")
	flag.BoolVar(&githubAllFromNPM, "github-all-from-npm", false, "fetch GitHub metadata for all repositories referenced by fetched npm packages")
	flag.Parse()

	ctx := context.Background()
	now := time.Now()

	switch {
	case githubRepo != "":
		metadata, cacheFile, err := githubdata.FetchRepositoryMetadata(ctx, githubRepo, now)
		if err != nil {
			log.Fatalf("sync github repository: %v", err)
		}
		writeJSON(metadata)
		fmt.Fprintf(os.Stderr, "fetched github metadata for %s using cache file %s\n", githubRepo, cacheFile)
		return

	case githubAllFromNPM:
		packages, npmCacheFile, err := npm.FetchAndExtractPackages(ctx, nil, now)
		if err != nil {
			log.Fatalf("load npm packages: %v", err)
		}
		result, err := githubdata.FetchRepositoriesFromPackages(ctx, packages, now)
		if err != nil {
			log.Fatalf("sync github repositories from npm packages: %v", err)
		}
		writeJSON(result)
		fmt.Fprintf(os.Stderr, "processed %d npm packages from %s and cached %d GitHub repositories in %s\n", result.Requested, npmCacheFile, len(result.Fetched), result.CacheFile)
		return
	}

	records, cacheFile, err := npm.FetchAndExtractPackages(ctx, nil, now)
	if err != nil {
		log.Fatalf("sync npm packages: %v", err)
	}
	writeJSON(records)
	fmt.Fprintf(os.Stderr, "extracted %d packages using cache file %s\n", len(records), cacheFile)
}

func writeJSON(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		log.Fatalf("encode output: %v", err)
	}
}
