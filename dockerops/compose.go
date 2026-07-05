package dockerops

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"docker-go/domains"
)

func (e *Executor) ComposeValidate(ctx context.Context, rawPayload json.RawMessage) (domains.ComposeTaskResult, error) {
	payload, err := parseComposeTaskPayload(rawPayload)
	if err != nil {
		return domains.ComposeTaskResult{}, err
	}
	payload.Action = "validate"
	result, err := e.prepareComposeResult(payload)
	if err != nil {
		return domains.ComposeTaskResult{}, err
	}
	step, err := runComposeCommand(ctx, result.Workdir, result.ProjectName, payload.Profiles, []string{"config"})
	result.Steps = append(result.Steps, step)
	if err != nil {
		return result, err
	}
	return result, nil
}

func (e *Executor) ComposeAction(ctx context.Context, taskType string, rawPayload json.RawMessage) (domains.ComposeTaskResult, error) {
	payload, err := parseComposeTaskPayload(rawPayload)
	if err != nil {
		return domains.ComposeTaskResult{}, err
	}
	payload.Action = actionFromTaskType(taskType, payload.Action)
	result, err := e.prepareComposeResult(payload)
	if err != nil {
		return domains.ComposeTaskResult{}, err
	}
	profiles := payload.Profiles
	services := payload.Services
	run := func(args []string) error {
		step, stepErr := runComposeCommand(ctx, result.Workdir, result.ProjectName, profiles, args)
		result.Steps = append(result.Steps, step)
		return stepErr
	}

	switch payload.Action {
	case domains.DeployReleaseActionPull:
		err = run(append([]string{"pull"}, services...))
	case domains.DeployReleaseActionDown:
		args := []string{"down"}
		if payload.RemoveOrphans {
			args = append(args, "--remove-orphans")
		}
		err = run(args)
	case domains.DeployReleaseActionRestart:
		err = run(append([]string{"restart"}, services...))
	default:
		if payload.Pull {
			if err = run(append([]string{"pull"}, services...)); err != nil {
				return result, err
			}
		}
		args := []string{"up", "-d"}
		if payload.RemoveOrphans {
			args = append(args, "--remove-orphans")
		}
		err = run(append(args, services...))
	}
	if err != nil {
		return result, err
	}
	return result, nil
}

func (e *Executor) prepareComposeResult(payload domains.ComposeTaskPayload) (domains.ComposeTaskResult, error) {
	projectName := normalizeComposeProjectName(payload.ProjectName, payload.ConfigGuid, payload.VersionGuid)
	workdir := filepath.Join(e.workspaceDir, projectName)
	if err := os.MkdirAll(workdir, 0700); err != nil {
		return domains.ComposeTaskResult{}, err
	}
	composeFile := filepath.Join(workdir, "compose.yaml")
	if err := os.WriteFile(composeFile, []byte(payload.Content), 0600); err != nil {
		return domains.ComposeTaskResult{}, err
	}
	return domains.ComposeTaskResult{
		ConfigGuid:  payload.ConfigGuid,
		VersionGuid: payload.VersionGuid,
		ProjectName: projectName,
		Action:      payload.Action,
		Workdir:     workdir,
		Steps:       []domains.ComposeCommandStep{},
	}, nil
}

func runComposeCommand(ctx context.Context, workdir string, projectName string, profiles []string, args []string) (domains.ComposeCommandStep, error) {
	if len(args) == 0 {
		return domains.ComposeCommandStep{}, errors.New("missing docker compose command")
	}
	composeFile := filepath.Join(workdir, "compose.yaml")
	fullArgs := []string{"compose", "-f", composeFile, "-p", projectName}
	for _, profile := range profiles {
		fullArgs = append(fullArgs, "--profile", profile)
	}
	fullArgs = append(fullArgs, args...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	startedAt := time.Now().UnixMilli()
	cmd := exec.CommandContext(ctx, "docker", fullArgs...)
	cmd.Dir = workdir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	finishedAt := time.Now().UnixMilli()
	exitCode := 0
	if cmd.ProcessState != nil {
		exitCode = cmd.ProcessState.ExitCode()
	} else if err != nil {
		exitCode = -1
	}
	step := domains.ComposeCommandStep{
		Name:       args[0],
		Command:    append([]string{"docker"}, fullArgs...),
		Stdout:     stdout.String(),
		Stderr:     stderr.String(),
		ExitCode:   exitCode,
		StartedAt:  startedAt,
		FinishedAt: finishedAt,
	}
	if err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = strings.TrimSpace(stdout.String())
		}
		if message == "" {
			message = err.Error()
		}
		return step, fmt.Errorf("docker compose %s failed: %s", strings.Join(args, " "), message)
	}
	return step, nil
}

