package alerts

import (
	"strings"
	"time"

	"sync/internal/githubdata"
	"sync/internal/npm"
)

func Rules() []AlertRule {
	return []AlertRule{
		MissingGitHubRepositoryMetadataRule{},
		VeryNewPackageHighRule{},
		VeryNewPackageMediumRule{},
		VeryNewRepositoryHighRule{},
		VeryNewRepositoryMediumRule{},
		NoMaintainerRule{},
		SingleMaintainerRule{},
		LowOwnerReputationHighRule{},
		LowOwnerReputationMediumRule{},
		LowOwnerReputationLowRule{},
		NewOwnerAccountHighRule{},
		NewOwnerAccountMediumRule{},
		TinyCommitHistoryHighRule{},
		TinyCommitHistoryMediumRule{},
		TinyCommitHistoryLowRule{},
		NoReleasesRule{},
		NoLicenseRule{},
		LowWeeklyDownloadsHighRule{},
		LowWeeklyDownloadsMediumRule{},
		LowWeeklyDownloadsLowRule{},
		LowMonthlyDownloadsMediumRule{},
		LowMonthlyDownloadsLowRule{},
		StalePackageUpdatesHighRule{},
		StalePackageUpdatesMediumRule{},
		StalePackageUpdatesLowRule{},
		LowStarsMediumRule{},
		LowStarsLowRule{},
		LowForksLowRule{},
		LowForksMediumOldPackageRule{},
		LowWatchesRule{},
		LowOwnerTrackRecordMediumRule{},
		LowOwnerTrackRecordLowRule{},
		PackageRepoAgeMismatchMediumRule{},
		PackageRepoAgeMismatchLowRule{},
		DownloadSpikeHighRule{},
		DownloadSpikeCriticalRule{},
		MaintainerMissingAllIdentityRule{},
		MaintainerMissingPartialIdentityRule{},
		MissingDescriptionRule{},
		BrokenTrustLinkSurfaceMediumRule{},
		BrokenTrustLinkSurfaceCriticalRule{},
		WeakConsistencyAcrossSignalsRule{},
		LowMaintenanceFootprintRule{},
	}
}

func packageAgeDays(pkg npm.PackageRecord, now time.Time) int {
	if pkg.CreatedAt.IsZero() {
		return 0
	}
	return int(now.UTC().Sub(pkg.CreatedAt.UTC()).Hours() / 24)
}

func daysSince(t time.Time, now time.Time) int {
	if t.IsZero() {
		return 0
	}
	return int(now.UTC().Sub(t.UTC()).Hours() / 24)
}

func hasAnyValue(values ...string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return true
		}
	}
	return false
}

func hasAllValue(values ...string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return false
		}
	}
	return true
}

type MissingGitHubRepositoryMetadataRule struct{ BaseRule }

func (r MissingGitHubRepositoryMetadataRule) init() MissingGitHubRepositoryMetadataRule {
	r.BaseRule = BaseRule{"missing_github_repository_metadata", "Missing GitHub repository metadata", SeverityCritical, "Package has no usable GitHub repository link or no cached GitHub metadata, so provenance and maintenance signals cannot be verified."}
	return r
}
func (r MissingGitHubRepositoryMetadataRule) ID() string { return r.init().BaseRule.ID() }
func (r MissingGitHubRepositoryMetadataRule) ToJSON() AlertDefinition {
	return r.init().BaseRule.ToJSON()
}
func (r MissingGitHubRepositoryMetadataRule) Check(pkg npm.PackageRecord, repo *githubdata.RepositoryMetadata, now time.Time) bool {
	return strings.TrimSpace(pkg.Links.Repo) == "" || repo == nil
}

type VeryNewPackageHighRule struct{ BaseRule }

func (r VeryNewPackageHighRule) init() VeryNewPackageHighRule {
	r.BaseRule = BaseRule{"very_new_package_high", "Very new package", SeverityHigh, "Package was published less than 30 days ago and has had little time for scrutiny."}
	return r
}
func (r VeryNewPackageHighRule) ID() string              { return r.init().BaseRule.ID() }
func (r VeryNewPackageHighRule) ToJSON() AlertDefinition { return r.init().BaseRule.ToJSON() }
func (r VeryNewPackageHighRule) Check(pkg npm.PackageRecord, repo *githubdata.RepositoryMetadata, now time.Time) bool {
	age := packageAgeDays(pkg, now)
	return age > 0 && age < 30
}

type VeryNewPackageMediumRule struct{ BaseRule }

