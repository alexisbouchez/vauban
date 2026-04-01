package config_test

import (
	"strings"
	"testing"
	"vauban/config"
)

func TestValidateAcceptsValidConfig(t *testing.T) {
	cfg := &config.Config{
		Jobs: map[string]config.Job{
			"test": {
				Name:  "Run tests",
				Image: "golang:1.26.1",
				Steps: []config.Step{
					{
						Name: "Execute Go tests",
						Run:  "go test ./...",
					},
				},
			},
		},
	}

	if err := config.Validate(cfg); err != nil {
		t.Fatalf("Validate() returned error: %v", err)
	}
}

func TestValidateRejectsInvalidConfig(t *testing.T) {
	cfg := &config.Config{
		Jobs: map[string]config.Job{
			"test": {
				Name:  "Run tests",
				Image: "",
				Steps: []config.Step{
					{
						Name: "Execute Go tests",
						Run:  "go test ./...",
					},
				},
			},
		},
	}

	if err := config.Validate(cfg); err == nil {
		t.Fatal("Validate() returned nil error for invalid config")
	}
}

func TestJobReturnsNamedJob(t *testing.T) {
	cfg := &config.Config{
		Jobs: map[string]config.Job{
			"test": {
				Name:  "Run tests",
				Image: "golang:1.26.1",
			},
		},
	}

	job, err := cfg.Job("test")
	if err != nil {
		t.Fatalf("Job() returned error: %v", err)
	}

	if job.Name != "Run tests" {
		t.Fatalf("Job() returned unexpected job name: %q", job.Name)
	}
}

func TestJobRejectsMissingJob(t *testing.T) {
	cfg := &config.Config{
		Jobs: map[string]config.Job{},
	}

	if _, err := cfg.Job("missing"); err == nil {
		t.Fatal("Job() returned nil error for missing job")
	}
}

func TestJobNamesReturnsSortedNames(t *testing.T) {
	cfg := &config.Config{
		Jobs: map[string]config.Job{
			"z": {},
			"a": {},
			"m": {},
		},
	}

	names, err := cfg.JobNames()
	if err != nil {
		t.Fatalf("JobNames() returned error: %v", err)
	}

	got := strings.Join(names, ",")
	if got != "a,m,z" {
		t.Fatalf("JobNames() = %q, want %q", got, "a,m,z")
	}
}
