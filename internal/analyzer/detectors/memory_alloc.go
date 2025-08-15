package detectors

import (
	"fmt"
	"go/ast"
	"go/token"
	"gophercheck/internal/config"
	"gophercheck/internal/models"
	"strings"
)

type MemoryAllocDetector struct {
	config *config.Config
}

func NewMemoryAllocDetector() *MemoryAllocDetector {
	return &MemoryAllocDetector{}
}

func NewMemoryAllocDetectorWithConfig(cfg *config.Config) *MemoryAllocDetector {
	return &MemoryAllocDetector{
		config: cfg,
	}
}

func (d *MemoryAllocDetector) SetConfig(cfg *config.Config) {
	d.config = cfg
}

func (d *MemoryAllocDetector) Name() string {
	return "Memory Allocation Detector"
}

func (d *MemoryAllocDetector) Detect(file *ast.File, fset *token.FileSet, filename string) []models.Issue {
	detector := &memoryAllocVisitor{
		fset:        fset,
		filename:    filename,
		issues:      make([]models.Issue, 0),
		loopDepth:   0,
		currentFunc: "",
		detector:    d,
	}
	ast.Walk(detector, file)
	return detector.issues
}

type memoryAllocVisitor struct {
	fset        *token.FileSet
	filename    string
	issues      []models.Issue
	loopDepth   int
	currentFunc string
	inLoop      bool
	detector    *MemoryAllocDetector
}

func (v *memoryAllocVisitor) Visit(node ast.Node) ast.Visitor {
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

		for _, stmt := range getLoopBody(n) {
			ast.Walk(v, stmt)
		}

		v.loopDepth--
		v.inLoop = oldInLoop
		return nil
	case *ast.CallExpr:
		if v.inLoop {
			v.checkAllocationInLoop(n)
		}
		v.checkInefficientAllocation(n)
		return v
	case *ast.AssignStmt:
		if v.inLoop {
			v.checkAppendWithoutPrealloc(n)
		}
		return v
	default:
		return v
	}
}

func (v *memoryAllocVisitor) checkAllocationInLoop(call *ast.CallExpr) {
	detectInLoops := true // default
	if v.detector.config != nil && v.detector.config.Rules.Memory.Allocation.Enabled {
		detectInLoops = v.detector.config.Rules.Memory.Allocation.DetectInLoops
	}

	if !detectInLoops {
		return
	}

	if v.isAllocationCall(call) {
		allocType := v.getAllocationType(call)
		v.createIssue(call, fmt.Sprintf("Memory allocation (%s) inside loop", allocType), v.generateLoopAllocationSuggestion(allocType), models.SeverityHigh)
	}

}

func (v *memoryAllocVisitor) checkInefficientAllocation(call *ast.CallExpr) {
	requireCapacityHints := true // default
	if v.detector.config != nil && v.detector.config.Rules.Memory.Allocation.Enabled {
		requireCapacityHints = v.detector.config.Rules.Memory.Allocation.RequireCapacityHints
	}

	if !requireCapacityHints {
		return
	}

	if v.isMakeSliceWithoutCapacity(call) {
		v.createIssue(call,
			"Slice created without capacity hint - may cause multiple reallocations",
			v.generateCapacitySuggestion(),
			models.SeverityMedium)
	}

	if v.isMakeMapWithoutSize(call) {
		v.createIssue(call,
			"Map created without size hint - may cause rehashing",
			v.generateMapSizeSuggestion(),
			models.SeverityLow)
	}
}

func (v *memoryAllocVisitor) checkAppendWithoutPrealloc(assign *ast.AssignStmt) {
	minLoopIterations := 5 // default
	if v.detector.config != nil && v.detector.config.Rules.Memory.Allocation.Enabled {
		minLoopIterations = v.detector.config.Rules.Memory.Allocation.MinLoopIterations
	}

	if v.loopDepth < minLoopIterations {
		return
	}

	if len(assign.Rhs) == 1 {
		if call, ok := assign.Rhs[0].(*ast.CallExpr); ok {
			if v.isAppendCall(call) && v.loopDepth > 0 {
				v.createIssue(assign,
					"append() in loop without preallocation - causes slice growth",
					v.generateAppendSuggestion(),
					models.SeverityMedium)
			}
		}
	}
}

// Helper functions to identify allocation patterns

func (v *memoryAllocVisitor) isAllocationCall(call *ast.CallExpr) bool {
	if ident, ok := call.Fun.(*ast.Ident); ok {
		return ident.Name == "make" || ident.Name == "new"
	}
	return false
}

func (v *memoryAllocVisitor) getAllocationType(call *ast.CallExpr) string {
	if ident, ok := call.Fun.(*ast.Ident); ok {
		if ident.Name == "make" && len(call.Args) > 0 {
			return fmt.Sprintf("make(%s)", v.getTypeString(call.Args[0]))
		}
		if ident.Name == "new" && len(call.Args) > 0 {
			return fmt.Sprintf("new(%s)", v.getTypeString(call.Args[0]))
		}
		return ident.Name
	}
	return "allocation"
}