func (r VeryNewPackageMediumRule) init() VeryNewPackageMediumRule {
	r.BaseRule = BaseRule{"very_new_package_medium", "Relatively new package", SeverityMedium, "Package was published between 30 and 90 days ago and is still early in its lifecycle."}
	return r
}
func (r VeryNewPackageMediumRule) ID() string              { return r.init().BaseRule.ID() }
func (r VeryNewPackageMediumRule) ToJSON() AlertDefinition { return r.init().BaseRule.ToJSON() }
func (r VeryNewPackageMediumRule) Check(pkg npm.PackageRecord, repo *githubdata.RepositoryMetadata, now time.Time) bool {
	age := packageAgeDays(pkg, now)
	return age >= 30 && age <= 90
}

type VeryNewRepositoryHighRule struct{ BaseRule }

func (r VeryNewRepositoryHighRule) init() VeryNewRepositoryHighRule {
	r.BaseRule = BaseRule{"very_new_repository_high", "Very new repository", SeverityHigh, "Repository is less than 30 days old and likely not battle-tested."}
	return r
}
func (r VeryNewRepositoryHighRule) ID() string              { return r.init().BaseRule.ID() }
func (r VeryNewRepositoryHighRule) ToJSON() AlertDefinition { return r.init().BaseRule.ToJSON() }
func (r VeryNewRepositoryHighRule) Check(pkg npm.PackageRecord, repo *githubdata.RepositoryMetadata, now time.Time) bool {
	return repo != nil && repo.AgeDays > 0 && repo.AgeDays < 30
}

type VeryNewRepositoryMediumRule struct{ BaseRule }

func (r VeryNewRepositoryMediumRule) init() VeryNewRepositoryMediumRule {
	r.BaseRule = BaseRule{"very_new_repository_medium", "Relatively new repository", SeverityMedium, "Repository is between 30 and 90 days old and still early in maturity."}
	return r
}
func (r VeryNewRepositoryMediumRule) ID() string              { return r.init().BaseRule.ID() }
func (r VeryNewRepositoryMediumRule) ToJSON() AlertDefinition { return r.init().BaseRule.ToJSON() }
func (r VeryNewRepositoryMediumRule) Check(pkg npm.PackageRecord, repo *githubdata.RepositoryMetadata, now time.Time) bool {
	return repo != nil && repo.AgeDays >= 30 && repo.AgeDays <= 90
}

type NoMaintainerRule struct{ BaseRule }

func (r NoMaintainerRule) init() NoMaintainerRule {
	r.BaseRule = BaseRule{"no_maintainer_listed", "No maintainer listed", SeverityCritical, "Package exposes no maintainer records, making accountability and recovery difficult."}
	return r
}
func (r NoMaintainerRule) ID() string              { return r.init().BaseRule.ID() }
func (r NoMaintainerRule) ToJSON() AlertDefinition { return r.init().BaseRule.ToJSON() }
func (r NoMaintainerRule) Check(pkg npm.PackageRecord, repo *githubdata.RepositoryMetadata, now time.Time) bool {
	return len(pkg.Maintainers) == 0
}

type SingleMaintainerRule struct{ BaseRule }

func (r SingleMaintainerRule) init() SingleMaintainerRule {
	r.BaseRule = BaseRule{"single_maintainer", "Single maintainer package", SeverityHigh, "Package is controlled by one maintainer, creating high bus-factor and account-compromise risk."}
	return r
}
func (r SingleMaintainerRule) ID() string              { return r.init().BaseRule.ID() }
func (r SingleMaintainerRule) ToJSON() AlertDefinition { return r.init().BaseRule.ToJSON() }
func (r SingleMaintainerRule) Check(pkg npm.PackageRecord, repo *githubdata.RepositoryMetadata, now time.Time) bool {
	return len(pkg.Maintainers) == 1
}

type LowOwnerReputationHighRule struct{ BaseRule }

func (r LowOwnerReputationHighRule) init() LowOwnerReputationHighRule {
	r.BaseRule = BaseRule{"low_owner_reputation_high", "Low owner reputation", SeverityHigh, "Repository owner has zero followers, providing almost no public trust signal."}
	return r
}
func (r LowOwnerReputationHighRule) ID() string              { return r.init().BaseRule.ID() }
func (r LowOwnerReputationHighRule) ToJSON() AlertDefinition { return r.init().BaseRule.ToJSON() }
func (r LowOwnerReputationHighRule) Check(pkg npm.PackageRecord, repo *githubdata.RepositoryMetadata, now time.Time) bool {
	return repo != nil && repo.OwnerMetadata.Followers == 0
}

type LowOwnerReputationMediumRule struct{ BaseRule }

func (r LowOwnerReputationMediumRule) init() LowOwnerReputationMediumRule {
	r.BaseRule = BaseRule{"low_owner_reputation_medium", "Low owner reputation", SeverityMedium, "Repository owner has only one to three followers, a weak public trust signal."}
	return r
}
func (r LowOwnerReputationMediumRule) ID() string              { return r.init().BaseRule.ID() }
func (r LowOwnerReputationMediumRule) ToJSON() AlertDefinition { return r.init().BaseRule.ToJSON() }
func (r LowOwnerReputationMediumRule) Check(pkg npm.PackageRecord, repo *githubdata.RepositoryMetadata, now time.Time) bool {
	return repo != nil && repo.OwnerMetadata.Followers >= 1 && repo.OwnerMetadata.Followers <= 3
}

