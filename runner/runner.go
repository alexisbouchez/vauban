package runner

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	dockerclient "github.com/docker/go-sdk/client"
	dockercontainer "github.com/docker/go-sdk/container"
	"github.com/moby/moby/api/pkg/stdcopy"
	apicontainer "github.com/moby/moby/api/types/container"
	mobyclient "github.com/moby/moby/client"

	"vauban/config"
)

const stepTimeout = 24 * time.Hour
const workspaceDir = "/workspace"

type StepRunner interface {
	RunStep(ctx context.Context, jobName string, job *config.Job, step config.Step, stdout, stderr io.Writer) error
}

type Runner struct {
	StepRunner StepRunner
}

func New() *Runner {
	return &Runner{
		StepRunner: DockerStepRunner{},
	}
}

func (r *Runner) RunJob(ctx context.Context, cfg *config.Config, jobName string, stdout, stderr io.Writer) error {
	if r == nil {
		return fmt.Errorf("runner is required")
	}

	if r.StepRunner == nil {
		return fmt.Errorf("step runner is required")
	}

	job, err := cfg.Job(jobName)
	if err != nil {
		return err
	}

	for _, step := range job.Steps {
		if _, err := fmt.Fprintf(stdout, "==> %s\n", step.Name); err != nil {
			return fmt.Errorf("write step header: %w", err)
		}

		if err := r.StepRunner.RunStep(ctx, jobName, job, step, stdout, stderr); err != nil {
			return fmt.Errorf("job %q step %q failed: %w", jobName, step.Name, err)
		}
	}

	return nil
}

type DockerStepRunner struct{}

func (DockerStepRunner) RunStep(ctx context.Context, jobName string, job *config.Job, step config.Step, stdout, stderr io.Writer) error {
	restoreEnv := ensureDockerAuthConfig()
	defer restoreEnv()

	cli, err := newDockerClient(ctx)
	if err != nil {
		return fmt.Errorf("connect docker client: %w", err)
	}

	workspacePath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("resolve workspace: %w", err)
	}

	imageEnv, err := loadImageEnv(ctx, cli, job.Image)
	if err != nil {
		return fmt.Errorf("inspect image env: %w", err)
	}

	ctr, err := dockercontainer.Run(
		ctx,
		dockercontainer.WithClient(cli),
		dockercontainer.WithName(fmt.Sprintf("vauban-%s-%d", jobName, time.Now().UnixNano())),
		dockercontainer.WithImage(job.Image),
		dockercontainer.WithEnv(imageEnv),
		dockercontainer.WithCmd("sh", "-c", step.Run),
		dockercontainer.WithAdditionalConfigModifier(func(cfg *apicontainer.Config) {
			cfg.WorkingDir = workspaceDir
		}),
		dockercontainer.WithAdditionalHostConfigModifier(func(hostConfig *apicontainer.HostConfig) {
			hostConfig.Binds = append(hostConfig.Binds, workspacePath+":"+workspaceDir)
		}),
	)
	if err != nil {
		return fmt.Errorf("start container: %w", err)
	}

	defer func() {
		_ = ctr.Terminate(context.Background())
	}()

	logs, err := cli.ContainerLogs(ctx, ctr.ID(), mobyclient.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	})
	if err != nil {
		return fmt.Errorf("read logs: %w", err)
	}
	defer logs.Close()

	if _, err := stdcopy.StdCopy(stdout, stderr, logs); err != nil {
		return fmt.Errorf("stream logs: %w", err)
	}

	state, err := ctr.State(ctx)
	if err != nil {
		return fmt.Errorf("inspect container state: %w", err)
	}

	if state.ExitCode != 0 {
		return fmt.Errorf("container exited with code %d", state.ExitCode)
	}

	return nil
}

func newDockerClient(ctx context.Context) (dockerclient.SDKClient, error) {
	cli, err := dockerclient.New(ctx)
	if err == nil {
		return cli, nil
	}

	dockerHost := os.Getenv("DOCKER_HOST")
	if dockerHost == "" {
		dockerHost = mobyclient.DefaultDockerHost
	}

	fallback, fallbackErr := dockerclient.New(ctx, dockerclient.WithDockerHost(dockerHost))
	if fallbackErr == nil {
		return fallback, nil
	}

	return nil, fmt.Errorf("default client: %w; fallback client: %w", err, fallbackErr)
}

func ensureDockerAuthConfig() func() {
	if os.Getenv("DOCKER_AUTH_CONFIG") != "" {
		return func() {}
	}

	_ = os.Setenv("DOCKER_AUTH_CONFIG", "{}")
	return func() {
		_ = os.Unsetenv("DOCKER_AUTH_CONFIG")
	}
}

func loadImageEnv(ctx context.Context, cli dockerclient.SDKClient, image string) (map[string]string, error) {
	inspect, err := cli.ImageInspect(ctx, image)
	if err != nil {
		return nil, err
	}

	env := make(map[string]string, len(inspect.Config.Env))
	for _, item := range inspect.Config.Env {
		key, value, ok := strings.Cut(item, "=")
		if !ok || key == "" {
			continue
		}
		env[key] = value
	}

	return env, nil
}
