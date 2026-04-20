package npm

import "time"

type SearchResponse struct {
	Objects []SearchObject `json:"objects"`
	Total   int            `json:"total"`
	Time    string         `json:"time"`
}

type SearchObject struct {
	Downloads Downloads  `json:"downloads"`
	Updated   string     `json:"updated"`
	Package   RawPackage `json:"package"`
}

type Downloads struct {
	Monthly int `json:"monthly"`
	Weekly  int `json:"weekly"`
}

type RawPackage struct {
	Name        string       `json:"name"`
	Version     string       `json:"version"`
	Description string       `json:"description"`
	Maintainers []Maintainer `json:"maintainers"`
	License     string       `json:"license"`
	Date        string       `json:"date"`
	Links       RawLinks     `json:"links"`
}

type Maintainer struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

type RawLinks struct {
	Homepage   string `json:"homepage"`
	Repository string `json:"repository"`
	Bugs       string `json:"bugs"`
	NPM        string `json:"npm"`
}

type PackageLinks struct {
	Homepage string `json:"homepage"`
	Repo     string `json:"repo"`
	Bugs     string `json:"bugs"`
	NPM      string `json:"npm"`
}

type PackageRecord struct {
	Name             string       `json:"name"`
	Version          string       `json:"version"`
	Description      string       `json:"description"`
	License          string       `json:"license"`
	DownloadsMonthly int          `json:"downloads_monthly"`
	DownloadsWeekly  int          `json:"downloads_weekly"`
	UpdatedAt        time.Time    `json:"updated_at"`
	CreatedAt        time.Time    `json:"created_at"`
	Maintainers      []Maintainer `json:"maintainers"`
	Links            PackageLinks `json:"links"`
}