type LowOwnerReputationLowRule struct{ BaseRule }

func (r LowOwnerReputationLowRule) init() LowOwnerReputationLowRule {
	r.BaseRule = BaseRule{"low_owner_reputation_low", "Low owner reputation", SeverityLow, "Repository owner has only four to ten followers, which is still a weak reputation signal."}
	return r
}
func (r LowOwnerReputationLowRule) ID() string              { return r.init().BaseRule.ID() }
func (r LowOwnerReputationLowRule) ToJSON() AlertDefinition { return r.init().BaseRule.ToJSON() }
func (r LowOwnerReputationLowRule) Check(pkg npm.PackageRecord, repo *githubdata.RepositoryMetadata, now time.Time) bool {
	return repo != nil && repo.OwnerMetadata.Followers >= 4 && repo.OwnerMetadata.Followers <= 10
}

type NewOwnerAccountHighRule struct{ BaseRule }

func (r NewOwnerAccountHighRule) init() NewOwnerAccountHighRule {
	r.BaseRule = BaseRule{"new_owner_account_high", "New owner account", SeverityHigh, "Repository owner account is less than 90 days old and may not be established."}
	return r
}
func (r NewOwnerAccountHighRule) ID() string              { return r.init().BaseRule.ID() }
func (r NewOwnerAccountHighRule) ToJSON() AlertDefinition { return r.init().BaseRule.ToJSON() }
func (r NewOwnerAccountHighRule) Check(pkg npm.PackageRecord, repo *githubdata.RepositoryMetadata, now time.Time) bool {
	return repo != nil && repo.OwnerMetadata.AgeDays > 0 && repo.OwnerMetadata.AgeDays < 90
}

type NewOwnerAccountMediumRule struct{ BaseRule }

func (r NewOwnerAccountMediumRule) init() NewOwnerAccountMediumRule {
	r.BaseRule = BaseRule{"new_owner_account_medium", "Relatively new owner account", SeverityMedium, "Repository owner account is between 90 and 365 days old and still building history."}
	return r
}
func (r NewOwnerAccountMediumRule) ID() string              { return r.init().BaseRule.ID() }
func (r NewOwnerAccountMediumRule) ToJSON() AlertDefinition { return r.init().BaseRule.ToJSON() }
func (r NewOwnerAccountMediumRule) Check(pkg npm.PackageRecord, repo *githubdata.RepositoryMetadata, now time.Time) bool {
	return repo != nil && repo.OwnerMetadata.AgeDays >= 90 && repo.OwnerMetadata.AgeDays <= 365
}

type TinyCommitHistoryHighRule struct{ BaseRule }

func (r TinyCommitHistoryHighRule) init() TinyCommitHistoryHighRule {
	r.BaseRule = BaseRule{"tiny_commit_history_high", "Tiny commit history", SeverityHigh, "Repository has three or fewer commits, suggesting minimal development history."}
	return r
}
func (r TinyCommitHistoryHighRule) ID() string              { return r.init().BaseRule.ID() }
func (r TinyCommitHistoryHighRule) ToJSON() AlertDefinition { return r.init().BaseRule.ToJSON() }
func (r TinyCommitHistoryHighRule) Check(pkg npm.PackageRecord, repo *githubdata.RepositoryMetadata, now time.Time) bool {
	return repo != nil && repo.CommitsCount <= 3
}

type TinyCommitHistoryMediumRule struct{ BaseRule }

func (r TinyCommitHistoryMediumRule) init() TinyCommitHistoryMediumRule {
	r.BaseRule = BaseRule{"tiny_commit_history_medium", "Limited commit history", SeverityMedium, "Repository has only four to twenty commits, indicating low development depth."}
	return r
}
func (r TinyCommitHistoryMediumRule) ID() string              { return r.init().BaseRule.ID() }
func (r TinyCommitHistoryMediumRule) ToJSON() AlertDefinition { return r.init().BaseRule.ToJSON() }
func (r TinyCommitHistoryMediumRule) Check(pkg npm.PackageRecord, repo *githubdata.RepositoryMetadata, now time.Time) bool {
	return repo != nil && repo.CommitsCount >= 4 && repo.CommitsCount <= 20
}

type TinyCommitHistoryLowRule struct{ BaseRule }

