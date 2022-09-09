package output

import (
	"encoding/json"

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
	Vulnerabilities   []Vulnerability
	Misconfigurations []Misconfiguration
}

type DataSource struct {
	ID   string
	Name string
	URL  string
}

type Vulnerability struct {
	VulnerabilityID  string
	DataSource       *DataSource
	Title            string
	Description      string
	Severity         string
	SeveritySource   string
	PkgName          string
	PkgPath          string
	PrimaryURL       string
	InstalledVersion string
	FixedVersion     string
	References       []string
	CVSS             map[string]interface{}
}

type Code struct {
	Lines []string
}

type CauseMetadata struct {
	Resource string
	Provider string
	Service  string
	Code     Code
}

type Misconfiguration struct {
	Type          string
	ID            string
	Title         string
	Description   string
	Message       string
	Resolution    string
	Severity      string
	Status        string
	CauseMetadata CauseMetadata
	References    []string
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

	for _, result := range r.Results {
		for _, v := range result.Vulnerabilities {
			if _, ok := r.SeverityMap[v.Severity]; !ok {
				r.SeverityMap[v.Severity] = make([]*Result, 0)
			}
			sevMap := r.SeverityMap[v.Severity]

			var foundResult *Result
			var found bool
			for _, t := range sevMap {
				if result.Target == t.Target {
					foundResult = t
					found = true

					break
				}
			}
			if found {
				r.vulnerabilities++
				foundResult.Vulnerabilities = append(foundResult.Vulnerabilities, v)
			} else {
				foundResult = &Result{
					Target:          result.Target,
					Vulnerabilities: []Vulnerability{v},
				}
				sevMap = append(sevMap, foundResult)
			}

			r.SeverityMap[v.Severity] = sevMap
			r.SeverityCount[v.Severity]++
		}
		for _, m := range result.Misconfigurations {
			if _, ok := r.SeverityMap[m.Severity]; !ok {
				r.SeverityMap[m.Severity] = make([]*Result, 0)
			}
			sevMap := r.SeverityMap[m.Severity]

			var foundResult *Result
			var found bool
			for _, t := range sevMap {
				if result.Target == t.Target {
					foundResult = t
					found = true

					break
				}
			}
			if found {
				r.misconfigurations++
				foundResult.Misconfigurations = append(foundResult.Misconfigurations, m)
			} else {
				foundResult = &Result{
					Target:            result.Target,
					Misconfigurations: []Misconfiguration{m},
				}
				sevMap = append(sevMap, foundResult)
			}

			r.SeverityMap[m.Severity] = sevMap
			r.SeverityCount[m.Severity]++
		}
	}
}

func (r *Result) GetSeverityCounts() map[string]int {
	severities := make(map[string]int)
	for _, m := range r.Misconfigurations {
		severities[m.Severity]++
	}

	return severities
}

func (r *Result) GetMisconfigurationsForSeverity(severity string) []Misconfiguration {
	var misconfigs []Misconfiguration
	for _, m := range r.Misconfigurations {
		if m.Severity == severity || severity == "ALL" {
			misconfigs = append(misconfigs, m)
		}
	}

	return misconfigs
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
