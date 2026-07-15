package core

// DiagnosticStatus is the result severity of one doctor check.
type DiagnosticStatus string

const (
	// DiagnosticPass means the observed requirement is satisfied.
	DiagnosticPass DiagnosticStatus = "pass"
	// DiagnosticWarning means attention is useful but the doctor remains successful.
	DiagnosticWarning DiagnosticStatus = "warning"
	// DiagnosticFail means the project is not locally prepared.
	DiagnosticFail DiagnosticStatus = "fail"
)

// Diagnostic is one deterministic doctor observation.
type Diagnostic struct {
	Category string
	Name     string
	Status   DiagnosticStatus
	Message  string
}

// DiagnosticSummary counts doctor results by severity.
type DiagnosticSummary struct {
	Passed   int
	Warnings int
	Failed   int
}

// SummarizeDiagnostics counts checks without changing their order.
func SummarizeDiagnostics(checks []Diagnostic) DiagnosticSummary {
	var summary DiagnosticSummary
	for _, check := range checks {
		switch check.Status {
		case DiagnosticPass:
			summary.Passed++
		case DiagnosticWarning:
			summary.Warnings++
		case DiagnosticFail:
			summary.Failed++
		}
	}
	return summary
}
