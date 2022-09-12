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
}

type Result struct {
	Target            string
	Issues            []Issue
	Vulnerabilities   []Vulnerability
	Misconfigurations []Misconfiguration
	Secrets           []Secret
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

func (r *Report) Process() {
	r.SeverityMap = make(map[string][]*Result)
	r.SeverityCount = make(map[string]int)

	issueCount := 0

	for _, result := range r.Results {
		r.processVulnerability(result, result.Vulnerabilities, issueCount)
		r.processMisconfiguration(result, result.Misconfigurations, issueCount)
		r.processSecrets(result, result.Secrets, issueCount)
	}

	_ = r
}

func (r *Report) processVulnerability(result *Result, issues []Vulnerability, count int) {
	for _, m := range issues {
		count++
		result.Issues = append(result.Issues, m)
		r.processIssue(result.Target, m)
	}
}

func (r *Report) processMisconfiguration(result *Result, issues []Misconfiguration, count int) {
	for _, m := range issues {
		result.Issues = append(result.Issues, m)
		count++
		r.processIssue(result.Target, m)
	}
}

func (r *Report) processSecrets(result *Result, issues []Secret, count int) {
	for _, m := range issues {
		count++
		result.Issues = append(result.Issues, m)
		r.processIssue(result.Target, m)
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
