package output

import "encoding/json"

type Report struct {
	ImageName     string
	Results       []*Result
	SeverityMap   map[string][]*Result
	SeverityCount map[string]int
}

type Result struct {
	Target          string
	Vulnerabilities []Vulnerability
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

func FromJSON(imageName string, content string) (*Report, error) {
	var report Report
	if err := json.Unmarshal([]byte(content), &report); err != nil {
		return nil, err
	}
	report.processReport()
	report.ImageName = imageName

	return &report, nil
}

func (r *Report) processReport() {
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
	}
}
