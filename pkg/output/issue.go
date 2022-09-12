package output

type IssueType int

const (
	IssueTypeMisconfiguration IssueType = iota
	IssueTypeVulnerability
	IssueTypeSecret
	IssueTypeLicense
)

// Issue is the interface that makes the shared SummaryWidget possible
type Issue interface {
	// GetType returns the IssueType
	GetType() IssueType

	// GetID returns the ID of the issue.
	GetID() string

	// GetTitle returns the title of the issue.
	GetTitle() string

	// GetDescription returns the description of the issue.
	GetDescription() string

	// GetDatasourceName returns the name of the datasource.
	GetDatasourceName() string

	// GetSeverity returns the severity of the issue.
	GetSeverity() string

	// GetSeveritySource returns the severity source of the issue.
	GetSeveritySource() string

	// GetPackageName returns the package name
	GetPackageName() string

	// GetPackagePath returns the package path
	GetPackagePath() string

	// GetPrimaryURL returns the primary URL
	GetPrimaryURL() string

	// GetInstalledVersion returns the installed version
	GetInstalledVersion() string

	// GetFixedVersion returns the fixed version
	GetFixedVersion() string

	// GetReferences returns the references
	GetReferences() []string

	// GetCVSS returns the CVSS
	GetCVSS() map[string]interface{}

	// GetMisconfigurationType returns the misconfiguration type
	GetMisconfigurationType() string

	// GetMessage returns the message
	GetMessage() string

	// GetResolution returns the resolution
	GetResolution() string

	// GetStatus returns the status
	GetStatus() string

	// GetCauseMetadata returns the cause metadata
	GetCauseMetadata() CauseMetadata

	// GetMatch returns the match
	GetMatch() string

	// GetDeleted returns the deleted
	GetDeleted() string
}
