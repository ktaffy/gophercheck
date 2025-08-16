package analyzer

import (
	"encoding/json"
	"fmt"
	"path/filepath"
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

func (r *ReportGenerator) generateConsole(result *models.AnalysisResult) string {
	useVerbose := false
	if r.config != nil {
		useVerbose = r.config.Output.Verbose
	}
	if useVerbose {
		return r.generateVerboseConsole(result)
	} else {
		return r.generateMinimalConsole(result)
	}
}

func (r *ReportGenerator) generateMinimalConsole(result *models.AnalysisResult) string {
	var report strings.Builder

	useColors := true
	if r.config != nil {
		useColors = r.config.Output.Colors
	}

	// Header
	if useColors {
		report.WriteString(color.CyanString("🔍 GopherCheck Analysis (%d files analyzed)\n", len(result.Files)))
	} else {
		report.WriteString(fmt.Sprintf("GopherCheck Analysis (%d files analyzed)\n", len(result.Files)))
	}

	// Performance Score
	r.writePerformanceScore(&report, result)

	// Issues Summary
	r.writeIssuesSummary(&report, result, useColors)

	// Show only CRITICAL and HIGH issues
	highPriorityIssues := r.filterHighPriorityIssues(result.Issues)
	if len(highPriorityIssues) > 0 {
		r.writeHighPriorityIssues(&report, highPriorityIssues, useColors)
	}

	// Footer
	if useColors {
		report.WriteString(color.WhiteString("\n📊 Completed in %s\n\n", result.AnalysisDuration))
		report.WriteString(color.WhiteString("💡 Run with --verbose for details and suggestions\n"))
	} else {
		report.WriteString(fmt.Sprintf("\nCompleted in %s\n\n", result.AnalysisDuration))
		report.WriteString("Run with --verbose for details and suggestions\n")
	}

	return report.String()
}

func (r *ReportGenerator) generateVerboseConsole(result *models.AnalysisResult) string {
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

	sortedIssues := make([]models.Issue, len(result.Issues))
	copy(sortedIssues, result.Issues)

	sort.Slice(sortedIssues, func(i, j int) bool {
		return sortedIssues[i].Severity > sortedIssues[j].Severity
	})

	for i, issue := range sortedIssues {
		r.writeIssueCard(report, issue, i+1, useColors)
		report.WriteString("\n")
	}
}

func (r *ReportGenerator) writeIssueCard(report *strings.Builder, issue models.Issue, index int, useColors bool) {
	severity := issue.Severity.String()
	issueTypeUpper := strings.ToUpper(string(issue.Type))
	cardWidth := 50 // Increased width for better formatting

	if useColors {
		emoji, severityColor := r.getSeverityDisplay(severity)

		// Card header
		headerText := fmt.Sprintf(" %s ", severityColor(severity))
		paddingLen := cardWidth - len(severity) - 4
		if paddingLen < 0 {
			paddingLen = 0
		}
		report.WriteString(fmt.Sprintf("┌─%s%s┐\n", headerText, strings.Repeat("─", paddingLen)))

		// Issue type and number
		issueText := fmt.Sprintf(" %s Issue #%d - %s", emoji, index, issueTypeUpper)
		r.writeCardLine(report, issueText, cardWidth)

		// Location (truncated if too long)
		fileName := filepath.Base(issue.File) // Just filename, not full path
		locationText := fmt.Sprintf(" 📍 %s:%d:%d", fileName, issue.Line, issue.Column)
		if issue.Function != "" {
			funcName := issue.Function
			if len(funcName) > 20 {
				funcName = funcName[:17] + "..."
			}
			locationText += fmt.Sprintf(" in %s()", funcName)
		}
		r.writeCardLine(report, locationText, cardWidth)

		// Complexity
		if issue.Complexity != "" {
			complexityText := fmt.Sprintf(" 📊 %s", issue.Complexity)
			r.writeCardLine(report, complexityText, cardWidth)
		}

		// Brief message (truncated)
		messageText := fmt.Sprintf(" 💭 %s", r.truncateMessage(issue.Message, cardWidth-6))
		r.writeCardLine(report, messageText, cardWidth)

		// Empty line separator
		r.writeCardLine(report, "", cardWidth)

		// Suggestion header
		r.writeCardLine(report, " 💡 Suggestion:", cardWidth)

		// Suggestion content (properly wrapped)
		suggestionLines := r.wrapSuggestion(issue.Suggestion, cardWidth-4)
		for _, line := range suggestionLines {
			if strings.TrimSpace(line) != "" {
				r.writeCardLine(report, " "+line, cardWidth)
			} else {
				r.writeCardLine(report, "", cardWidth)
			}
		}

		// Card footer
		report.WriteString("└" + strings.Repeat("─", cardWidth-2) + "┘\n")

	} else {
		// Plain text version (unchanged but cleaner)
		report.WriteString(fmt.Sprintf("Issue #%d - %s %s\n", index, severity, issueTypeUpper))
		report.WriteString(fmt.Sprintf("Location: %s:%d:%d", filepath.Base(issue.File), issue.Line, issue.Column))
		if issue.Function != "" {
			report.WriteString(fmt.Sprintf(" in %s()", issue.Function))
		}
		report.WriteString("\n")

		if issue.Complexity != "" {
			report.WriteString(fmt.Sprintf("Complexity: %s\n", issue.Complexity))
		}

		report.WriteString(fmt.Sprintf("Issue: %s\n", issue.Message))
		report.WriteString("Suggestion:\n")
		suggestionLines := strings.Split(issue.Suggestion, "\n")
		for _, line := range suggestionLines {
			if strings.TrimSpace(line) != "" {
				report.WriteString(fmt.Sprintf("  %s\n", strings.TrimSpace(line)))
			}
		}
		report.WriteString(strings.Repeat("-", 50) + "\n")
	}
}

func (r *ReportGenerator) truncateMessage(message string, maxLen int) string {
	if len(message) <= maxLen {
		return message
	}
	return message[:maxLen-3] + "..."
}

func (r *ReportGenerator) wrapSuggestion(suggestion string, maxLen int) []string {
	lines := strings.Split(suggestion, "\n")
	var wrapped []string

	for _, line := range lines {
		if len(line) <= maxLen {
			wrapped = append(wrapped, line)
		} else {
			words := strings.Fields(line)
			currentLine := ""
			for _, word := range words {
				if len(currentLine+" "+word) <= maxLen {
					if currentLine == "" {
						currentLine = word
					} else {
						currentLine += " " + word
					}
				} else {
					if currentLine != "" {
						wrapped = append(wrapped, currentLine)
					}
					currentLine = word
				}
			}
			if currentLine != "" {
				wrapped = append(wrapped, currentLine)
			}
		}
	}

	return wrapped
}

func (r *ReportGenerator) writeIssuesSummary(report *strings.Builder, result *models.AnalysisResult, useColors bool) {
	if useColors {
		report.WriteString(color.WhiteString("\nIssues Summary:\n"))
	} else {
		report.WriteString("\nIssues Summary:\n")
	}

	// Count issues by severity
	critical := result.IssuesBySeverity["CRITICAL"]
	high := result.IssuesBySeverity["HIGH"]
	medium := result.IssuesBySeverity["MEDIUM"]
	low := result.IssuesBySeverity["LOW"]

	if useColors {
		report.WriteString(fmt.Sprintf("  🚨 %d CRITICAL   ❌ %d HIGH   ⚠️ %d MEDIUM   ℹ️ %d LOW\n",
			critical, high, medium, low))
	} else {
		report.WriteString(fmt.Sprintf("  %d CRITICAL   %d HIGH   %d MEDIUM   %d LOW\n",
			critical, high, medium, low))
	}
}

func (r *ReportGenerator) filterHighPriorityIssues(issues []models.Issue) []models.Issue {
	var highPriority []models.Issue
	for _, issue := range issues {
		if issue.Severity == models.SeverityCritical || issue.Severity == models.SeverityHigh {
			highPriority = append(highPriority, issue)
		}
	}
	return highPriority
}

func (r *ReportGenerator) writeHighPriorityIssues(report *strings.Builder, issues []models.Issue, useColors bool) {
	if useColors {
		report.WriteString(color.WhiteString("\nCritical & High Priority:\n"))
	} else {
		report.WriteString("\nCritical & High Priority:\n")
	}

	sortedIssues := make([]models.Issue, len(issues))
	copy(sortedIssues, issues)
	sort.Slice(sortedIssues, func(i, j int) bool {
		return sortedIssues[i].Severity > sortedIssues[j].Severity
	})

	for _, issue := range sortedIssues {
		severity := issue.Severity.String()
		issueType := strings.ReplaceAll(string(issue.Type), "_", " ")
		issueType = strings.ToUpper(issueType)

		emoji, colorFunc := r.getSeverityDisplay(severity)

		fileName := filepath.Base(issue.File)
		description := r.getShortDescription(issue)

		locationCol := fmt.Sprintf("%s:%d", fileName, issue.Line)

		if useColors {
			severityCol := fmt.Sprintf("%s %s", emoji, colorFunc(severity))
			report.WriteString(fmt.Sprintf("  %-20s %-18s %-25s %s\n",
				locationCol, severityCol, issueType, description))
		} else {
			severityCol := severity
			report.WriteString(fmt.Sprintf("  %-20s %-12s %-25s %s\n",
				locationCol, severityCol, issueType, description))
		}
	}
}

func (r *ReportGenerator) getShortDescription(issue models.Issue) string {
	funcName := issue.Function
	if len(funcName) > 20 {
		funcName = funcName[:17] + "..."
	}

	switch issue.Type {
	case models.IssueFunctionLength:
		return fmt.Sprintf("%s() (%s)", funcName, issue.Complexity)
	case models.IssueCyclomaticComplex:
		return fmt.Sprintf("%s() (%s)", funcName, issue.Complexity)
	case models.IssueNestedLoops:
		return fmt.Sprintf("%s() (%s)", funcName, issue.Complexity)
	case models.IssueMemoryAlloc:
		return fmt.Sprintf("%s() (%s)", funcName, issue.Complexity)
	case models.IssueSliceGrowth:
		return fmt.Sprintf("%s() (%s)", funcName, issue.Complexity)
	case models.IssueInefficinetDS:
		return fmt.Sprintf("%s() (%s)", funcName, issue.Complexity)
	case models.IssueStringConcat:
		return fmt.Sprintf("%s() (%s)", funcName, issue.Complexity)
	case models.IssueImportCycle:
		return issue.Complexity // For import cycles, complexity field contains cycle info
	default:
		return fmt.Sprintf("%s()", funcName)
	}
}

func (r *ReportGenerator) truncateToDisplayWidth(text string, maxWidth int) string {
	if r.calculateDisplayWidth(text) <= maxWidth {
		return text
	}

	// Simple truncation for now
	runes := []rune(text)
	for i := len(runes) - 1; i >= 0; i-- {
		candidate := string(runes[:i])
		if r.calculateDisplayWidth(candidate) <= maxWidth {
			return candidate
		}
	}
	return ""
}

func (r *ReportGenerator) calculateDisplayWidth(text string) int {
	// Simple approximation: count emojis as 2 display characters
	emojiCount := 0
	for _, char := range text {
		if char > 127 { // Non-ASCII, likely emoji
			emojiCount++
		}
	}
	// Rough approximation: each emoji takes about 2 display characters but 4+ string characters
	return len(text) - emojiCount*2
}

func (r *ReportGenerator) writeCardLine(report *strings.Builder, text string, cardWidth int) {
	// Calculate actual display width (emojis count as 2 characters in display but 4+ in string length)
	displayWidth := r.calculateDisplayWidth(text)
	paddingNeeded := cardWidth - displayWidth - 2 // -2 for the │ characters

	if paddingNeeded < 0 {
		// Truncate if too long
		text = r.truncateToDisplayWidth(text, cardWidth-5) + "..."
		paddingNeeded = 0
	}

	report.WriteString(fmt.Sprintf("│%s%s│\n", text, strings.Repeat(" ", paddingNeeded)))
}