func parseComposeTaskPayload(raw json.RawMessage) (domains.ComposeTaskPayload, error) {
	var payload domains.ComposeTaskPayload
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &payload); err != nil {
			return payload, err
		}
	}
	payload.ConfigGuid = strings.TrimSpace(payload.ConfigGuid)
	payload.VersionGuid = strings.TrimSpace(payload.VersionGuid)
	payload.ProjectName = strings.TrimSpace(payload.ProjectName)
	payload.Action = normalizeDockerComposeAction(payload.Action)
	services, err := cleanComposeNames(payload.Services, "service")
	if err != nil {
		return payload, err
	}
	profiles, err := cleanComposeNames(payload.Profiles, "profile")
	if err != nil {
		return payload, err
	}
	payload.Services = services
	payload.Profiles = profiles
	if strings.TrimSpace(payload.Content) == "" {
		return payload, errors.New("missing compose yaml content")
	}
	return payload, nil
}

func actionFromTaskType(taskType string, action string) string {
	switch taskType {
	case domains.TaskTypeDockerComposeDown:
		return domains.DeployReleaseActionDown
	case domains.TaskTypeDockerComposeRestart:
		return domains.DeployReleaseActionRestart
	case domains.TaskTypeDockerComposePull:
		return domains.DeployReleaseActionPull
	case domains.TaskTypeDockerComposeUp:
		return domains.DeployReleaseActionUp
	default:
		return normalizeDockerComposeAction(action)
	}
}

func normalizeDockerComposeAction(action string) string {
	switch strings.TrimSpace(action) {
	case domains.DeployReleaseActionDown:
		return domains.DeployReleaseActionDown
	case domains.DeployReleaseActionRestart:
		return domains.DeployReleaseActionRestart
	case domains.DeployReleaseActionPull:
		return domains.DeployReleaseActionPull
	default:
		return domains.DeployReleaseActionUp
	}
}

func normalizeWorkspaceDir(value string) string {
	value = strings.TrimSpace(value)
	if value != "" {
		if abs, err := filepath.Abs(value); err == nil {
			return abs
		}
		return value
	}
	if home, err := os.UserHomeDir(); err == nil && strings.TrimSpace(home) != "" {
		return filepath.Join(home, ".nav-docker", "workspaces")
	}
	return filepath.Join(os.TempDir(), "nav-docker-workspaces")
}

func normalizeComposeProjectName(values ...string) string {
	source := ""
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			source = strings.TrimSpace(value)
			break
		}
	}
	source = strings.ToLower(source)
	var builder strings.Builder
	lastSep := false
	for _, r := range source {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
			lastSep = false
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
			lastSep = false
		case r == '-' || r == '_':
			if !lastSep {
				builder.WriteRune(r)
				lastSep = true
			}
		default:
			if !lastSep {
				builder.WriteByte('-')
				lastSep = true
			}
		}
	}
	result := strings.Trim(builder.String(), "-_")
	if result == "" {
		result = "nav-docker"
	}
	first := result[0]
	if !((first >= 'a' && first <= 'z') || (first >= '0' && first <= '9')) {
		result = "nav-" + result
	}
	if len(result) > 80 {
		result = strings.Trim(result[:80], "-_")
	}
	return result
}

func cleanComposeNames(items []string, label string) ([]string, error) {
	result := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if strings.HasPrefix(item, "-") {
			return nil, fmt.Errorf("invalid compose %s: %s", label, item)
		}
		for _, r := range item {
			if (r >= 'a' && r <= 'z') ||
				(r >= 'A' && r <= 'Z') ||
				(r >= '0' && r <= '9') ||
				r == '-' ||
				r == '_' ||
				r == '.' {
				continue
			}
			return nil, fmt.Errorf("invalid compose %s: %s", label, item)
		}
		result = append(result, item)
	}
	return result, nil
}