func (r TinyCommitHistoryLowRule) init() TinyCommitHistoryLowRule {
	r.BaseRule = BaseRule{"tiny_commit_history_low", "Shallow commit history", SeverityLow, "Repository has only twenty-one to fifty commits, still a modest maintenance history."}
	return r
}
func (r TinyCommitHistoryLowRule) ID() string              { return r.init().BaseRule.ID() }
func (r TinyCommitHistoryLowRule) ToJSON() AlertDefinition { return r.init().BaseRule.ToJSON() }
func (r TinyCommitHistoryLowRule) Check(pkg npm.PackageRecord, repo *githubdata.RepositoryMetadata, now time.Time) bool {
	return repo != nil && repo.CommitsCount >= 21 && repo.CommitsCount <= 50
}

type NoReleasesRule struct{ BaseRule }

func (r NoReleasesRule) init() NoReleasesRule {
	r.BaseRule = BaseRule{"no_releases", "No releases", SeverityMedium, "Repository has no tagged releases, suggesting weak release hygiene."}
	return r
}
func (r NoReleasesRule) ID() string              { return r.init().BaseRule.ID() }
func (r NoReleasesRule) ToJSON() AlertDefinition { return r.init().BaseRule.ToJSON() }
func (r NoReleasesRule) Check(pkg npm.PackageRecord, repo *githubdata.RepositoryMetadata, now time.Time) bool {
	return repo != nil && repo.ReleasesCount == 0
}

type NoLicenseRule struct{ BaseRule }

func (r NoLicenseRule) init() NoLicenseRule {
	r.BaseRule = BaseRule{"no_license", "No license", SeverityHigh, "Package has no license, creating legal and maturity risk."}
	return r
}
func (r NoLicenseRule) ID() string              { return r.init().BaseRule.ID() }
func (r NoLicenseRule) ToJSON() AlertDefinition { return r.init().BaseRule.ToJSON() }
func (r NoLicenseRule) Check(pkg npm.PackageRecord, repo *githubdata.RepositoryMetadata, now time.Time) bool {
	return strings.TrimSpace(pkg.License) == ""
}

type LowWeeklyDownloadsHighRule struct{ BaseRule }

func (r LowWeeklyDownloadsHighRule) init() LowWeeklyDownloadsHighRule {
	r.BaseRule = BaseRule{"low_weekly_downloads_high", "Low weekly adoption", SeverityHigh, "Package has zero weekly downloads, implying no observed active usage."}
	return r
}
func (r LowWeeklyDownloadsHighRule) ID() string              { return r.init().BaseRule.ID() }
func (r LowWeeklyDownloadsHighRule) ToJSON() AlertDefinition { return r.init().BaseRule.ToJSON() }
func (r LowWeeklyDownloadsHighRule) Check(pkg npm.PackageRecord, repo *githubdata.RepositoryMetadata, now time.Time) bool {
	return pkg.DownloadsWeekly == 0
}

type LowWeeklyDownloadsMediumRule struct{ BaseRule }

func (r LowWeeklyDownloadsMediumRule) init() LowWeeklyDownloadsMediumRule {
	r.BaseRule = BaseRule{"low_weekly_downloads_medium", "Low weekly adoption", SeverityMedium, "Package has only one to ten weekly downloads, indicating limited active usage."}
	return r
}
func (r LowWeeklyDownloadsMediumRule) ID() string              { return r.init().BaseRule.ID() }
func (r LowWeeklyDownloadsMediumRule) ToJSON() AlertDefinition { return r.init().BaseRule.ToJSON() }
func (r LowWeeklyDownloadsMediumRule) Check(pkg npm.PackageRecord, repo *githubdata.RepositoryMetadata, now time.Time) bool {
	return pkg.DownloadsWeekly >= 1 && pkg.DownloadsWeekly <= 10
}

type LowWeeklyDownloadsLowRule struct{ BaseRule }

func (r LowWeeklyDownloadsLowRule) init() LowWeeklyDownloadsLowRule {
	r.BaseRule = BaseRule{"low_weekly_downloads_low", "Low weekly adoption", SeverityLow, "Package has only eleven to fifty weekly downloads, still a small active footprint."}
	return r
}
func (r LowWeeklyDownloadsLowRule) ID() string              { return r.init().BaseRule.ID() }
func (r LowWeeklyDownloadsLowRule) ToJSON() AlertDefinition { return r.init().BaseRule.ToJSON() }
func (r LowWeeklyDownloadsLowRule) Check(pkg npm.PackageRecord, repo *githubdata.RepositoryMetadata, now time.Time) bool {
	return pkg.DownloadsWeekly >= 11 && pkg.DownloadsWeekly <= 50
}

type LowMonthlyDownloadsMediumRule struct{ BaseRule }

func (r LowMonthlyDownloadsMediumRule) init() LowMonthlyDownloadsMediumRule {
	r.BaseRule = BaseRule{"low_monthly_downloads_medium", "Low monthly adoption", SeverityMedium, "Package has fewer than twenty monthly downloads, indicating almost no ecosystem footprint."}
	return r
}
func (r LowMonthlyDownloadsMediumRule) ID() string              { return r.init().BaseRule.ID() }
func (r LowMonthlyDownloadsMediumRule) ToJSON() AlertDefinition { return r.init().BaseRule.ToJSON() }
func (r LowMonthlyDownloadsMediumRule) Check(pkg npm.PackageRecord, repo *githubdata.RepositoryMetadata, now time.Time) bool {
	return pkg.DownloadsMonthly < 20
}

