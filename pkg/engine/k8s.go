package engine

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/aquasecurity/trivy/pkg/commands"
	"github.com/owenrumney/lazytrivy/pkg/config"
	"github.com/owenrumney/lazytrivy/pkg/logger"
	"github.com/owenrumney/lazytrivy/pkg/output"
)

func (c *Client) ScanKubernetes(ctx context.Context, kubeContext string, cfg *config.Config, progress Progress) (*output.Report, error) {
	progress.UpdateStatus(fmt.Sprintf("Scanning Kubernetes cluster with context: %s", kubeContext))

	command := []string{"k8s", "--quiet"}
	command = c.updateCommand(command, cfg)

	// Add context if specified
	if kubeContext != "" {
		command = append(command, kubeContext)
	}

	// Scan the entire cluster
	command = append(command, "--format", "json", "cluster")

	tempFile := filepath.Join(os.TempDir(), "lazytrivy-k8s-output.json")
	defer func() { _ = os.Remove(tempFile) }()
	command = append(command, "--output", tempFile)

	app := commands.NewApp()
	app.SetArgs(command)

	logger.Infof("Running K8s scan with args: %s", strings.Join(command, " "))

	if err := app.ExecuteContext(ctx); err != nil {
		progress.UpdateStatus(fmt.Sprintf("Error scanning K8s cluster: %v", err))
		return nil, fmt.Errorf("failed to scan K8s cluster: %w", err)
	}

	content, err := os.ReadFile(tempFile)
	if err != nil {
		logger.Errorf("Error reading trivy K8s output file: %s", err)
		return nil, err
	}

	// Use the new K8s-specific parser
	rep, err := output.FromK8sJSON(fmt.Sprintf("cluster-%s", kubeContext), string(content))
	if err != nil {
		logger.Tracef("Error parsing trivy K8s output, response: %s", string(content))
		progress.UpdateStatus(fmt.Sprintf("Error parsing K8s scan results for context %s", kubeContext))
		return nil, err
	}

	progress.UpdateStatus(fmt.Sprintf("Scanning cluster %s...done", kubeContext))
	return rep, nil
}

// GetKubernetesContexts returns a list of available Kubernetes contexts
func (c *Client) GetKubernetesContexts() ([]string, error) {
	// Use kubectl to get available contexts with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "kubectl", "config", "get-contexts", "-o", "name")
	output, err := cmd.Output()
	if err != nil {
		logger.Debugf("Error getting kubectl contexts: %v", err)
		return []string{}, fmt.Errorf("failed to get kubectl contexts: %w", err)
	}

	contexts := strings.Split(strings.TrimSpace(string(output)), "\n")
	var validContexts []string
	for _, ctx := range contexts {
		ctx = strings.TrimSpace(ctx)
		if ctx != "" {
			validContexts = append(validContexts, ctx)
		}
	}

	return validContexts, nil
}

// GetCurrentKubernetesContext returns the currently active Kubernetes context
func (c *Client) GetCurrentKubernetesContext() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "kubectl", "config", "current-context")
	output, err := cmd.Output()
	if err != nil {
		logger.Debugf("Error getting current kubectl context: %v", err)
		return "", fmt.Errorf("failed to get current kubectl context: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}
