package checks

type Severity string

const (
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

type Finding struct {
	CheckID  string         `json:"checkId"`
	Severity Severity       `json:"severity"`
	Title    string         `json:"title"`
	Detail   map[string]any `json:"detail"`
}
