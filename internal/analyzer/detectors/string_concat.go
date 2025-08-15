package detectors

import (
	"go/ast"
	"go/token"
	"strings"

	"gophercheck/internal/config"
	"gophercheck/internal/models"
)

type StringConcatDetector struct {
	config *config.Config
}

func NewStringConcatDetector() *StringConcatDetector {
	return &StringConcatDetector{}
}

func NewStringConcatDetectorWithConfig(cfg *config.Config) *StringConcatDetector {
	return &StringConcatDetector{
		config: cfg,
	}
}

func (d *StringConcatDetector) SetConfig(cfg *config.Config) {
	d.config = cfg
}

func (d *StringConcatDetector) Name() string {
	return "String Concatenation Detector"
}

func (d *StringConcatDetector) Detect(file *ast.File, fset *token.FileSet, filename string) []models.Issue {
	detector := &stringConcatVisitor{
		fset:     fset,
		filename: filename,
		issues:   make([]models.Issue, 0),
		detector: d,
	}

	ast.Walk(detector, file)
	return detector.issues
}

type stringConcatVisitor struct {
	fset        *token.FileSet
	filename    string
	issues      []models.Issue
	inLoop      bool
	currentFunc string
	detector    *StringConcatDetector
}

func (v *stringConcatVisitor) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.FuncDecl:
		if n.Name != nil {
			v.currentFunc = n.Name.Name
		}
		return v

	case *ast.ForStmt, *ast.RangeStmt:
		oldInLoop := v.inLoop
		v.inLoop = true

		for _, stmt := range getLoopBody(n) {
			ast.Walk(v, stmt)
		}

		v.inLoop = oldInLoop
		return nil

	case *ast.AssignStmt:
		if v.inLoop {
			v.checkStringConcatenation(n)
		}
		return v

	default:
		return v
	}
}

func (v *stringConcatVisitor) checkStringConcatenation(assign *ast.AssignStmt) {
	detectInLoops := true // default
	if v.detector.config != nil && v.detector.config.Rules.Performance.StringConcat.Enabled {
		detectInLoops = v.detector.config.Rules.Performance.StringConcat.DetectInLoops
	}

	if !v.inLoop || !detectInLoops {
		return
	}

	if len(assign.Lhs) != 1 || len(assign.Rhs) != 1 {
		return
	}

	if assign.Tok == token.ADD_ASSIGN {
		if v.isStringVariable(assign.Lhs[0]) {
			v.createIssue(assign, "String concatenation using += in loop")
		}
		return
	}

	if assign.Tok == token.ASSIGN {
		if binExpr, ok := assign.Rhs[0].(*ast.BinaryExpr); ok {
			if binExpr.Op == token.ADD && v.isStringVariable(assign.Lhs[0]) {
				// Check if left side of addition matches the assignment target
				if v.sameVariable(assign.Lhs[0], binExpr.X) {
					v.createIssue(assign, "String concatenation using + in loop")
				}
			}
		}
	}
}

// This is simplified - a full implementation would use type information
func (v *stringConcatVisitor) isStringVariable(expr ast.Expr) bool {
	// For now, we'll use heuristics based on common string variable names
	if ident, ok := expr.(*ast.Ident); ok {
		name := ident.Name

		// Use config string variable names if available
		if v.detector.config != nil && v.detector.config.Rules.Performance.StringConcat.Enabled {
			configNames := v.detector.config.Rules.Performance.StringConcat.StringVarNames
			if len(configNames) > 0 {
				for _, configName := range configNames {
					if name == configName || strings.Contains(strings.ToLower(name), strings.ToLower(configName)) {
						return true
					}
				}
				return false // If config names specified, only use those
			}
		}

		// Common string variable names
		stringNames := []string{"str", "result", "output", "text", "content", "message", "data"}
		for _, sname := range stringNames {
			if name == sname ||
				len(name) >= 3 && (name[:3] == "str" || name[len(name)-3:] == "Str") {
				return true
			}
		}
	}
	return false
}

func (v *stringConcatVisitor) sameVariable(expr1, expr2 ast.Expr) bool {
	ident1, ok1 := expr1.(*ast.Ident)
	ident2, ok2 := expr2.(*ast.Ident)

	if ok1 && ok2 {
		return ident1.Name == ident2.Name
	}
	return false
}

func (v *stringConcatVisitor) createIssue(assign *ast.AssignStmt, message string) {
	position := v.fset.Position(assign.Pos())

	issue := models.Issue{
		Type:        models.IssueStringConcat,
		Severity:    models.SeverityMedium,
		File:        v.filename,
		Line:        position.Line,
		Column:      position.Column,
		Function:    v.currentFunc,
		Message:     message + " - creates new strings on each iteration",
		Suggestion:  v.generateSuggestion(),
		Complexity:  "O(n²) due to string copying",
		CodeSnippet: position.String(),
	}

	v.issues = append(v.issues, issue)
}

func (v *stringConcatVisitor) generateSuggestion() string {
	return `Use strings.Builder for efficient string concatenation:
	
var builder strings.Builder
for _, item := range items {
    builder.WriteString(item)
}
result := builder.String()

This provides O(n) performance instead of O(n²).`
}
