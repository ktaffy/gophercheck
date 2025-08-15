package analyzer

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"gophercheck/internal/config"
	"gophercheck/internal/models"

	"github.com/fatih/color"
)

// ReportGenerator handles formatting and displaying analysis results
type ReportGenerator struct {
	format string
	config *config.Config
}

// NewReportGenerator creates a new report generator
func NewReportGenerator(format string) *ReportGenerator {
	return &ReportGenerator{
		format: format,
		config: config.DefaultConfig(),
	}
}

func NewReportGeneratorWithConfig(cfg *config.Config) *ReportGenerator {
	return &ReportGenerator{
		format: cfg.Output.Format,
		config: cfg,
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

	// Check if colors should be used
	useColors := true
	verbose := false
	showSuggestions := true

	if r.config != nil {
		useColors = r.config.Output.Colors
		verbose = r.config.Output.Verbose
		showSuggestions = r.config.Output.ShowSuggestions
	}

	// Header
	if useColors {
		report.WriteString(color.CyanString("🔍 GopherCheck Analysis Report\n"))
		report.WriteString(color.WhiteString("═══════════════════════════════════════\n\n"))
	} else {
		report.WriteString("GopherCheck Analysis Report\n")
		report.WriteString("=======================================\n\n")
	}

	// Show configuration info if verbose
	if verbose && r.config != nil {
		r.writeConfigInfo(&report, useColors)
	}

	// Summary
	r.writeSummaryWithColors(&report, result, useColors)

	// Performance Score
	r.writePerformanceScore(&report, result)

	// Issues by severity
	if len(result.Issues) > 0 {
		r.writeIssuesSummaryWithColors(&report, result, useColors)

		if showSuggestions {
			report.WriteString("\n")
			r.writeDetailedIssuesWithColors(&report, result, useColors)
		}
	} else {
		if useColors {
			report.WriteString(color.GreenString("🎉 No performance issues detected! Great job!\n\n"))
		} else {
			report.WriteString("No performance issues detected! Great job!\n\n")
		}
	}

	// Footer
	if useColors {
		report.WriteString(color.WhiteString("Analysis completed in %s\n", result.AnalysisDuration))
	} else {
		report.WriteString(fmt.Sprintf("Analysis completed in %s\n", result.AnalysisDuration))
	}

	return report.String()
}

// writePerformanceScore writes the performance score with color coding
func (r *ReportGenerator) writePerformanceScore(report *strings.Builder, result *models.AnalysisResult) {
	score := result.PerformanceScore
	var scoreColor func(a ...interface{}) string
	var emoji string
	var excellent, good, fair int
	if r.config != nil {
		excellent = r.config.Analysis.ScoreThresholds.Excellent
		good = r.config.Analysis.ScoreThresholds.Good
		fair = r.config.Analysis.ScoreThresholds.Fair
	} else {
		excellent = 90
		good = 75
		fair = 50
	}

	switch {
	case score >= excellent:
		scoreColor = color.New(color.FgGreen).SprintFunc()
		emoji = "🌟"
	case score >= good:
		scoreColor = color.New(color.FgYellow).SprintFunc()
		emoji = "⚡"
	case score >= fair:
		scoreColor = color.New(color.FgHiYellow).SprintFunc()
		emoji = "⚠️"
	default:
		scoreColor = color.New(color.FgRed).SprintFunc()
		emoji = "🚨"
	}
	useColors := true
	if r.config != nil {
		useColors = r.config.Output.Colors
	}

	if useColors {
		scoreText := scoreColor(fmt.Sprintf("%d", score))
		report.WriteString(fmt.Sprintf("%s Performance Score: %s/100\n\n", emoji, scoreText))
	} else {
		report.WriteString(fmt.Sprintf("Performance Score: %d/100\n\n", score))
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

// CONFIG HELPERS
func (r *ReportGenerator) writeConfigInfo(report *strings.Builder, useColors bool) {
	if useColors {
		report.WriteString(color.WhiteString("📋 Configuration:\n"))
		report.WriteString(fmt.Sprintf("   Enabled categories: %s\n",
			color.CyanString(strings.Join(r.config.Analysis.EnabledCategories, ", "))))
		report.WriteString(fmt.Sprintf("   Score thresholds: %s\n",
			color.CyanString("%d/%d/%d",
				r.config.Analysis.ScoreThresholds.Excellent,
				r.config.Analysis.ScoreThresholds.Good,
				r.config.Analysis.ScoreThresholds.Fair)))
	} else {
		report.WriteString("Configuration:\n")
		report.WriteString(fmt.Sprintf("   Enabled categories: %s\n", strings.Join(r.config.Analysis.EnabledCategories, ", ")))
		report.WriteString(fmt.Sprintf("   Score thresholds: %d/%d/%d\n",
			r.config.Analysis.ScoreThresholds.Excellent,
			r.config.Analysis.ScoreThresholds.Good,
			r.config.Analysis.ScoreThresholds.Fair))
	}
	report.WriteString("\n")
}

func (r *ReportGenerator) writeSummaryWithColors(report *strings.Builder, result *models.AnalysisResult, useColors bool) {
	if useColors {
		report.WriteString(color.WhiteString("📊 Summary:\n"))
	} else {
		report.WriteString("Summary:\n")
	}
	report.WriteString(fmt.Sprintf("   Files analyzed: %d\n", len(result.Files)))
	report.WriteString(fmt.Sprintf("   Issues found: %d\n", result.TotalIssues))
	report.WriteString("\n")
}

func (r *ReportGenerator) writeIssuesSummaryWithColors(report *strings.Builder, result *models.AnalysisResult, useColors bool) {
	if useColors {
		report.WriteString(color.WhiteString("📋 Issues by Severity:\n"))
	} else {
		report.WriteString("Issues by Severity:\n")
	}

	severities := []string{"CRITICAL", "HIGH", "MEDIUM", "LOW"}
	for _, severity := range severities {
		count := result.IssuesBySeverity[severity]
		if count > 0 {
			if useColors {
				emoji, colorFunc := r.getSeverityDisplay(severity)
				countText := colorFunc(fmt.Sprintf("%d", count))
				report.WriteString(fmt.Sprintf("   %s %s: %s\n", emoji, severity, countText))
			} else {
				report.WriteString(fmt.Sprintf("   %s: %d\n", severity, count))
			}
		}
	}
}

func (r *ReportGenerator) writeDetailedIssuesWithColors(report *strings.Builder, result *models.AnalysisResult, useColors bool) {
	if useColors {
		report.WriteString(color.WhiteString("\n🔍 Detailed Issues:\n"))
	} else {
		report.WriteString("\nDetailed Issues:\n")
	}
	report.WriteString(strings.Repeat("─", 50) + "\n\n")

	// Sort issues by severity (critical first)
	sortedIssues := make([]models.Issue, len(result.Issues))
	copy(sortedIssues, result.Issues)

	sort.Slice(sortedIssues, func(i, j int) bool {
		return sortedIssues[i].Severity > sortedIssues[j].Severity
	})

	for i, issue := range sortedIssues {
		r.writeIssueDetailWithColors(report, issue, i+1, useColors)
		report.WriteString("\n")
	}
}

func (r *ReportGenerator) writeIssueDetailWithColors(report *strings.Builder, issue models.Issue, index int, useColors bool) {
	if useColors {
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
	} else {
		// Plain text version
		report.WriteString(fmt.Sprintf("Issue #%d - %s %s\n",
			index, issue.Severity.String(), strings.ToUpper(string(issue.Type))))

		report.WriteString(fmt.Sprintf("   Location: %s:%d:%d", issue.File, issue.Line, issue.Column))
		if issue.Function != "" {
			report.WriteString(fmt.Sprintf(" in function '%s'", issue.Function))
		}
		report.WriteString("\n")

		report.WriteString(fmt.Sprintf("   Issue: %s\n", issue.Message))

		if issue.Complexity != "" {
			report.WriteString(fmt.Sprintf("   Complexity: %s\n", issue.Complexity))
		}

		report.WriteString("   Suggestion:\n")
		suggestionLines := strings.Split(issue.Suggestion, "\n")
		for _, line := range suggestionLines {
			if strings.TrimSpace(line) != "" {
				report.WriteString(fmt.Sprintf("      %s\n", strings.TrimSpace(line)))
			}
		}
	}
}