type LowMonthlyDownloadsLowRule struct{ BaseRule }

func (r LowMonthlyDownloadsLowRule) init() LowMonthlyDownloadsLowRule {
	r.BaseRule = BaseRule{"low_monthly_downloads_low", "Low monthly adoption", SeverityLow, "Package has only twenty to one hundred monthly downloads, still a weak adoption signal."}
	return r
}
func (r LowMonthlyDownloadsLowRule) ID() string              { return r.init().BaseRule.ID() }
func (r LowMonthlyDownloadsLowRule) ToJSON() AlertDefinition { return r.init().BaseRule.ToJSON() }
func (r LowMonthlyDownloadsLowRule) Check(pkg npm.PackageRecord, repo *githubdata.RepositoryMetadata, now time.Time) bool {
	return pkg.DownloadsMonthly >= 20 && pkg.DownloadsMonthly <= 100
}

type StalePackageUpdatesHighRule struct{ BaseRule }

func (r StalePackageUpdatesHighRule) init() StalePackageUpdatesHighRule {
	r.BaseRule = BaseRule{"stale_package_updates_high", "Stale package updates", SeverityHigh, "Package has not been updated for more than a year."}
	return r
}
func (r StalePackageUpdatesHighRule) ID() string              { return r.init().BaseRule.ID() }
func (r StalePackageUpdatesHighRule) ToJSON() AlertDefinition { return r.init().BaseRule.ToJSON() }
func (r StalePackageUpdatesHighRule) Check(pkg npm.PackageRecord, repo *githubdata.RepositoryMetadata, now time.Time) bool {
	days := daysSince(pkg.UpdatedAt, now)
	return days > 365
}

type StalePackageUpdatesMediumRule struct{ BaseRule }

func (r StalePackageUpdatesMediumRule) init() StalePackageUpdatesMediumRule {
	r.BaseRule = BaseRule{"stale_package_updates_medium", "Stale package updates", SeverityMedium, "Package has not been updated for 180 to 365 days."}
	return r
}
func (r StalePackageUpdatesMediumRule) ID() string              { return r.init().BaseRule.ID() }
func (r StalePackageUpdatesMediumRule) ToJSON() AlertDefinition { return r.init().BaseRule.ToJSON() }
func (r StalePackageUpdatesMediumRule) Check(pkg npm.PackageRecord, repo *githubdata.RepositoryMetadata, now time.Time) bool {
	days := daysSince(pkg.UpdatedAt, now)
	return days >= 180 && days <= 365
}

type StalePackageUpdatesLowRule struct{ BaseRule }

func (r StalePackageUpdatesLowRule) init() StalePackageUpdatesLowRule {
	r.BaseRule = BaseRule{"stale_package_updates_low", "Aging package updates", SeverityLow, "Package has not been updated for 90 to 179 days."}
	return r
}
func (r StalePackageUpdatesLowRule) ID() string              { return r.init().BaseRule.ID() }
func (r StalePackageUpdatesLowRule) ToJSON() AlertDefinition { return r.init().BaseRule.ToJSON() }
func (r StalePackageUpdatesLowRule) Check(pkg npm.PackageRecord, repo *githubdata.RepositoryMetadata, now time.Time) bool {
	days := daysSince(pkg.UpdatedAt, now)
	return days >= 90 && days < 180
}

type LowStarsMediumRule struct{ BaseRule }

func (r LowStarsMediumRule) init() LowStarsMediumRule {
	r.BaseRule = BaseRule{"low_stars_medium", "Low GitHub attention", SeverityMedium, "Repository has zero stars, giving no external validation signal."}
	return r
}
func (r LowStarsMediumRule) ID() string              { return r.init().BaseRule.ID() }
func (r LowStarsMediumRule) ToJSON() AlertDefinition { return r.init().BaseRule.ToJSON() }
func (r LowStarsMediumRule) Check(pkg npm.PackageRecord, repo *githubdata.RepositoryMetadata, now time.Time) bool {
	return repo != nil && repo.Stars == 0
}

type LowStarsLowRule struct{ BaseRule }

func (r LowStarsLowRule) init() LowStarsLowRule {
	r.BaseRule = BaseRule{"low_stars_low", "Low GitHub attention", SeverityLow, "Repository has only one to four stars, still a weak validation signal."}
	return r
}
func (r LowStarsLowRule) ID() string              { return r.init().BaseRule.ID() }
func (r LowStarsLowRule) ToJSON() AlertDefinition { return r.init().BaseRule.ToJSON() }
func (r LowStarsLowRule) Check(pkg npm.PackageRecord, repo *githubdata.RepositoryMetadata, now time.Time) bool {
	return repo != nil && repo.Stars >= 1 && repo.Stars <= 4
}

