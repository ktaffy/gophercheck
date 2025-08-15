package detectors

import (
	"fmt"
	"go/ast"
	"go/token"

	"gophercheck/internal/models"
)

type NestedLoopDetector struct{}

func NewNestedLoopDetector() *NestedLoopDetector {
	return &NestedLoopDetector{}
}

func (d *NestedLoopDetector) Name() string {
	return "Nested Loop Detector"
}

func (d *NestedLoopDetector) Detect(file *ast.File, fset *token.FileSet, filename string) []models.Issue {
	detector := &nestedLoopVisitor{
		fset:     fset,
		filename: filename,
		issues:   make([]models.Issue, 0),
	}
	ast.Walk(detector, file)
	return detector.issues
}

type nestedLoopVisitor struct {
	fset        *token.FileSet
	filename    string
	issues      []models.Issue
	loopDepth   int
	currentFunc string
}

func (v *nestedLoopVisitor) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.FuncDecl:
		// Track current function for better error reporting
		if n.Name != nil {
			v.currentFunc = n.Name.Name
		}
		return v

	case *ast.ForStmt, *ast.RangeStmt:
		// We found a loop
		v.loopDepth++

		// If we're already in a loop (depth > 1), this is nested
		if v.loopDepth > 1 {
			v.detectNestedLoop(n)
		}

		// Continue visiting children
		for _, child := range getLoopBody(n) {
			ast.Walk(v, child)
		}

		v.loopDepth--
		return nil // Don't visit children again

	default:
		return v
	}
}

func (v *nestedLoopVisitor) detectNestedLoop(node ast.Node) {
	pos := getNodePosition(node)
	position := v.fset.Position(pos)

	issue := models.Issue{
		Type:        models.IssueNestedLoops,
		Severity:    v.calculateSeverity(),
		File:        v.filename,
		Line:        position.Line,
		Column:      position.Column,
		Function:    v.currentFunc,
		Message:     v.generateMessage(),
		Suggestion:  v.generateSuggestion(),
		Complexity:  fmt.Sprintf("O(n^%d)", v.loopDepth),
		CodeSnippet: position.String(),
	}

	v.issues = append(v.issues, issue)
}

func (v *nestedLoopVisitor) calculateSeverity() models.Severity {
	switch v.loopDepth {
	case 2:
		return models.SeverityMedium // O(n²) is concerning but common
	case 3:
		return models.SeverityHigh // O(n³) is usually problematic
	default:
		return models.SeverityCritical // O(n⁴+) is almost always wrong
	}
}

func (v *nestedLoopVisitor) generateMessage() string {
	if v.loopDepth == 2 {
		return fmt.Sprintf("Nested loop detected in function '%s' - potential O(n²) complexity", v.currentFunc)
	}
	return fmt.Sprintf("Deeply nested loops detected in function '%s' - O(n^%d) complexity", v.currentFunc, v.loopDepth)
}

func (v *nestedLoopVisitor) generateSuggestion() string {
	suggestions := []string{
		"Consider using a map for O(1) lookups instead of nested iteration",
		"Pre-process data into a more efficient structure (e.g., hash map)",
		"Use algorithms like binary search if data is sorted",
		"Consider if you can break/continue early to reduce iterations",
		"Profile this code section to measure actual performance impact",
	}

	// Return different suggestions based on depth
	if v.loopDepth == 2 {
		return suggestions[0] + ". " + suggestions[1]
	}
	return suggestions[2] + ". " + suggestions[4]
}

// Helper functions

func getLoopBody(node ast.Node) []ast.Stmt {
	switch n := node.(type) {
	case *ast.ForStmt:
		if n.Body != nil {
			return n.Body.List
		}
	case *ast.RangeStmt:
		if n.Body != nil {
			return n.Body.List
		}
	}
	return nil
}

func getNodePosition(node ast.Node) token.Pos {
	switch n := node.(type) {
	case *ast.ForStmt:
		return n.For
	case *ast.RangeStmt:
		return n.For
	default:
		return token.NoPos
	}
}
