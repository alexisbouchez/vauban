package runner

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"vauban/config"
)

type fakeStepRunner struct {
	seen []string
	err  error
}

func (f *fakeStepRunner) RunStep(_ context.Context, _ string, _ *config.Job, step config.Step, _, _ io.Writer) error {
	f.seen = append(f.seen, step.Name)
	return f.err
}

func TestRunJobRunsAllSteps(t *testing.T) {
	cfg := &config.Config{
		Jobs: map[string]config.Job{
			"test": {
				Name:  "Run tests",
				Image: "golang:1.26.1",
				Steps: []config.Step{
					{Name: "one", Run: "echo one"},
					{Name: "two", Run: "echo two"},
				},
			},
		},
	}

	stepRunner := &fakeStepRunner{}
	r := &Runner{StepRunner: stepRunner}
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if err := r.RunJob(context.Background(), cfg, "test", &stdout, &stderr); err != nil {
		t.Fatalf("RunJob() returned error: %v", err)
	}

	if len(stepRunner.seen) != 2 {
		t.Fatalf("RunJob() executed %d steps, want 2", len(stepRunner.seen))
	}
}

func TestRunJobWrapsStepFailure(t *testing.T) {
	cfg := &config.Config{
		Jobs: map[string]config.Job{
			"test": {
				Name:  "Run tests",
				Image: "golang:1.26.1",
				Steps: []config.Step{
					{Name: "one", Run: "exit 1"},
				},
			},
		},
	}

	stepRunner := &fakeStepRunner{err: errors.New("boom")}
	r := &Runner{StepRunner: stepRunner}
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if err := r.RunJob(context.Background(), cfg, "test", &stdout, &stderr); err == nil {
		t.Fatal("RunJob() returned nil error for failed step")
	}
}

func TestRunJobRejectsMissingRunner(t *testing.T) {
	cfg := &config.Config{
		Jobs: map[string]config.Job{
			"test": {
				Name:  "Run tests",
				Image: "golang:1.26.1",
				Steps: []config.Step{{Name: "one", Run: "echo one"}},
			},
		},
	}

	r := &Runner{}
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if err := r.RunJob(context.Background(), cfg, "test", &stdout, &stderr); err == nil {
		t.Fatal("RunJob() returned nil error with missing step runner")
	}
}