type LowForksLowRule struct{ BaseRule }

func (r LowForksLowRule) init() LowForksLowRule {
	r.BaseRule = BaseRule{"low_forks_low", "Low fork activity", SeverityLow, "Repository has zero forks, suggesting little downstream engagement."}
	return r
}
func (r LowForksLowRule) ID() string              { return r.init().BaseRule.ID() }
func (r LowForksLowRule) ToJSON() AlertDefinition { return r.init().BaseRule.ToJSON() }
func (r LowForksLowRule) Check(pkg npm.PackageRecord, repo *githubdata.RepositoryMetadata, now time.Time) bool {
	return repo != nil && repo.Forks == 0 && packageAgeDays(pkg, now) <= 180
}

type LowForksMediumOldPackageRule struct{ BaseRule }

func (r LowForksMediumOldPackageRule) init() LowForksMediumOldPackageRule {
	r.BaseRule = BaseRule{"low_forks_medium_old_package", "Low fork activity on older package", SeverityMedium, "Repository has zero forks despite the package being older than 180 days."}
	return r
}
func (r LowForksMediumOldPackageRule) ID() string              { return r.init().BaseRule.ID() }
func (r LowForksMediumOldPackageRule) ToJSON() AlertDefinition { return r.init().BaseRule.ToJSON() }
func (r LowForksMediumOldPackageRule) Check(pkg npm.PackageRecord, repo *githubdata.RepositoryMetadata, now time.Time) bool {
	return repo != nil && repo.Forks == 0 && packageAgeDays(pkg, now) > 180
}

type LowWatchesRule struct{ BaseRule }

func (r LowWatchesRule) init() LowWatchesRule {
	r.BaseRule = BaseRule{"low_watches", "Low watch activity", SeverityLow, "Repository has zero watchers, indicating low ongoing community attention."}
	return r
}
func (r LowWatchesRule) ID() string              { return r.init().BaseRule.ID() }
func (r LowWatchesRule) ToJSON() AlertDefinition { return r.init().BaseRule.ToJSON() }
func (r LowWatchesRule) Check(pkg npm.PackageRecord, repo *githubdata.RepositoryMetadata, now time.Time) bool {
	return repo != nil && repo.Watches == 0
}

type LowOwnerTrackRecordMediumRule struct{ BaseRule }

func (r LowOwnerTrackRecordMediumRule) init() LowOwnerTrackRecordMediumRule {
	r.BaseRule = BaseRule{"low_owner_track_record_medium", "Owner has little track record", SeverityMedium, "Owner has three or fewer public repositories, indicating a limited visible history."}
	return r
}
func (r LowOwnerTrackRecordMediumRule) ID() string              { return r.init().BaseRule.ID() }
func (r LowOwnerTrackRecordMediumRule) ToJSON() AlertDefinition { return r.init().BaseRule.ToJSON() }
func (r LowOwnerTrackRecordMediumRule) Check(pkg npm.PackageRecord, repo *githubdata.RepositoryMetadata, now time.Time) bool {
	return repo != nil && repo.OwnerMetadata.PublicRepos <= 3
}

type LowOwnerTrackRecordLowRule struct{ BaseRule }

func (r LowOwnerTrackRecordLowRule) init() LowOwnerTrackRecordLowRule {
	r.BaseRule = BaseRule{"low_owner_track_record_low", "Owner has modest track record", SeverityLow, "Owner has only four to ten public repositories, still a fairly small track record."}
	return r
}
func (r LowOwnerTrackRecordLowRule) ID() string              { return r.init().BaseRule.ID() }
func (r LowOwnerTrackRecordLowRule) ToJSON() AlertDefinition { return r.init().BaseRule.ToJSON() }
func (r LowOwnerTrackRecordLowRule) Check(pkg npm.PackageRecord, repo *githubdata.RepositoryMetadata, now time.Time) bool {
	return repo != nil && repo.OwnerMetadata.PublicRepos >= 4 && repo.OwnerMetadata.PublicRepos <= 10
}

type PackageRepoAgeMismatchMediumRule struct{ BaseRule }

func (r PackageRepoAgeMismatchMediumRule) init() PackageRepoAgeMismatchMediumRule {
	r.BaseRule = BaseRule{"package_repo_age_mismatch_medium", "Large package-repository age mismatch", SeverityMedium, "Package age differs from repository age by more than two years, which is anomalous enough to review."}
	return r
}
func (r PackageRepoAgeMismatchMediumRule) ID() string              { return r.init().BaseRule.ID() }
func (r PackageRepoAgeMismatchMediumRule) ToJSON() AlertDefinition { return r.init().BaseRule.ToJSON() }
func (r PackageRepoAgeMismatchMediumRule) Check(pkg npm.PackageRecord, repo *githubdata.RepositoryMetadata, now time.Time) bool {
	if repo == nil {
		return false
	}
	diff := packageAgeDays(pkg, now) - repo.AgeDays
	if diff < 0 {
		diff = -diff
	}
	return diff > 730
}

