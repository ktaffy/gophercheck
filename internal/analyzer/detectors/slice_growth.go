package detectors

import (
	"fmt"
	"go/ast"
	"go/token"
	"gophercheck/internal/models"
)

// SliceGrowthDetector finds inefficient slice growth patterns
type SliceGrowthDetector struct{}

func NewSliceGrowthDetector() *SliceGrowthDetector {
	return &SliceGrowthDetector{}
}

func (d *SliceGrowthDetector) Name() string {
	return "Slice Growth Pattern Detector"
}

func (d *SliceGrowthDetector) Detect(file *ast.File, fset *token.FileSet, filename string) []models.Issue {
	detector := &sliceGrowthVisitor{
		fset:        fset,
		filename:    filename,
		issues:      make([]models.Issue, 0),
		sliceVars:   make(map[string]*sliceInfo),
		currentFunc: "",
	}

	ast.Walk(detector, file)
	return detector.issues
}

type sliceInfo struct {
	name         string
	declaredLine int
	hasCapacity  bool
	usedInLoop   bool
	appendCount  int
}

type sliceGrowthVisitor struct {
	fset        *token.FileSet
	filename    string
	issues      []models.Issue
	sliceVars   map[string]*sliceInfo
	currentFunc string
	inLoop      bool
	loopDepth   int
}

func (v *sliceGrowthVisitor) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.FuncDecl:
		// Reset slice tracking for each function
		v.sliceVars = make(map[string]*sliceInfo)
		if n.Name != nil {
			v.currentFunc = n.Name.Name
		}
		return v

	case *ast.ForStmt, *ast.RangeStmt:
		v.loopDepth++
		oldInLoop := v.inLoop
		v.inLoop = true

		// Mark existing slices as used in loop
		for _, info := range v.sliceVars {
			info.usedInLoop = true
		}

		// Visit loop body
		for _, stmt := range getLoopBody(n) {
			ast.Walk(v, stmt)
		}

		v.loopDepth--
		v.inLoop = oldInLoop
		return nil

	case *ast.AssignStmt:
		v.checkSliceAssignment(n)
		return v

	case *ast.GenDecl:
		v.checkSliceDeclaration(n)
		return v

	default:
		return v
	}
}

func (v *sliceGrowthVisitor) checkSliceDeclaration(decl *ast.GenDecl) {
	if decl.Tok != token.VAR {
		return
	}

	for _, spec := range decl.Specs {
		if valueSpec, ok := spec.(*ast.ValueSpec); ok {
			for i, name := range valueSpec.Names {
				if i < len(valueSpec.Values) {
					if v.isSliceMake(valueSpec.Values[i]) {
						position := v.fset.Position(name.Pos())
						hasCapacity := v.sliceMakeHasCapacity(valueSpec.Values[i])

						v.sliceVars[name.Name] = &sliceInfo{
							name:         name.Name,
							declaredLine: position.Line,
							hasCapacity:  hasCapacity,
							usedInLoop:   false,
							appendCount:  0,
						}

						// Check for slice without capacity
						if !hasCapacity {
							v.createSliceGrowthIssue(name, "Slice declared without capacity hint")
						}
					}
				}
			}
		}
	}
}

func (v *sliceGrowthVisitor) checkSliceAssignment(assign *ast.AssignStmt) {
	// Check for slice := make([]T, 0) patterns
	if assign.Tok == token.DEFINE && len(assign.Lhs) == 1 && len(assign.Rhs) == 1 {
		if ident, ok := assign.Lhs[0].(*ast.Ident); ok {
			if v.isSliceMake(assign.Rhs[0]) {
				position := v.fset.Position(ident.Pos())
				hasCapacity := v.sliceMakeHasCapacity(assign.Rhs[0])

				v.sliceVars[ident.Name] = &sliceInfo{
					name:         ident.Name,
					declaredLine: position.Line,
					hasCapacity:  hasCapacity,
					usedInLoop:   v.inLoop,
					appendCount:  0,
				}

				if !hasCapacity && v.inLoop {
					v.createSliceGrowthIssue(ident, "Slice created in loop without capacity")
				}
			}
		}
	}

	// Check for append operations
	if len(assign.Rhs) == 1 {
		if call, ok := assign.Rhs[0].(*ast.CallExpr); ok {
			if v.isAppendCall(call) {
				v.trackAppendUsage(assign, call)
			}
		}
	}
}

