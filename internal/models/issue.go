package models

import "go/token"

type Severity int

const (
	SeverityLow Severity = iota
	SeverityMedium
	SeverityHigh
	SeverityCritical
)

func (s Severity) String() string {
	switch s {
	case SeverityLow:
		return "LOW"
	case SeverityMedium:
		return "MEDIUM"
	case SeverityHigh:
		return "HIGH"
	case SeverityCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

type IssueType string

const (
	IssueNestedLoops       IssueType = "nested_loops"
	IssueStringConcat      IssueType = "string_concatenation"
	IssueInefficinetDS     IssueType = "inefficient_data_structure"
	IssueCyclomaticComplex IssueType = "cyclomatic_complexity"
	IssueMemoryAlloc       IssueType = "memory_allocation"
	IssueSliceGrowth       IssueType = "slice_growth"    // New: Slice growth patterns
	IssueFunctionLength    IssueType = "function_length" // New: Function length analysis
	IssueImportCycle       IssueType = "import_cycle"    // New: Import cycle detection
)

type Issue struct {
	Type        IssueType `json:"type"`
	Severity    Severity  `json:"severity"`
	File        string    `json:"file"`
	Line        int       `json:"line"`
	Column      int       `json:"column"`
	Function    string    `json:"function,omitempty"`
	Message     string    `json:"message"`
	Suggestion  string    `json:"suggestion"`
	Complexity  string    `json:"complexity,omitempty"` // e.g., "O(n²)", "O(n)"
	CodeSnippet string    `json:"code_snippet,omitempty"`
}

func (i *Issue) Position() token.Pos {
	return token.Pos(i.Line<<16 | i.Column)
}

type AnalysisResult struct {
	Files            []string       `json:"files_analyzed"`
	TotalIssues      int            `json:"total_issues"`
	IssuesBySeverity map[string]int `json:"issues_by_severity"`
	Issues           []Issue        `json:"issues"`
	PerformanceScore int            `json:"performance_score"` // 0-100 scale
	AnalysisDuration string         `json:"analysis_duration"`
}

func NewAnalysisResult() *AnalysisResult {
	return &AnalysisResult{
		Files:            make([]string, 0),
		Issues:           make([]Issue, 0),
		IssuesBySeverity: make(map[string]int),
	}
}

func (ar *AnalysisResult) AddIssue(issue Issue) {
	ar.Issues = append(ar.Issues, issue)
	ar.TotalIssues++
	ar.IssuesBySeverity[issue.Severity.String()]++
}

func (ar *AnalysisResult) CalculateScore() {
	if ar.TotalIssues == 0 {
		ar.PerformanceScore = 100
		return
	}

	// Enhanced scoring algorithm with new issue types
	penalty := 0
	for _, issue := range ar.Issues {
		basePenalty := 0
		switch issue.Severity {
		case SeverityLow:
			basePenalty = 5
		case SeverityMedium:
			basePenalty = 15
		case SeverityHigh:
			basePenalty = 30
		case SeverityCritical:
			basePenalty = 50
		}

		// Apply multipliers for certain issue types
		switch issue.Type {
		case IssueCyclomaticComplex, IssueFunctionLength:
			basePenalty = int(float64(basePenalty) * 1.2) // 20% more penalty for maintainability issues
		case IssueNestedLoops, IssueMemoryAlloc:
			basePenalty = int(float64(basePenalty) * 1.5) // 50% more penalty for performance issues
		case IssueImportCycle:
			basePenalty = int(float64(basePenalty) * 1.8) // 80% more penalty for architecture issues
		}

		penalty += basePenalty
	}

	score := max(100-penalty, 0)
	ar.PerformanceScore = score
}
