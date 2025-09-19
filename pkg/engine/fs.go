package engine

import (
	"context"
	"fmt"
	"strings"

	"github.com/owenrumney/lazytrivy/pkg/config"
	"github.com/owenrumney/lazytrivy/pkg/output"
)

func (c *Client) ScanFilesystem(ctx context.Context, path string, cfg *config.Config, progress Progress) (*output.Report, error) {
	progress.UpdateStatus(fmt.Sprintf("Scanning filesystem %s...", path))
	command := []string{"fs", "--quiet"}
	command = c.updateCommand(command, cfg)
	command = append(command, "-f=json", path)

	return c.scan(ctx, command, path, []string{}, progress, "lazytrivy:1.0.0", EngineDocker, path)
}

func (c *Client) updateCommand(command []string, cfg *config.Config) []string {
	if cfg.Insecure {
		command = append(command, "--insecure")
	}
	var scanChecks []string
	if cfg.Scanner.ScanVulnerabilities {
		scanChecks = append(scanChecks, "vuln")
	}
	if cfg.Scanner.ScanMisconfiguration {
		scanChecks = append(scanChecks, "misconfig")
	}
	if cfg.Scanner.ScanSecrets {
		scanChecks = append(scanChecks, "secret")
	}

	if len(scanChecks) > 0 {
		command = append(command, "--scanners", strings.Join(scanChecks, ","))
	}

	return command
}
