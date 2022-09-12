package output

type Secret struct {
	RuleID   string
	Category string
	Severity string
	Title    string
	Match    string
	Deleted  bool
}

func (s Secret) GetType() IssueType {
	return IssueTypeSecret
}

func (s Secret) GetID() string {
	return s.RuleID
}

func (s Secret) GetTitle() string {
	return s.Title
}

func (s Secret) GetDescription() string {
	return s.Category
}

func (s Secret) GetDatasourceName() string {
	return ""
}

func (s Secret) GetSeverity() string {
	return s.Severity
}

func (s Secret) GetSeveritySource() string {
	return ""
}

func (s Secret) GetPackageName() string {
	return ""
}

func (s Secret) GetPackagePath() string {
	return ""
}

func (s Secret) GetPrimaryURL() string {
	return ""
}

func (s Secret) GetInstalledVersion() string {
	return ""
}

func (s Secret) GetFixedVersion() string {
	return ""
}

func (s Secret) GetReferences() []string {
	return nil
}

func (s Secret) GetCVSS() map[string]interface{} {
	return nil
}

func (s Secret) GetMisconfigurationType() string {
	return ""
}

func (s Secret) GetMessage() string {
	return ""
}

func (s Secret) GetResolution() string {
	return ""
}

func (s Secret) GetStatus() string {
	return ""
}

func (s Secret) GetCauseMetadata() CauseMetadata {
	return CauseMetadata{}
}

func (s Secret) GetMatch() string {
	return s.Match
}

func (s Secret) GetDeleted() string {
	if s.Deleted {
		return "Yes"
	}
	return "No"
}
