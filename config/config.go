package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
)

const DefaultPath = "vauban.json"

type Config struct {
	Jobs map[string]Job `json:"jobs"`
}

type Job struct {
	Name  string `json:"name"`
	Image string `json:"image"`
	Steps []Step `json:"steps"`
}

type Step struct {
	Name string `json:"name"`
	Run  string `json:"run"`
}

func (cfg *Config) Job(name string) (*Job, error) {
	if cfg == nil {
		return nil, errors.New("config is required")
	}

	job, ok := cfg.Jobs[name]
	if !ok {
		return nil, fmt.Errorf("job %q not found", name)
	}

	return &job, nil
}

func (cfg *Config) JobNames() ([]string, error) {
	if cfg == nil {
		return nil, errors.New("config is required")
	}

	names := make([]string, 0, len(cfg.Jobs))
	for name := range cfg.Jobs {
		names = append(names, name)
	}

	slices.Sort(names)
	return names, nil
}

func Load(path string) (*Config, error) {
	if strings.TrimSpace(path) == "" {
		path = DefaultPath
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %q: %w", path, err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("decode config %q: %w", path, err)
	}

	return &cfg, nil
}

func LoadAndValidate(path string) (*Config, error) {
	cfg, err := Load(path)
	if err != nil {
		return nil, err
	}

	if err := Validate(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func Validate(cfg *Config) error {
	if cfg == nil {
		return errors.New("config is required")
	}

	if len(cfg.Jobs) == 0 {
		return errors.New("config.jobs must contain at least one job")
	}

	for key, job := range cfg.Jobs {
		if strings.TrimSpace(key) == "" {
			return errors.New("config.jobs contains an empty job key")
		}

		if strings.TrimSpace(job.Name) == "" {
			return fmt.Errorf("config.jobs.%s.name is required", key)
		}

		if strings.TrimSpace(job.Image) == "" {
			return fmt.Errorf("config.jobs.%s.image is required", key)
		}

		if len(job.Steps) == 0 {
			return fmt.Errorf("config.jobs.%s.steps must contain at least one step", key)
		}

		for i, step := range job.Steps {
			if strings.TrimSpace(step.Name) == "" {
				return fmt.Errorf("config.jobs.%s.steps[%d].name is required", key, i)
			}

			if strings.TrimSpace(step.Run) == "" {
				return fmt.Errorf("config.jobs.%s.steps[%d].run is required", key, i)
			}
		}
	}

	return nil
}