func (v *memoryAllocVisitor) isMakeSliceWithoutCapacity(call *ast.CallExpr) bool {
	if ident, ok := call.Fun.(*ast.Ident); ok && ident.Name == "make" {
		if len(call.Args) >= 2 {
			// Check if it's a slice type
			if v.isSliceType(call.Args[0]) {
				// Has length but no capacity (make([]T, len) vs make([]T, len, cap))
				return len(call.Args) == 2
			}
		}
	}
	return false
}

func (v *memoryAllocVisitor) isMakeMapWithoutSize(call *ast.CallExpr) bool {
	if ident, ok := call.Fun.(*ast.Ident); ok && ident.Name == "make" {
		if len(call.Args) == 1 && v.isMapType(call.Args[0]) {
			return true
		}
	}
	return false
}

func (v *memoryAllocVisitor) isAppendCall(call *ast.CallExpr) bool {
	if ident, ok := call.Fun.(*ast.Ident); ok {
		return ident.Name == "append"
	}
	return false
}

func (v *memoryAllocVisitor) isSliceType(expr ast.Expr) bool {
	// Simple heuristic - check if type looks like []T
	typeStr := v.getTypeString(expr)
	return strings.HasPrefix(typeStr, "[]")
}

func (v *memoryAllocVisitor) isMapType(expr ast.Expr) bool {
	// Simple heuristic - check if type looks like map[K]V
	typeStr := v.getTypeString(expr)
	return strings.HasPrefix(typeStr, "map[")
}

func (v *memoryAllocVisitor) getTypeString(expr ast.Expr) string {
	// Simplified type string extraction
	switch t := expr.(type) {
	case *ast.ArrayType:
		if t.Len == nil { // slice
			return "[]" + v.getTypeString(t.Elt)
		}
		return fmt.Sprintf("[%s]%s", v.getExprString(t.Len), v.getTypeString(t.Elt))
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", v.getTypeString(t.Key), v.getTypeString(t.Value))
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return v.getExprString(t.X) + "." + t.Sel.Name
	default:
		return "unknown"
	}
}

func (v *memoryAllocVisitor) getExprString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.BasicLit:
		return e.Value
	case *ast.SelectorExpr:
		return v.getExprString(e.X) + "." + e.Sel.Name
	default:
		return "expr"
	}
}

// Suggestion generators

func (v *memoryAllocVisitor) generateLoopAllocationSuggestion(allocType string) string {
	return fmt.Sprintf(`Move %s allocation outside the loop or reuse existing allocations:

// Instead of:
for i := 0; i < n; i++ {
    slice := make([]T, size)  // Allocates each iteration
    // use slice...
}

// Do this:
slice := make([]T, size)  // Allocate once
for i := 0; i < n; i++ {
    slice = slice[:0]  // Reset length, keep capacity
    // use slice...
}

Or consider using sync.Pool for frequent allocations.`, allocType)
}

func (v *memoryAllocVisitor) generateCapacitySuggestion() string {
	return `Specify capacity when creating slices with known size:

// Instead of:
slice := make([]T, 0)  // Will grow as needed

// Do this:
slice := make([]T, 0, expectedSize)  // Pre-allocate capacity

This prevents multiple memory allocations and copying during growth.`
}

func (v *memoryAllocVisitor) generateMapSizeSuggestion() string {
	return `Specify initial size for maps when size is predictable:

// Instead of:
m := make(map[string]int)

// Do this:
m := make(map[string]int, expectedSize)

This reduces hash table rehashing and improves performance.`
}

func (v *memoryAllocVisitor) generateAppendSuggestion() string {
	return `Pre-allocate slice capacity to avoid growth in loops:

// Instead of:
var result []T
for _, item := range items {
    result = append(result, process(item))  // Grows each time
}

// Do this:
result := make([]T, 0, len(items))  // Pre-allocate capacity
for _, item := range items {
    result = append(result, process(item))  // No reallocation
}`
}

// createIssue creates a memory allocation issue
func (v *memoryAllocVisitor) createIssue(node ast.Node, message, suggestion string, severity models.Severity) {
	var pos token.Pos
	switch n := node.(type) {
	case *ast.CallExpr:
		pos = n.Pos()
	case *ast.AssignStmt:
		pos = n.Pos()
	default:
		pos = token.NoPos
	}

	position := v.fset.Position(pos)

	issue := models.Issue{
		Type:        models.IssueMemoryAlloc,
		Severity:    severity,
		File:        v.filename,
		Line:        position.Line,
		Column:      position.Column,
		Function:    v.currentFunc,
		Message:     message,
		Suggestion:  suggestion,
		Complexity:  v.getComplexityNote(severity),
		CodeSnippet: position.String(),
	}

	v.issues = append(v.issues, issue)
}

func (v *memoryAllocVisitor) getComplexityNote(severity models.Severity) string {
	switch severity {
	case models.SeverityHigh:
		return "O(n) allocations in loop"
	case models.SeverityMedium:
		return "Potential O(n) growth cost"
	case models.SeverityLow:
		return "Constant factor improvement"
	default:
		return "Memory optimization"
	}
}
