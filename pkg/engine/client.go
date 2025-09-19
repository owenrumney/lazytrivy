package engine

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/aquasecurity/trivy/pkg/commands"
	"github.com/owenrumney/lazytrivy/pkg/config"
	"github.com/owenrumney/lazytrivy/pkg/logger"
	"github.com/owenrumney/lazytrivy/pkg/output"
)

type engineType string

const (
	EngineDocker engineType = "docker"
	EnginePodman engineType = "podman"
)

type Progress interface {
	UpdateStatus(status string)
	ClearStatus()
}

type Client struct {
	imageNames []string
	config     *config.Config
}

func (c *Client) ListImages() ([]string, error) {
	cmds := [][]string{
		{"docker", "images", "--format", "{{.Repository}}:{{.Tag}}"},
		{"podman", "images", "--format", "{{.Repository}}:{{.Tag}}"},
	}

	var output []byte
	var err error

	for _, cmd := range cmds {
		output, err = exec.Command(cmd[0], cmd[1:]...).Output()
		if err == nil {
			break
		}
	}

	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	images := make([]string, 0, len(lines))
	seen := make(map[string]struct{})
	for _, line := range lines {
		img := strings.TrimSpace(line)
		if img == "" || strings.Contains(img, "<none>") {
			continue
		}
		if _, exists := seen[img]; !exists {
			images = append(images, img)
			seen[img] = struct{}{}
		}
	}
	return images, nil
}

func NewClient(cfg *config.Config) (*Client, error) {
	return &Client{
		config: cfg,
	}, nil
}

func (c *Client) scan(ctx context.Context, args []string, scanTarget string, env []string, progress Progress, scanImageName string, engineType engineType, additionalBinds ...string) (*output.Report, error) {

	app := commands.NewApp()

	tempFile := filepath.Join(os.TempDir(), "lazytrivy-output.json")
	defer func() { _ = os.Remove(tempFile) }()
	args = append(args, "--output", tempFile, "-q")

	logger.Infof("Running with args: %s", strings.Join(args, " "))

	app.SetArgs(args)
	app.SetErr(io.Discard) // or use io.Discard if you want to discard error output: app.SetErr(io.Discard)
	app.SetOut(io.Discard)
	if err := app.ExecuteContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to scan project: %w", err)
	}

	content, err := os.ReadFile(tempFile)
	if err != nil {
		logger.Errorf("Error reading trivy output file: %s", err)
		return nil, err
	}

	rep, err := output.FromJSON(scanTarget, string(content))
	if err != nil {
		logger.Tracef("Error parsing trivy output, response from container: %s", string(content))
		progress.UpdateStatus(fmt.Sprintf("Error scanning image %s with %s", scanTarget, engineType))
		return nil, err
	}

	progress.UpdateStatus(fmt.Sprintf("Scanning %s...done", scanTarget))
	return rep, nil
}
