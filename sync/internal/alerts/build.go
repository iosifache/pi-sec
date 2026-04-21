package alerts

import (
	"sort"
	"time"

	"sync/internal/githubdata"
	"sync/internal/npm"
)

func Build(packages []npm.PackageRecord, repositories map[string]githubdata.RepositoryMetadata, now time.Time) AlertsFile {
	definitions := map[string]AlertDefinition{}
	rules := Rules()
	for _, rule := range rules {
		definitions[rule.ID()] = rule.ToJSON()
	}

	packagesMap := map[string]PackageAlerts{}
	packageNames := make([]string, 0, len(packages))
	for _, pkg := range packages {
		packageNames = append(packageNames, pkg.Name)
	}
	sort.Strings(packageNames)

	byName := make(map[string]npm.PackageRecord, len(packages))
	for _, pkg := range packages {
		byName[pkg.Name] = pkg
	}

	for _, packageName := range packageNames {
		pkg := byName[packageName]
		repo := resolveRepository(pkg, repositories)
		alertIDs := make([]string, 0)
		for _, rule := range rules {
			if rule.Check(pkg, repo, now) {
				alertIDs = append(alertIDs, rule.ID())
			}
		}
		packagesMap[packageName] = PackageAlerts{AlertIDs: alertIDs}
	}

	return AlertsFile{
		Date:        now.Format("2006-01-02"),
		Definitions: definitions,
		Packages:    packagesMap,
	}
}

func resolveRepository(pkg npm.PackageRecord, repositories map[string]githubdata.RepositoryMetadata) *githubdata.RepositoryMetadata {
	repoURL := pkg.Links.Repo
	if repoURL == "" {
		return nil
	}
	fullName, err := githubdata.NormalizeGitHubRepository(repoURL)
	if err != nil {
		return nil
	}
	repo, ok := repositories[fullName]
	if !ok {
		return nil
	}
	copy := repo
	return &copy
}
