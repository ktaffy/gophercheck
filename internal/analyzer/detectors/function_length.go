package detectors

import (
	"fmt"
	"go/ast"
	"go/token"
	"gophercheck/internal/config"
	"gophercheck/internal/models"
)

// FunctionLengthDetector finds overly long functions that should be refactored
type FunctionLengthDetector struct {
	config *config.Config
}

func NewFunctionLengthDetector() *FunctionLengthDetector {
	return &FunctionLengthDetector{}
}

func NewFunctionLengthDetectorWithConfig(cfg *config.Config) *FunctionLengthDetector {
	return &FunctionLengthDetector{
		config: cfg,
	}
}

func (d *FunctionLengthDetector) SetConfig(cfg *config.Config) {
	d.config = cfg
}

func (d *FunctionLengthDetector) Name() string {
	return "Function Length Detector"
}

func (d *FunctionLengthDetector) Detect(file *ast.File, fset *token.FileSet, filename string) []models.Issue {
	detector := &functionLengthVisitor{
		fset:     fset,
		filename: filename,
		issues:   make([]models.Issue, 0),
		detector: d,
	}

	ast.Walk(detector, file)
	return detector.issues
}

type functionLengthVisitor struct {
	fset     *token.FileSet
	filename string
	issues   []models.Issue
	detector *FunctionLengthDetector
}

const (
	// Thresholds for function length (lines of code)
	MediumThreshold   = 50  // Medium complexity warning
	HighThreshold     = 100 // High complexity warning
	CriticalThreshold = 200 // Critical - definitely too long
)

func (v *functionLengthVisitor) Visit(node ast.Node) ast.Visitor {
	if fn, ok := node.(*ast.FuncDecl); ok && fn.Body != nil {
		v.analyzeFunctionLength(fn)
	}
	return v
}

func (v *functionLengthVisitor) analyzeFunctionLength(fn *ast.FuncDecl) {
	startPos := v.fset.Position(fn.Pos())
	endPos := v.fset.Position(fn.End())

	totalLines := endPos.Line - startPos.Line + 1

	// Count actual lines of code (excluding braces, empty lines, etc.)
	actualLOC := v.countActualLinesOfCode(fn.Body)

	funcName := v.getFunctionName(fn)

	mediumThreshold := 50
	if v.detector.config != nil && v.detector.config.Rules.Complexity.FunctionLength.Enabled {
		mediumThreshold = v.detector.config.Rules.Complexity.FunctionLength.MediumThreshold
	}
	if actualLOC >= mediumThreshold {
		severity := v.calculateSeverity(actualLOC)
		v.createLengthIssue(fn, funcName, actualLOC, totalLines, severity)
	}
}

func (v *functionLengthVisitor) getFunctionName(fn *ast.FuncDecl) string {
	if fn.Name != nil {
		return fn.Name.Name
	}
	return "anonymous"
}

func (v *functionLengthVisitor) countActualLinesOfCode(body *ast.BlockStmt) int {
	linesSeen := make(map[int]bool)

	ast.Inspect(body, func(n ast.Node) bool {
		if n != nil {
			pos := v.fset.Position(n.Pos())
			linesSeen[pos.Line] = true
		}
		return true
	})

	// Return the count of unique lines
	return len(linesSeen)
}

func (v *functionLengthVisitor) calculateSeverity(loc int) models.Severity {
	mediumThreshold := 50
	highThreshold := 100
	criticalThreshold := 200

	if v.detector.config != nil && v.detector.config.Rules.Complexity.FunctionLength.Enabled {
		mediumThreshold = v.detector.config.Rules.Complexity.FunctionLength.MediumThreshold
		highThreshold = v.detector.config.Rules.Complexity.FunctionLength.HighThreshold
		criticalThreshold = v.detector.config.Rules.Complexity.FunctionLength.CriticalThreshold
	}

	switch {
	case loc >= criticalThreshold:
		return models.SeverityCritical
	case loc >= highThreshold:
		return models.SeverityHigh
	case loc >= mediumThreshold:
		return models.SeverityMedium
	default:
		return models.SeverityLow
	}
}

func (v *functionLengthVisitor) createLengthIssue(fn *ast.FuncDecl, funcName string, actualLOC, totalLines int, severity models.Severity) {
	position := v.fset.Position(fn.Pos())

	issue := models.Issue{
		Type:        models.IssueFunctionLength,
		Severity:    severity,
		File:        v.filename,
		Line:        position.Line,
		Column:      position.Column,
		Function:    funcName,
		Message:     v.generateMessage(funcName, actualLOC, totalLines),
		Suggestion:  v.generateSuggestion(severity, actualLOC),
		Complexity:  fmt.Sprintf("Function length: %d lines", actualLOC),
		CodeSnippet: position.String(),
	}

	v.issues = append(v.issues, issue)
}

func (v *functionLengthVisitor) generateMessage(funcName string, actualLOC, totalLines int) string {
	return fmt.Sprintf("Function '%s' is too long (%d lines of code, %d total lines) - consider breaking into smaller functions",
		funcName, actualLOC, totalLines)
}

func (v *functionLengthVisitor) generateSuggestion(severity models.Severity, loc int) string {
	baseAdvice := `Long functions are harder to understand, test, and maintain. Consider refactoring using these techniques:

1. **Extract Method**: Move logical blocks into separate functions
2. **Single Responsibility**: Ensure function does only one thing  
3. **Reduce Nesting**: Use early returns to flatten conditional logic
4. **Group Related Code**: Extract helper functions for repeated patterns`

	switch severity {
	case models.SeverityMedium:
		return baseAdvice + `

Target: Break into 2-3 smaller functions of ~20-30 lines each.

Example refactoring:
// Instead of one 60-line function:
func ProcessData() { /* 60 lines */ }

// Break into:
func ProcessData() {
    data := loadData()
    validated := validateData(data)
    return transformData(validated)
}
func loadData() { /* 15 lines */ }
func validateData() { /* 20 lines */ }
func transformData() { /* 15 lines */ }`

	case models.SeverityHigh:
		return baseAdvice + fmt.Sprintf(`

PRIORITY: This %d-line function significantly exceeds recommended limits.

Refactoring strategy:
1. Identify 3-5 main logical sections
2. Extract each section into a separate function
3. Use meaningful function names that describe intent
4. Consider if this indicates a class/struct is needed

Target: Break into 4-6 functions of ~15-25 lines each.`, loc)

	case models.SeverityCritical:
		return baseAdvice + fmt.Sprintf(`

🚨 CRITICAL: This %d-line function is extremely difficult to maintain!

Immediate action required:
1. **Stop adding features** to this function
2. **Extract at least 5-8 smaller functions** immediately  
3. **Consider architectural changes** - may need multiple files/packages
4. **Add comprehensive tests** before refactoring
5. **Document the refactoring plan** before starting

This function likely violates Single Responsibility Principle.
Consider if it needs to be split into multiple types/interfaces.`, loc)

	default:
		return baseAdvice
	}
}
