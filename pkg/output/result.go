package output

import "encoding/json"

type Report struct {
	Results []Result
}

type Result struct {
	Target          string
	Vulnerabilities []Vulnerability
}

type Vulnerability struct {
	VulnerabilityID string
	Title           string
	Description     string
	Severity        string
	PkgName         string
	PrimaryURL      string
}

func FromJson(content string) (Report, error) {
	var report Report
	err := json.Unmarshal([]byte(content), &report)
	return report, err
}
