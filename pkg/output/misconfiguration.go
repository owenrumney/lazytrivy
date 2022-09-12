package output

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

func (m Misconfiguration) GetType() IssueType {
	return IssueTypeMisconfiguration
}

func (m Misconfiguration) GetID() string {
	return m.ID
}

func (m Misconfiguration) GetTitle() string {
	return m.Title
}

func (m Misconfiguration) GetDescription() string {
	return m.Description
}

func (m Misconfiguration) GetDatasourceName() string {
	return ""
}

func (m Misconfiguration) GetSeverity() string {
	return m.Severity
}

func (m Misconfiguration) GetSeveritySource() string {
	return ""
}

func (m Misconfiguration) GetPackageName() string {
	return ""
}

func (m Misconfiguration) GetPackagePath() string {
	return ""
}

func (m Misconfiguration) GetPrimaryURL() string {
	if m.References != nil && len(m.References) > 0 {
		return m.References[0]
	}
	return ""
}

func (m Misconfiguration) GetInstalledVersion() string {
	return ""
}

func (m Misconfiguration) GetFixedVersion() string {
	return ""
}

func (m Misconfiguration) GetReferences() []string {
	return nil
}

func (m Misconfiguration) GetCVSS() map[string]interface{} {
	return nil
}

func (m Misconfiguration) GetMisconfigurationType() string {
	return m.Type
}

func (m Misconfiguration) GetMessage() string {
	return m.Message
}

func (m Misconfiguration) GetResolution() string {
	return m.Resolution
}

func (m Misconfiguration) GetStatus() string {
	return m.Status
}

func (m Misconfiguration) GetCauseMetadata() CauseMetadata {
	return m.CauseMetadata
}

func (m Misconfiguration) GetMatch() string {
	return ""
}

func (m Misconfiguration) GetDeleted() string {
	return "No"
}

type Code struct {
	Lines []Line
}

type Line struct {
	Number     int
	Content    string
	IsCause    bool
	Annotation string
	Truncated  bool
	FirstCause bool
	LastCause  bool
}

type CauseMetadata struct {
	Resource string
	Provider string
	Service  string
	Code     Code
}
