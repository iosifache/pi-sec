package npm

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

func FetchAndExtractPackages(ctx context.Context, client *http.Client, now time.Time) ([]PackageRecord, string, error) {
	return LoadOrFetchPackages(ctx, client, now)
}

func ExtractPackages(ctx context.Context, client *http.Client, resp SearchResponse) ([]PackageRecord, error) {
	records := make([]PackageRecord, 0, len(resp.Objects))
	pacer := &requestPacer{}

	for i, obj := range resp.Objects {
		updatedAt, err := parseTime(obj.Updated)
		if err != nil {
			return nil, fmt.Errorf("parse updated timestamp for %s: %w", obj.Package.Name, err)
		}

		createdAt, err := parseTime(obj.Package.Date)
		if err != nil {
			return nil, fmt.Errorf("parse package date for %s: %w", obj.Package.Name, err)
		}

		monthly, err := fetchDownloadCount(ctx, client, pacer, "last-month", obj.Package.Name, i+1, len(resp.Objects))
		if err != nil {
			return nil, err
		}

		weekly, err := fetchDownloadCount(ctx, client, pacer, "last-week", obj.Package.Name, i+1, len(resp.Objects))
		if err != nil {
			return nil, err
		}

		records = append(records, PackageRecord{
			Name:             obj.Package.Name,
			Version:          obj.Package.Version,
			Description:      obj.Package.Description,
			License:          obj.Package.License,
			DownloadsMonthly: monthly,
			DownloadsWeekly:  weekly,
			UpdatedAt:        updatedAt,
			CreatedAt:        createdAt,
			Maintainers:      append([]Maintainer(nil), obj.Package.Maintainers...),
			Links: PackageLinks{
				Homepage: obj.Package.Links.Homepage,
				Repo:     obj.Package.Links.Repository,
				Bugs:     obj.Package.Links.Bugs,
				NPM:      obj.Package.Links.NPM,
			},
		})
	}

	return records, nil
}

func parseTime(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339, value)
}
