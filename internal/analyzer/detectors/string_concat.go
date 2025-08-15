package detectors

import (
	"go/ast"
	"go/token"

	"gophercheck/internal/models"
)

// StringConcatDetector finds inefficient string concatenation in loops
type StringConcatDetector struct{}

// NewStringConcatDetector creates a new string concatenation detector
func NewStringConcatDetector() *StringConcatDetector {
	return &StringConcatDetector{}
}

// Name returns the detector name
func (d *StringConcatDetector) Name() string {
	return "String Concatenation Detector"
}

// Detect finds inefficient string concatenation patterns
func (d *StringConcatDetector) Detect(file *ast.File, fset *token.FileSet, filename string) []models.Issue {
	detector := &stringConcatVisitor{
		fset:     fset,
		filename: filename,
		issues:   make([]models.Issue, 0),
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
}

// Visit implements the ast.Visitor interface
func (v *stringConcatVisitor) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.FuncDecl:
		if n.Name != nil {
			v.currentFunc = n.Name.Name
		}
		return v

	case *ast.ForStmt, *ast.RangeStmt:
		// Mark that we're entering a loop
		oldInLoop := v.inLoop
		v.inLoop = true

		// Visit loop body
		for _, stmt := range getLoopBody(n) {
			ast.Walk(v, stmt)
		}

		v.inLoop = oldInLoop
		return nil // Don't visit children again

	case *ast.AssignStmt:
		if v.inLoop {
			v.checkStringConcatenation(n)
		}
		return v

	default:
		return v
	}
}

// checkStringConcatenation looks for string concatenation patterns in assignments
func (v *stringConcatVisitor) checkStringConcatenation(assign *ast.AssignStmt) {
	// Look for patterns like: str += something or str = str + something
	if len(assign.Lhs) != 1 || len(assign.Rhs) != 1 {
		return
	}

	// Check for += operator
	if assign.Tok == token.ADD_ASSIGN {
		if v.isStringVariable(assign.Lhs[0]) {
			v.createIssue(assign, "String concatenation using += in loop")
		}
		return
	}

	// Check for str = str + something pattern
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

// isStringVariable checks if an expression likely represents a string variable
// This is simplified - a full implementation would use type information
func (v *stringConcatVisitor) isStringVariable(expr ast.Expr) bool {
	// For now, we'll use heuristics based on common string variable names
	if ident, ok := expr.(*ast.Ident); ok {
		name := ident.Name
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

// sameVariable checks if two expressions refer to the same variable
func (v *stringConcatVisitor) sameVariable(expr1, expr2 ast.Expr) bool {
	ident1, ok1 := expr1.(*ast.Ident)
	ident2, ok2 := expr2.(*ast.Ident)

	if ok1 && ok2 {
		return ident1.Name == ident2.Name
	}
	return false
}

// createIssue creates a performance issue for string concatenation
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

// generateSuggestion provides actionable advice for string concatenation
func (v *stringConcatVisitor) generateSuggestion() string {
	return `Use strings.Builder for efficient string concatenation:
	
var builder strings.Builder
for _, item := range items {
    builder.WriteString(item)
}
result := builder.String()

This provides O(n) performance instead of O(n²).`
}
