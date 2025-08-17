package detectors

import (
	"fmt"
	"go/ast"
	"go/token"

	"gophercheck/internal/config"
	"gophercheck/internal/context"
	"gophercheck/internal/models"
)

type NestedLoopDetector struct {
	config *config.Config
}

func NewNestedLoopDetector() *NestedLoopDetector {
	return &NestedLoopDetector{}
}

func NewNestedLoopDetectorWithConfig(cfg *config.Config) *NestedLoopDetector {
	return &NestedLoopDetector{
		config: cfg,
	}
}

func (d *NestedLoopDetector) SetConfig(cfg *config.Config) {
	d.config = cfg
}

func (d *NestedLoopDetector) Name() string {
	return "Nested Loop Detector"
}

func (d *NestedLoopDetector) Detect(file *ast.File, fset *token.FileSet, filename string, ctx *context.AnalysisContext) []models.Issue {
	detector := &nestedLoopVisitor{
		fset:     fset,
		filename: filename,
		issues:   make([]models.Issue, 0),
		detector: d,
		context:  ctx,
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
	detector    *NestedLoopDetector
	context     *context.AnalysisContext
}

func (v *nestedLoopVisitor) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.FuncDecl:
		if n.Name != nil {
			v.currentFunc = n.Name.Name
		}
		return v
	case *ast.ForStmt, *ast.RangeStmt:
		v.loopDepth++
		maxDepth := 1
		if v.detector.config != nil && v.detector.config.Rules.Performance.NestedLoops.Enabled {
			maxDepth = v.detector.config.Rules.Performance.NestedLoops.MaxDepth
		}
		if v.loopDepth > maxDepth {
			v.detectNestedLoop(n)
		}
		for _, child := range getLoopBody(n) {
			ast.Walk(v, child)
		}

		v.loopDepth--
		return nil
	default:
		return v
	}
}

func (v *nestedLoopVisitor) detectNestedLoop(node ast.Node) {
	loopInfo, hasInfo := v.context.LoopContext[node]

	if hasInfo && v.shouldSkipSmallLoop(loopInfo) {
		return
	}

	pos := getNodePosition(node)
	position := v.fset.Position(pos)

	confidence := v.calculateConfidence(loopInfo, hasInfo)

	if confidence < 0.6 {
		return
	}

	issue := models.Issue{
		Type:        models.IssueNestedLoops,
		Severity:    v.calculateSeverityWithContext(loopInfo, hasInfo),
		File:        v.filename,
		Line:        position.Line,
		Column:      position.Column,
		Function:    v.currentFunc,
		Message:     v.generateContextualMessage(loopInfo, hasInfo),
		Suggestion:  v.generateContextualSuggestion(loopInfo, hasInfo),
		Complexity:  v.generateComplexityInfo(loopInfo, hasInfo),
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

func (v *nestedLoopVisitor) shouldSkipSmallLoop(loopInfo *context.LoopInfo) bool {
	if loopInfo == nil {
		return false
	}

	if loopInfo.BoundType == context.BoundConstant && loopInfo.EstimatedMax > 0 && loopInfo.EstimatedMax <= 10 {
		return true
	}

	if loopInfo.HasEarlyExit {
		return true
	}

	return false
}

func (v *nestedLoopVisitor) calculateConfidence(loopInfo *context.LoopInfo, hasInfo bool) float64 {
	if !hasInfo {
		return 0.5 // Medium confidence if we don't have context
	}

	confidence := 0.8 // Base confidence

	if loopInfo.BoundType == context.BoundVariable ||
		(loopInfo.EstimatedMax > 100) {
		confidence += 0.2
	}

	if loopInfo.HasEarlyExit {
		confidence -= 0.3
	}

	if v.loopDepth >= 3 {
		confidence += 0.1
	}

	return min(confidence, 1.0)
}

func (v *nestedLoopVisitor) calculateSeverityWithContext(loopInfo *context.LoopInfo, hasInfo bool) models.Severity {
	baseSeverity := v.calculateSeverity() // Original method

	if !hasInfo {
		return baseSeverity
	}

	if loopInfo.BoundType == context.BoundConstant && loopInfo.EstimatedMax <= 50 {
		if baseSeverity == models.SeverityCritical {
			return models.SeverityHigh
		}
		if baseSeverity == models.SeverityHigh {
			return models.SeverityMedium
		}
	}

	if loopInfo.EstimatedMax > 1000 {
		if baseSeverity == models.SeverityMedium {
			return models.SeverityHigh
		}
		if baseSeverity == models.SeverityHigh {
			return models.SeverityCritical
		}
	}

	return baseSeverity
}

func (v *nestedLoopVisitor) generateContextualMessage(loopInfo *context.LoopInfo, hasInfo bool) string {
	baseMsg := v.generateMessage() // Original method

	if !hasInfo {
		return baseMsg
	}

	switch loopInfo.BoundType {
	case context.BoundConstant:
		if loopInfo.EstimatedMax > 0 {
			return fmt.Sprintf("%s (processing ~%d×%d = %d operations)",
				baseMsg, loopInfo.EstimatedMax, loopInfo.EstimatedMax,
				loopInfo.EstimatedMax*loopInfo.EstimatedMax)
		}
	case context.BoundLinear:
		return fmt.Sprintf("%s (complexity depends on input size - could be O(n²))", baseMsg)
	case context.BoundVariable:
		return fmt.Sprintf("%s (unbounded - potentially very expensive)", baseMsg)
	}

	return baseMsg
}

func (v *nestedLoopVisitor) generateContextualSuggestion(loopInfo *context.LoopInfo, hasInfo bool) string {
	baseSuggestion := v.generateSuggestion() // Original method

	if !hasInfo {
		return baseSuggestion
	}

	var contextSuggestion string

	switch {
	case loopInfo.BoundType == context.BoundConstant && loopInfo.EstimatedMax <= 100:
		contextSuggestion = "\n\nNote: Since this involves small, bounded loops (~" +
			fmt.Sprintf("%d", loopInfo.EstimatedMax) + " iterations), " +
			"the performance impact may be acceptable. Consider profiling to confirm."

	case loopInfo.HasEarlyExit:
		contextSuggestion = "\n\nDetected early exit pattern - this might be optimized search logic. " +
			"Consider: 1) Use a map for O(1) lookups, 2) Sort data and use binary search, " +
			"3) Break outer loop when inner condition is met."

	case loopInfo.BoundType == context.BoundLinear:
		contextSuggestion = "\n\nThis appears to iterate over data structures. " +
			"Consider: 1) Pre-processing data into a map, 2) Using a single loop with smarter logic, " +
			"3) Algorithm change (sort + merge vs nested iteration)."

	case v.loopDepth >= 3:
		contextSuggestion = "\n\nCRITICAL: Triple-nested loops often indicate algorithmic issues. " +
			"This likely needs a complete algorithmic redesign, not just optimization."
	}

	return baseSuggestion + contextSuggestion
}

func (v *nestedLoopVisitor) generateComplexityInfo(loopInfo *context.LoopInfo, hasInfo bool) string {
	baseComplexity := fmt.Sprintf("O(n^%d)", v.loopDepth)

	if !hasInfo {
		return baseComplexity
	}

	if loopInfo.BoundType == context.BoundConstant && loopInfo.EstimatedMax > 0 {
		return fmt.Sprintf("O(%d) - constant time with ~%d operations",
			loopInfo.EstimatedMax*loopInfo.EstimatedMax, loopInfo.EstimatedMax*loopInfo.EstimatedMax)
	}

	return baseComplexity
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
