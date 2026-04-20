package npm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func FetchAndExtractPackages(ctx context.Context, client *http.Client, now time.Time) ([]PackageRecord, string, error) {
	data, cacheFile, err := LoadOrFetchRawJSON(ctx, client, now)
	if err != nil {
		return nil, cacheFile, err
	}

	records, err := ExtractPackages(data)
	if err != nil {
		return nil, cacheFile, err
	}

	return records, cacheFile, nil
}

func ExtractPackages(data []byte) ([]PackageRecord, error) {
	var resp SearchResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal npm response: %w", err)
	}

	records := make([]PackageRecord, 0, len(resp.Objects))
	for _, obj := range resp.Objects {
		updatedAt, err := parseTime(obj.Updated)
		if err != nil {
			return nil, fmt.Errorf("parse updated timestamp for %s: %w", obj.Package.Name, err)
		}

		createdAt, err := parseTime(obj.Package.Date)
		if err != nil {
			return nil, fmt.Errorf("parse package date for %s: %w", obj.Package.Name, err)
		}

		records = append(records, PackageRecord{
			Name:             obj.Package.Name,
			Version:          obj.Package.Version,
			Description:      obj.Package.Description,
			License:          obj.Package.License,
			DownloadsMonthly: obj.Downloads.Monthly,
			DownloadsWeekly:  obj.Downloads.Weekly,
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
