package analyzer

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"gophercheck/internal/models"

	"github.com/fatih/color"
)

// ReportGenerator handles formatting and displaying analysis results
type ReportGenerator struct {
	format string
}

// NewReportGenerator creates a new report generator
func NewReportGenerator(format string) *ReportGenerator {
	return &ReportGenerator{
		format: format,
	}
}

// Generate creates a formatted report from analysis results
func (r *ReportGenerator) Generate(result *models.AnalysisResult) string {
	switch r.format {
	case "json":
		return r.generateJSON(result)
	default:
		return r.generateConsole(result)
	}
}

// generateJSON creates a JSON report
func (r *ReportGenerator) generateJSON(result *models.AnalysisResult) string {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error generating JSON report: %v", err)
	}
	return string(data)
}

// generateConsole creates a colorized console report
func (r *ReportGenerator) generateConsole(result *models.AnalysisResult) string {
	var report strings.Builder

	// Header
	report.WriteString(color.CyanString("🔍 GopherCheck Analysis Report\n"))
	report.WriteString(color.WhiteString("═══════════════════════════════════════\n\n"))

	// Summary
	r.writeSummary(&report, result)

	// Performance Score
	r.writePerformanceScore(&report, result)

	// Issues by severity
	if len(result.Issues) > 0 {
		r.writeIssuesSummary(&report, result)
		report.WriteString("\n")
		r.writeDetailedIssues(&report, result)
	} else {
		report.WriteString(color.GreenString("🎉 No performance issues detected! Great job!\n\n"))
	}

	// Footer
	report.WriteString(color.WhiteString("Analysis completed in %s\n", result.AnalysisDuration))

	return report.String()
}

// writeSummary writes the analysis summary
func (r *ReportGenerator) writeSummary(report *strings.Builder, result *models.AnalysisResult) {
	report.WriteString(color.WhiteString("📊 Summary:\n"))
	report.WriteString(fmt.Sprintf("   Files analyzed: %d\n", len(result.Files)))
	report.WriteString(fmt.Sprintf("   Issues found: %d\n", result.TotalIssues))
	report.WriteString("\n")
}

// writePerformanceScore writes the performance score with color coding
func (r *ReportGenerator) writePerformanceScore(report *strings.Builder, result *models.AnalysisResult) {
	score := result.PerformanceScore
	var scoreColor func(a ...interface{}) string
	var emoji string

	switch {
	case score >= 90:
		scoreColor = color.New(color.FgGreen).SprintFunc()
		emoji = "🌟"
	case score >= 75:
		scoreColor = color.New(color.FgYellow).SprintFunc()
		emoji = "⚡"
	case score >= 50:
		scoreColor = color.New(color.FgHiYellow).SprintFunc()
		emoji = "⚠️"
	default:
		scoreColor = color.New(color.FgRed).SprintFunc()
		emoji = "🚨"
	}

	// Fixed: Use proper string formatting
	scoreText := scoreColor(fmt.Sprintf("%d", score))
	report.WriteString(fmt.Sprintf("%s Performance Score: %s/100\n\n", emoji, scoreText))
}

// writeIssuesSummary writes the issues breakdown by severity
func (r *ReportGenerator) writeIssuesSummary(report *strings.Builder, result *models.AnalysisResult) {
	report.WriteString(color.WhiteString("📋 Issues by Severity:\n"))

	severities := []string{"CRITICAL", "HIGH", "MEDIUM", "LOW"}
	for _, severity := range severities {
		count := result.IssuesBySeverity[severity]
		if count > 0 {
			emoji, colorFunc := r.getSeverityDisplay(severity)
			// Fixed: Use proper string formatting
			countText := colorFunc(fmt.Sprintf("%d", count))
			report.WriteString(fmt.Sprintf("   %s %s: %s\n", emoji, severity, countText))
		}
	}
}

// writeDetailedIssues writes detailed information about each issue
func (r *ReportGenerator) writeDetailedIssues(report *strings.Builder, result *models.AnalysisResult) {
	report.WriteString(color.WhiteString("\n🔍 Detailed Issues:\n"))
	report.WriteString(strings.Repeat("─", 50) + "\n\n")

	// Sort issues by severity (critical first)
	sortedIssues := make([]models.Issue, len(result.Issues))
	copy(sortedIssues, result.Issues)

	sort.Slice(sortedIssues, func(i, j int) bool {
		return sortedIssues[i].Severity > sortedIssues[j].Severity
	})

	for i, issue := range sortedIssues {
		r.writeIssueDetail(report, issue, i+1)
		report.WriteString("\n")
	}
}

// writeIssueDetail writes a single issue's details
func (r *ReportGenerator) writeIssueDetail(report *strings.Builder, issue models.Issue, index int) {
	emoji, severityColor := r.getSeverityDisplay(issue.Severity.String())

	// Issue header
	report.WriteString(fmt.Sprintf("%s Issue #%d - %s %s\n",
		emoji, index, severityColor(issue.Severity.String()),
		color.WhiteString(strings.ToUpper(string(issue.Type)))))

	// Location
	report.WriteString(color.CyanString("   📍 Location: %s:%d:%d",
		issue.File, issue.Line, issue.Column))

	if issue.Function != "" {
		report.WriteString(color.CyanString(" in function '%s'", issue.Function))
	}
	report.WriteString("\n")

	// Message
	report.WriteString(color.WhiteString("   💭 Issue: %s\n", issue.Message))

	// Complexity (if available)
	if issue.Complexity != "" {
		report.WriteString(color.YellowString("   📊 Complexity: %s\n", issue.Complexity))
	}

	// Suggestion
	report.WriteString(color.GreenString("   💡 Suggestion:\n"))
	suggestionLines := strings.Split(issue.Suggestion, "\n")
	for _, line := range suggestionLines {
		if strings.TrimSpace(line) != "" {
			report.WriteString(color.GreenString("      %s\n", strings.TrimSpace(line)))
		}
	}
}

// getSeverityDisplay returns emoji and color function for a severity level
func (r *ReportGenerator) getSeverityDisplay(severity string) (string, func(a ...interface{}) string) {
	switch severity {
	case "CRITICAL":
		return "🚨", color.New(color.FgRed, color.Bold).SprintFunc()
	case "HIGH":
		return "❌", color.New(color.FgRed).SprintFunc()
	case "MEDIUM":
		return "⚠️", color.New(color.FgYellow).SprintFunc()
	case "LOW":
		return "ℹ️", color.New(color.FgBlue).SprintFunc()
	default:
		return "❓", color.New(color.FgWhite).SprintFunc()
	}
}
