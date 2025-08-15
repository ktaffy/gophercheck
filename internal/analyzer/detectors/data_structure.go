package detectors

import (
	"fmt"
	"go/ast"
	"go/token"
	"gophercheck/internal/models"
)

// DataStructureDetector analyzes data access patterns and suggests optimal structures
type DataStructureDetector struct{}

func NewDataStructureDetector() *DataStructureDetector {
	return &DataStructureDetector{}
}

func (d *DataStructureDetector) Name() string {
	return "Data Structure Usage Detector"
}

func (d *DataStructureDetector) Detect(file *ast.File, fset *token.FileSet, filename string) []models.Issue {
	detector := &dataStructureVisitor{
		fset:        fset,
		filename:    filename,
		issues:      make([]models.Issue, 0),
		currentFunc: "",
		inLoop:      false,
		loopDepth:   0,
	}

	ast.Walk(detector, file)
	return detector.issues
}

type dataStructureVisitor struct {
	fset        *token.FileSet
	filename    string
	issues      []models.Issue
	currentFunc string
	inLoop      bool
	loopDepth   int
}

func (v *dataStructureVisitor) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.FuncDecl:
		if n.Name != nil {
			v.currentFunc = n.Name.Name
		}
		return v

	case *ast.ForStmt, *ast.RangeStmt:
		v.loopDepth++
		oldInLoop := v.inLoop
		v.inLoop = true

		// Check for linear search patterns in range loops
		if rangeStmt, ok := n.(*ast.RangeStmt); ok {
			v.checkForLinearSearch(rangeStmt)
		}

		// Visit loop body
		for _, stmt := range getLoopBody(n) {
			ast.Walk(v, stmt)
		}

		v.loopDepth--
		v.inLoop = oldInLoop
		return nil

	default:
		return v
	}
}

// checkForLinearSearch looks for range loops that contain equality comparisons
func (v *dataStructureVisitor) checkForLinearSearch(rangeStmt *ast.RangeStmt) {
	// Look for patterns like: for _, item := range slice { if item.field == target { ... } }
	if rangeStmt.Body != nil {
		foundComparison := false
		var comparisonPos token.Pos

		ast.Inspect(rangeStmt.Body, func(n ast.Node) bool {
			// Look for binary expressions with equality operators
			if binExpr, ok := n.(*ast.BinaryExpr); ok {
				if binExpr.Op == token.EQL { // == operator
					foundComparison = true
					comparisonPos = binExpr.Pos()
					return false // Stop searching
				}
			}
			return true
		})

		if foundComparison {
			v.createLinearSearchIssue(rangeStmt, comparisonPos)
		}
	}
}

// createLinearSearchIssue creates an issue for linear search patterns
func (v *dataStructureVisitor) createLinearSearchIssue(rangeStmt *ast.RangeStmt, comparisonPos token.Pos) {
	position := v.fset.Position(rangeStmt.Pos())

	// Try to get the slice variable name
	sliceName := "slice"
	if ident, ok := rangeStmt.X.(*ast.Ident); ok {
		sliceName = ident.Name
	}

	issue := models.Issue{
		Type:        models.IssueInefficinetDS,
		Severity:    models.SeverityMedium,
		File:        v.filename,
		Line:        position.Line,
		Column:      position.Column,
		Function:    v.currentFunc,
		Message:     fmt.Sprintf("Linear search detected in range loop over '%s' - consider using a map for O(1) lookups", sliceName),
		Suggestion:  v.generateLinearSearchSuggestion(sliceName),
		Complexity:  "O(n) search → O(1) with map",
		CodeSnippet: position.String(),
	}

	v.issues = append(v.issues, issue)
}

func (v *dataStructureVisitor) generateLinearSearchSuggestion(sliceName string) string {
	return fmt.Sprintf(`Consider using a map for O(1) lookups instead of O(n) linear search:

// Instead of:
for _, item := range %s {
    if item.ID == targetID {  // O(n) search
        return item
    }
}

// Do this:
%sMap := make(map[int]Item, len(%s))  // Pre-size for efficiency
for _, item := range %s {
    %sMap[item.ID] = item
}
result := %sMap[targetID]  // O(1) lookup

This changes complexity from O(n) to O(1) for lookups.
If you need to do multiple searches, the preprocessing cost is amortized.`,
		sliceName, sliceName, sliceName, sliceName, sliceName, sliceName)
}
