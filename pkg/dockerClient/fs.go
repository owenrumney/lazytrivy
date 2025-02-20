package dockerClient

import (
	"context"
	"fmt"
	"strings"

	"github.com/owenrumney/lazytrivy/pkg/output"
)

func (c *Client) ScanFilesystem(ctx context.Context, path string, requiredChecks []string, insecure bool, progress Progress) (*output.Report, error) {
	checks := strings.Join(requiredChecks, ",")
	progress.UpdateStatus(fmt.Sprintf("Scanning filesystem %s...", path))
	command := []string{"fs", "--quiet"}
	if insecure {
		command = append(command, "--insecure")
	}
	command = append(command, "--security-checks", checks, "-f=json", "/target")

	return c.scan(ctx, command, path, []string{}, progress, "lazytrivy:1.0.0", EngineDocker, fmt.Sprintf("%s:/target", path))
}