func (v *sliceGrowthVisitor) trackAppendUsage(assign *ast.AssignStmt, call *ast.CallExpr) {
	if len(assign.Lhs) > 0 {
		if ident, ok := assign.Lhs[0].(*ast.Ident); ok {
			if info, exists := v.sliceVars[ident.Name]; exists {
				info.appendCount++

				// Issue if appending in loop without capacity
				if v.inLoop && !info.hasCapacity && info.appendCount > 1 {
					v.createAppendIssue(assign, fmt.Sprintf("Multiple appends to slice '%s' in loop without pre-allocation", ident.Name))
				}
			}
		}
	}
}

func (v *sliceGrowthVisitor) isSliceMake(expr ast.Expr) bool {
	if call, ok := expr.(*ast.CallExpr); ok {
		if ident, ok := call.Fun.(*ast.Ident); ok && ident.Name == "make" {
			if len(call.Args) > 0 {
				return v.isSliceType(call.Args[0])
			}
		}
	}
	return false
}

func (v *sliceGrowthVisitor) sliceMakeHasCapacity(expr ast.Expr) bool {
	if call, ok := expr.(*ast.CallExpr); ok {
		if ident, ok := call.Fun.(*ast.Ident); ok && ident.Name == "make" {
			// make([]T, len, cap) has 3 args, make([]T, len) has 2
			return len(call.Args) >= 3
		}
	}
	return false
}

func (v *sliceGrowthVisitor) isSliceType(expr ast.Expr) bool {
	if arrayType, ok := expr.(*ast.ArrayType); ok {
		return arrayType.Len == nil // slice if no length specified
	}
	return false
}

func (v *sliceGrowthVisitor) isAppendCall(call *ast.CallExpr) bool {
	if ident, ok := call.Fun.(*ast.Ident); ok {
		return ident.Name == "append"
	}
	return false
}

func (v *sliceGrowthVisitor) createSliceGrowthIssue(node ast.Node, message string) {
	var pos token.Pos
	switch n := node.(type) {
	case *ast.Ident:
		pos = n.Pos()
	default:
		pos = token.NoPos
	}

	position := v.fset.Position(pos)

	issue := models.Issue{
		Type:        models.IssueSliceGrowth,
		Severity:    models.SeverityMedium,
		File:        v.filename,
		Line:        position.Line,
		Column:      position.Column,
		Function:    v.currentFunc,
		Message:     message + " - may cause multiple reallocations",
		Suggestion:  v.generateSliceGrowthSuggestion(),
		Complexity:  "O(n) amortized growth cost",
		CodeSnippet: position.String(),
	}

	v.issues = append(v.issues, issue)
}

func (v *sliceGrowthVisitor) createAppendIssue(assign *ast.AssignStmt, message string) {
	position := v.fset.Position(assign.Pos())

	issue := models.Issue{
		Type:        models.IssueSliceGrowth,
		Severity:    models.SeverityHigh,
		File:        v.filename,
		Line:        position.Line,
		Column:      position.Column,
		Function:    v.currentFunc,
		Message:     message,
		Suggestion:  v.generateAppendInLoopSuggestion(),
		Complexity:  "O(n log n) due to slice growth",
		CodeSnippet: position.String(),
	}

	v.issues = append(v.issues, issue)
}

func (v *sliceGrowthVisitor) generateSliceGrowthSuggestion() string {
	return `Pre-allocate slice capacity when size is known or predictable:

// Instead of:
slice := make([]T, 0)  // Will grow as needed

// Do this:
slice := make([]T, 0, expectedSize)  // Pre-allocate capacity

// Or if you know the exact size:
slice := make([]T, expectedSize)  // Pre-allocate length and capacity

This prevents multiple memory allocations and copying during growth.`
}

func (v *sliceGrowthVisitor) generateAppendInLoopSuggestion() string {
	return `Pre-allocate slice capacity before loop to avoid repeated growth:

// Instead of:
var results []T
for _, item := range items {
    results = append(results, process(item))  // Grows each iteration
}

// Do this:
results := make([]T, 0, len(items))  // Pre-allocate capacity
for _, item := range items {
    results = append(results, process(item))  // No reallocation needed
}

This changes complexity from O(n log n) to O(n).`
}
