package dockerClient

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/owenrumney/lazytrivy/pkg/logger"
	"github.com/owenrumney/lazytrivy/pkg/output"
)

func (c *Client) ScanAccount(ctx context.Context, accountNo, region string, progress Progress) (*output.Report, error) {
	return c.ScanService(ctx, "", accountNo, region, progress)
}

func (c *Client) ScanService(ctx context.Context, serviceName string, accountNo, region string, progress Progress) (*output.Report, error) {
	var env []string

	var updateCache bool
	target := accountNo
	additionalInfo := " maybe make a cuppa"
	if serviceName != "" {
		updateCache = true
		target = serviceName
		additionalInfo = ""
	}

	progress.UpdateStatus(fmt.Sprintf("Scanning %s...%s", target, additionalInfo))
	for _, envVar := range os.Environ() {
		if strings.HasPrefix(envVar, "AWS_") {
			env = append(env, envVar)
		}
	}

	command := []string{
		"aws", "--region", region, "-f=json",
	}

	if serviceName != "" {
		logger.Debugf("Scan will target service %s", serviceName)
		command = append(command, "--services", serviceName)
	}
	if updateCache {
		logger.Debugf("Cache will be updated for %s", serviceName)
		command = append(command, "--update-cache")
	}
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Debugf("Error getting user home dir: %s", err)
		userHomeDir = os.TempDir()
	}
	awsPath := filepath.Join(userHomeDir, ".aws")

	additionalBinds := []string{
		fmt.Sprintf("%s:/root/.aws", awsPath),
	}

	return c.scan(ctx, command, target, env, progress, "lazytrivy:1.0.0", additionalBinds...)
}