type PackageRepoAgeMismatchLowRule struct{ BaseRule }

func (r PackageRepoAgeMismatchLowRule) init() PackageRepoAgeMismatchLowRule {
	r.BaseRule = BaseRule{"package_repo_age_mismatch_low", "Package-repository age mismatch", SeverityLow, "Package age differs from repository age by more than one year, which is mildly anomalous."}
	return r
}
func (r PackageRepoAgeMismatchLowRule) ID() string              { return r.init().BaseRule.ID() }
func (r PackageRepoAgeMismatchLowRule) ToJSON() AlertDefinition { return r.init().BaseRule.ToJSON() }
func (r PackageRepoAgeMismatchLowRule) Check(pkg npm.PackageRecord, repo *githubdata.RepositoryMetadata, now time.Time) bool {
	if repo == nil {
		return false
	}
	diff := packageAgeDays(pkg, now) - repo.AgeDays
	if diff < 0 {
		diff = -diff
	}
	return diff > 365 && diff <= 730
}

type DownloadSpikeHighRule struct{ BaseRule }

func (r DownloadSpikeHighRule) init() DownloadSpikeHighRule {
	r.BaseRule = BaseRule{"download_spike_high", "Download spike versus maturity", SeverityHigh, "Very new package already has unusually high monthly downloads relative to age."}
	return r
}
func (r DownloadSpikeHighRule) ID() string              { return r.init().BaseRule.ID() }
func (r DownloadSpikeHighRule) ToJSON() AlertDefinition { return r.init().BaseRule.ToJSON() }
func (r DownloadSpikeHighRule) Check(pkg npm.PackageRecord, repo *githubdata.RepositoryMetadata, now time.Time) bool {
	age := packageAgeDays(pkg, now)
	return age > 0 && age < 30 && pkg.DownloadsMonthly > 1000
}

type DownloadSpikeCriticalRule struct{ BaseRule }

func (r DownloadSpikeCriticalRule) init() DownloadSpikeCriticalRule {
	r.BaseRule = BaseRule{"download_spike_critical", "Extreme download spike versus maturity", SeverityCritical, "Package younger than 90 days has extremely high monthly downloads, which warrants urgent review."}
	return r
}
func (r DownloadSpikeCriticalRule) ID() string              { return r.init().BaseRule.ID() }
func (r DownloadSpikeCriticalRule) ToJSON() AlertDefinition { return r.init().BaseRule.ToJSON() }
func (r DownloadSpikeCriticalRule) Check(pkg npm.PackageRecord, repo *githubdata.RepositoryMetadata, now time.Time) bool {
	age := packageAgeDays(pkg, now)
	return age > 0 && age < 90 && pkg.DownloadsMonthly > 5000
}

type MaintainerMissingAllIdentityRule struct{ BaseRule }

func (r MaintainerMissingAllIdentityRule) init() MaintainerMissingAllIdentityRule {
	r.BaseRule = BaseRule{"maintainer_missing_all_identity", "Maintainer missing all identity fields", SeverityHigh, "At least one maintainer record is missing both username and email."}
	return r
}
func (r MaintainerMissingAllIdentityRule) ID() string              { return r.init().BaseRule.ID() }
func (r MaintainerMissingAllIdentityRule) ToJSON() AlertDefinition { return r.init().BaseRule.ToJSON() }
func (r MaintainerMissingAllIdentityRule) Check(pkg npm.PackageRecord, repo *githubdata.RepositoryMetadata, now time.Time) bool {
	for _, m := range pkg.Maintainers {
		if strings.TrimSpace(m.Username) == "" && strings.TrimSpace(m.Email) == "" {
			return true
		}
	}
	return false
}

type MaintainerMissingPartialIdentityRule struct{ BaseRule }

func (r MaintainerMissingPartialIdentityRule) init() MaintainerMissingPartialIdentityRule {
	r.BaseRule = BaseRule{"maintainer_missing_partial_identity", "Maintainers have partial identity data", SeverityMedium, "All maintainer records are missing either username or email, reducing accountability."}
	return r
}
func (r MaintainerMissingPartialIdentityRule) ID() string { return r.init().BaseRule.ID() }
func (r MaintainerMissingPartialIdentityRule) ToJSON() AlertDefinition {
	return r.init().BaseRule.ToJSON()
}
func (r MaintainerMissingPartialIdentityRule) Check(pkg npm.PackageRecord, repo *githubdata.RepositoryMetadata, now time.Time) bool {
	if len(pkg.Maintainers) == 0 {
		return false
	}
	for _, m := range pkg.Maintainers {
		if hasAllValue(m.Username, m.Email) {
			return false
		}
	}
	return true
}

