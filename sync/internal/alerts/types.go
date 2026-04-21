package alerts

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"sync/internal/githubdata"
	"sync/internal/npm"
)

type Severity string

const (
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

type AlertDefinition struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Severity    Severity `json:"severity"`
}

type AlertRule interface {
	ID() string
	ToJSON() AlertDefinition
	Check(pkg npm.PackageRecord, repo *githubdata.RepositoryMetadata, now time.Time) bool
}

type BaseRule struct {
	id          string
	name        string
	severity    Severity
	description string
}

func (r BaseRule) ID() string { return r.id }

func (r BaseRule) ToJSON() AlertDefinition {
	return AlertDefinition{
		ID:          r.id,
		Name:        r.name,
		Description: r.description,
		Severity:    r.severity,
	}
}

type PackageAlerts struct {
	AlertIDs []string `json:"alert_ids"`
}

type AlertsFile struct {
	Date        string                     `json:"date"`
	Definitions map[string]AlertDefinition `json:"definitions"`
	Packages    map[string]PackageAlerts   `json:"packages"`
}

func WriteFile(path string, alertsFile AlertsFile) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create alerts dir: %w", err)
	}
	data, err := json.MarshalIndent(alertsFile, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal alerts file: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write alerts file %s: %w", path, err)
	}
	return nil
}
