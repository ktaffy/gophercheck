package detectors

import (
	"fmt"
	"go/ast"
	"go/token"

	"gophercheck/internal/config"
	"gophercheck/internal/context"
	"gophercheck/internal/models"
)

// ComplexityDetector calculates cyclomatic complexity of functions
type ComplexityDetector struct {
	config *config.Config
}

// NewComplexityDetector creates a new complexity detector
func NewComplexityDetector() *ComplexityDetector {
	return &ComplexityDetector{}
}

func NewComplexityDetectorWithConfig(cfg *config.Config) *ComplexityDetector {
	return &ComplexityDetector{
		config: cfg,
	}
}

func (d *ComplexityDetector) SetConfig(cfg *config.Config) {
	d.config = cfg
}

func (d *ComplexityDetector) Name() string {
	return "Cyclomatic Complexity Detector"
}

func (d *ComplexityDetector) Detect(file *ast.File, fset *token.FileSet, filename string, ctx *context.AnalysisContext) []models.Issue {
	detector := &complexityVisitor{
		fset:     fset,
		filename: filename,
		issues:   make([]models.Issue, 0),
		detector: d,
		context:  ctx,
	}

	ast.Walk(detector, file)
	return detector.issues
}

type complexityVisitor struct {
	fset     *token.FileSet
	filename string
	issues   []models.Issue
	detector *ComplexityDetector
	context  *context.AnalysisContext
}

func (v *complexityVisitor) Visit(node ast.Node) ast.Visitor {
	if fn, ok := node.(*ast.FuncDecl); ok && fn.Body != nil {
		complexity := v.calculateComplexity(fn.Body)
		threshold := 10
		if v.detector.config != nil && v.detector.config.Rules.Complexity.CyclomaticComplexity.Enabled {
			threshold = v.detector.config.Rules.Complexity.CyclomaticComplexity.MediumThreshold
		}
		if complexity > threshold {
			v.createComplexityIssue(fn, complexity)
		}
	}
	return v
}

func (v *complexityVisitor) calculateComplexity(body *ast.BlockStmt) int {
	complexity := 1 // Base complexity

	ast.Inspect(body, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.IfStmt:
			complexity++
			// Count else if as additional complexity
			if node.Else != nil {
				if _, ok := node.Else.(*ast.IfStmt); ok {
					// This is handled when we visit the else if
				} else {
					// This is a simple else
					complexity++
				}
			}

		case *ast.ForStmt, *ast.RangeStmt:
			complexity++

		case *ast.TypeSwitchStmt, *ast.SwitchStmt:
			complexity++

		case *ast.CaseClause:
			// Each case adds complexity (except default)
			if node.List != nil { // nil means default case
				complexity++
			}

		case *ast.CommClause:
			// Select statement cases
			complexity++

		case *ast.FuncLit:
			// Don't count complexity inside function literals
			return false

		case *ast.BinaryExpr:
			// Logical operators add complexity
			if node.Op == token.LAND || node.Op == token.LOR {
				complexity++
			}
		}
		return true
	})

	return complexity
}

func (v *complexityVisitor) createComplexityIssue(fn *ast.FuncDecl, complexity int) {
	position := v.fset.Position(fn.Pos())
	funcName := "anonymous"
	if fn.Name != nil {
		funcName = fn.Name.Name
	}

	issue := models.Issue{
		Type:        models.IssueCyclomaticComplex,
		Severity:    v.calculateSeverity(complexity),
		File:        v.filename,
		Line:        position.Line,
		Column:      position.Column,
		Function:    funcName,
		Message:     fmt.Sprintf("Function '%s' has high cyclomatic complexity: %d", funcName, complexity),
		Suggestion:  v.generateComplexitySuggestion(complexity),
		Complexity:  fmt.Sprintf("Complexity: %d", complexity),
		CodeSnippet: position.String(),
	}

	v.issues = append(v.issues, issue)
}

func (v *complexityVisitor) calculateSeverity(complexity int) models.Severity {
	mediumThreshold := 10
	highThreshold := 15
	criticalThreshold := 25

	if v.detector.config != nil && v.detector.config.Rules.Complexity.CyclomaticComplexity.Enabled {
		mediumThreshold = v.detector.config.Rules.Complexity.CyclomaticComplexity.MediumThreshold
		highThreshold = v.detector.config.Rules.Complexity.CyclomaticComplexity.HighThreshold
		criticalThreshold = v.detector.config.Rules.Complexity.CyclomaticComplexity.CriticalThreshold
	}

	switch {
	case complexity <= mediumThreshold:
		return models.SeverityMedium
	case complexity <= highThreshold:
		return models.SeverityHigh
	case complexity <= criticalThreshold:
		return models.SeverityCritical
	default:
		return models.SeverityCritical
	}
}

func (v *complexityVisitor) generateComplexitySuggestion(complexity int) string {
	suggestions := []string{
		"Consider breaking this function into smaller, single-purpose functions",
		"Use early returns to reduce nesting levels",
		"Extract complex conditional logic into separate functions",
		"Consider using a state machine or strategy pattern for complex branching",
		"Use lookup tables or maps instead of long if-else chains",
	}

	if complexity <= 15 {
		return suggestions[0] + ". " + suggestions[1]
	} else if complexity <= 25 {
		return suggestions[0] + ". " + suggestions[2] + ". " + suggestions[1]
	} else {
		return suggestions[3] + ". " + suggestions[0] + ". " + suggestions[4]
	}
}