type MissingDescriptionRule struct{ BaseRule }

func (r MissingDescriptionRule) init() MissingDescriptionRule {
	r.BaseRule = BaseRule{"missing_description", "Missing description", SeverityMedium, "Package has no description, making intent and scope less transparent."}
	return r
}
func (r MissingDescriptionRule) ID() string              { return r.init().BaseRule.ID() }
func (r MissingDescriptionRule) ToJSON() AlertDefinition { return r.init().BaseRule.ToJSON() }
func (r MissingDescriptionRule) Check(pkg npm.PackageRecord, repo *githubdata.RepositoryMetadata, now time.Time) bool {
	return strings.TrimSpace(pkg.Description) == ""
}

type BrokenTrustLinkSurfaceMediumRule struct{ BaseRule }

func (r BrokenTrustLinkSurfaceMediumRule) init() BrokenTrustLinkSurfaceMediumRule {
	r.BaseRule = BaseRule{"broken_trust_link_surface_medium", "Missing trust links", SeverityMedium, "Package is missing both homepage and bug tracker links, reducing auditability."}
	return r
}
func (r BrokenTrustLinkSurfaceMediumRule) ID() string              { return r.init().BaseRule.ID() }
func (r BrokenTrustLinkSurfaceMediumRule) ToJSON() AlertDefinition { return r.init().BaseRule.ToJSON() }
func (r BrokenTrustLinkSurfaceMediumRule) Check(pkg npm.PackageRecord, repo *githubdata.RepositoryMetadata, now time.Time) bool {
	return !hasAnyValue(pkg.Links.Homepage) && !hasAnyValue(pkg.Links.Bugs)
}

type BrokenTrustLinkSurfaceCriticalRule struct{ BaseRule }

func (r BrokenTrustLinkSurfaceCriticalRule) init() BrokenTrustLinkSurfaceCriticalRule {
	r.BaseRule = BaseRule{"broken_trust_link_surface_critical", "No trust links", SeverityCritical, "Package is missing homepage, bug tracker, and repository links, leaving no external verification surface."}
	return r
}
func (r BrokenTrustLinkSurfaceCriticalRule) ID() string { return r.init().BaseRule.ID() }
func (r BrokenTrustLinkSurfaceCriticalRule) ToJSON() AlertDefinition {
	return r.init().BaseRule.ToJSON()
}
func (r BrokenTrustLinkSurfaceCriticalRule) Check(pkg npm.PackageRecord, repo *githubdata.RepositoryMetadata, now time.Time) bool {
	return !hasAnyValue(pkg.Links.Homepage, pkg.Links.Bugs, pkg.Links.Repo)
}

type WeakConsistencyAcrossSignalsRule struct{ BaseRule }

func (r WeakConsistencyAcrossSignalsRule) init() WeakConsistencyAcrossSignalsRule {
	r.BaseRule = BaseRule{"weak_consistency_across_signals", "Weak consistency across signals", SeverityHigh, "Older package still has near-zero adoption, stars, forks, releases, and commits, which strongly suggests abandonment or low credibility."}
	return r
}
func (r WeakConsistencyAcrossSignalsRule) ID() string              { return r.init().BaseRule.ID() }
func (r WeakConsistencyAcrossSignalsRule) ToJSON() AlertDefinition { return r.init().BaseRule.ToJSON() }
func (r WeakConsistencyAcrossSignalsRule) Check(pkg npm.PackageRecord, repo *githubdata.RepositoryMetadata, now time.Time) bool {
	if repo == nil {
		return false
	}
	return packageAgeDays(pkg, now) > 180 && pkg.DownloadsMonthly < 20 && repo.Stars == 0 && repo.Forks == 0 && repo.ReleasesCount == 0 && repo.CommitsCount <= 3
}

type LowMaintenanceFootprintRule struct{ BaseRule }

func (r LowMaintenanceFootprintRule) init() LowMaintenanceFootprintRule {
	r.BaseRule = BaseRule{"low_maintenance_footprint", "Low-maintenance footprint", SeverityHigh, "Old package has stale updates, no releases, and limited commit history, indicating likely abandonment."}
	return r
}
func (r LowMaintenanceFootprintRule) ID() string              { return r.init().BaseRule.ID() }
func (r LowMaintenanceFootprintRule) ToJSON() AlertDefinition { return r.init().BaseRule.ToJSON() }
func (r LowMaintenanceFootprintRule) Check(pkg npm.PackageRecord, repo *githubdata.RepositoryMetadata, now time.Time) bool {
	if repo == nil {
		return false
	}
	return packageAgeDays(pkg, now) > 365 && daysSince(pkg.UpdatedAt, now) > 180 && repo.ReleasesCount == 0 && repo.CommitsCount <= 20
}
