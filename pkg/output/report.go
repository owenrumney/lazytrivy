package output

import (
	"encoding/json"
	"fmt"

	"github.com/owenrumney/lazytrivy/pkg/logger"
)

type Report struct {
	ImageName         string
	Results           []*Result
	SeverityMap       map[string][]*Result
	SeverityCount     map[string]int
	vulnerabilities   int
	misconfigurations int
	// K8s specific fields
	ClusterName string         `json:"ClusterName,omitempty"`
	Resources   []*K8sResource `json:"Resources,omitempty"`
}

type Result struct {
	Target            string
	Issues            []Issue
	Vulnerabilities   []Vulnerability
	Misconfigurations []Misconfiguration
	Secrets           []Secret
}

type K8sResource struct {
	Kind      string                   `json:"Kind"`
	Name      string                   `json:"Name"`
	Namespace string                   `json:"Namespace,omitempty"`
	Metadata  []map[string]interface{} `json:"Metadata"`
	Results   []*Result                `json:"Results"`
}

func FromJSON(imageName string, content string) (*Report, error) {
	logger.Debugf("Parsing JSON report")
	var report Report
	if err := json.Unmarshal([]byte(content), &report); err != nil {
		logger.Errorf("Failed to parse JSON report. %s", err)
		return nil, err
	}
	report.Process()
	report.ImageName = imageName

	return &report, nil
}

func FromK8sJSON(clusterName, content string) (*Report, error) {
	logger.Debugf("Parsing K8s JSON report")
	var k8sReport struct {
		ClusterName string         `json:"ClusterName"`
		Resources   []*K8sResource `json:"Resources"`
	}

	if err := json.Unmarshal([]byte(content), &k8sReport); err != nil {
		logger.Errorf("Failed to parse K8s JSON report. %s", err)
		return nil, err
	}

	// Extract namespace from metadata if not set and parse from target
	for _, resource := range k8sReport.Resources {
		if resource.Namespace == "" {
			// For cluster-scoped resources, keep namespace empty
			if resource.Kind == "ClusterRole" || resource.Kind == "ClusterRoleBinding" ||
				resource.Kind == "PersistentVolume" || resource.Kind == "Node" {
				resource.Namespace = "" // Keep empty for cluster-scoped
			}
			// Could add more sophisticated parsing here based on target format
		}
	}

	report := &Report{
		ClusterName: k8sReport.ClusterName,
		Resources:   k8sReport.Resources,
		Results:     []*Result{}, // Initialize empty
	}

	// Flatten Resources into Results for compatibility with existing UI components
	for _, resource := range k8sReport.Resources {
		for _, result := range resource.Results {
			report.Results = append(report.Results, result)
		}
	}

	report.Process()
	return report, nil
}

func (r *Report) Process() {
	r.SeverityMap = make(map[string][]*Result)
	r.SeverityCount = make(map[string]int)

	issueCount := 0

	for _, result := range r.Results {
		processIssues(r, result, result.Vulnerabilities, issueCount)
		processIssues(r, result, result.Misconfigurations, issueCount)
		processIssues(r, result, result.Secrets, issueCount)
	}

	_ = r
}

func processIssues[T Vulnerability | Misconfiguration | Secret](r *Report, result *Result, issues []T, count int) {
	for _, i := range issues {
		count++
		result.Issues = append(result.Issues, Issue(i))
		r.processIssue(result.Target, Issue(i))
	}
}

func (r *Report) processIssue(target string, m Issue) {

	if _, ok := r.SeverityMap[m.GetSeverity()]; !ok {
		r.SeverityMap[m.GetSeverity()] = make([]*Result, 0)
	}
	sevMap := r.SeverityMap[m.GetSeverity()]

	var foundResult *Result
	var found bool
	for _, t := range sevMap {
		if target == t.Target {
			foundResult = t
			found = true

			break
		}
	}
	if found {

		foundResult.Issues = append(foundResult.Issues, m)
	} else {
		foundResult = &Result{
			Target: target,
			Issues: []Issue{m},
		}
		sevMap = append(sevMap, foundResult)
	}
	r.misconfigurations++

	r.SeverityMap[m.GetSeverity()] = sevMap
	r.SeverityCount[m.GetSeverity()]++
}

func (r *Result) GetSeverityCounts() map[string]int {
	severities := make(map[string]int)
	for _, m := range r.Misconfigurations {
		severities[m.Severity]++
	}
	for _, v := range r.Vulnerabilities {
		severities[v.Severity]++
	}
	for _, s := range r.Secrets {
		severities[s.Severity]++
	}
	return severities
}

func (r *Result) GetIssuesForSeverity(severity string) []Issue {
	var issues []Issue
	for _, m := range r.Issues {
		if m.GetSeverity() == severity || severity == "ALL" {
			issues = append(issues, m)
		}
	}

	return issues
}

func (r *Result) HasIssues() bool {
	return len(r.Vulnerabilities) > 0 || len(r.Misconfigurations) > 0 || len(r.Secrets) > 0
}

func (r *Report) GetTotalVulnerabilities() int {
	return r.vulnerabilities
}

func (r *Report) GetTotalMisconfigurations() int {
	return r.misconfigurations
}

func (r *Report) HasIssues() bool {
	return r.GetTotalVulnerabilities() > 0 || r.GetTotalMisconfigurations() > 0
}

func (r *Report) GetResultForTarget(target string) (*Result, error) {
	for _, result := range r.Results {
		if result.Target == target {
			return result, nil
		}
	}
	return nil, fmt.Errorf("couldn't find any results for the target %s", target)

}
