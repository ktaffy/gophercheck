package analyzer

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"path/filepath"
	"strings"
	"time"

	"gophercheck/internal/analyzer/detectors"
	"gophercheck/internal/config"
	"gophercheck/internal/context"
	"gophercheck/internal/models"
)

type Analyzer struct {
	fileSet   *token.FileSet
	detectors []Detector
	config    *config.Config
	context   *context.AnalysisContext
}

type Detector interface {
	Name() string
	Detect(file *ast.File, fset *token.FileSet, filename string, ctx *context.AnalysisContext) []models.Issue
}

func NewAnalyzer() *Analyzer {
	return NewAnalyzerWithConfig(config.DefaultConfig())
}

func NewAnalyzerWithConfig(cfg *config.Config) *Analyzer {
	analyzer := &Analyzer{
		fileSet: token.NewFileSet(),
		config:  cfg,
		context: &context.AnalysisContext{
			TypeInfo: &types.Info{
				Types:      make(map[ast.Expr]types.TypeAndValue),
				Defs:       make(map[*ast.Ident]types.Object),
				Uses:       make(map[*ast.Ident]types.Object),
				Selections: make(map[*ast.SelectorExpr]*types.Selection),
				Scopes:     make(map[ast.Node]*types.Scope),
			},
			CallGraph:   make(map[string]*context.CallInfo),
			LoopContext: make(map[ast.Node]*context.LoopInfo),
			DataSizes:   make(map[string]*context.DataSizeInfo),
		},
	}
	// Initialize detectors based on configuration
	analyzer.detectors = []Detector{}

	// Only add detectors that are enabled in config
	if cfg.IsRuleEnabled("nested_loops") {
		detector := detectors.NewNestedLoopDetectorWithConfig(cfg)
		analyzer.detectors = append(analyzer.detectors, detector)
	}

	if cfg.IsRuleEnabled("string_concat") {
		detector := detectors.NewStringConcatDetectorWithConfig(cfg)
		analyzer.detectors = append(analyzer.detectors, detector)
	}

	if cfg.IsRuleEnabled("cyclomatic_complexity") {
		detector := detectors.NewComplexityDetectorWithConfig(cfg)
		analyzer.detectors = append(analyzer.detectors, detector)
	}

	if cfg.IsRuleEnabled("memory_allocation") {
		detector := detectors.NewMemoryAllocDetectorWithConfig(cfg)
		analyzer.detectors = append(analyzer.detectors, detector)
	}

	if cfg.IsRuleEnabled("slice_growth") {
		detector := detectors.NewSliceGrowthDetectorWithConfig(cfg)
		analyzer.detectors = append(analyzer.detectors, detector)
	}

	if cfg.IsRuleEnabled("data_structure") {
		detector := detectors.NewDataStructureDetectorWithConfig(cfg)
		analyzer.detectors = append(analyzer.detectors, detector)
	}

	if cfg.IsRuleEnabled("function_length") {
		detector := detectors.NewFunctionLengthDetectorWithConfig(cfg)
		analyzer.detectors = append(analyzer.detectors, detector)
	}

	if cfg.IsRuleEnabled("import_cycles") {
		detector := detectors.NewImportCycleDetectorWithConfig(cfg)
		analyzer.detectors = append(analyzer.detectors, detector)
	}

	return analyzer
}

func (a *Analyzer) AnalyzeFiles(filenames []string) (*models.AnalysisResult, error) {
	startTime := time.Now()
	var result *models.AnalysisResult
	if a.config != nil {
		result = models.NewAnalysisResultWithConfig(a.config)
	} else {
		result = models.NewAnalysisResult()
	}

	files := make([]*ast.File, 0, len(filenames))
	for _, filename := range filenames {
		file, err := parser.ParseFile(a.fileSet, filename, nil, parser.ParseComments)
		if err != nil {
			continue // Skip files with parse errors
		}
		files = append(files, file)
		result.Files = append(result.Files, filename)
	}

	a.buildTypeInfo(files)

	a.buildAnalysisContext(files)

	for i, file := range files {
		filename := result.Files[i]
		issues := a.analyzeFileWithContext(file, filename)
		for _, issue := range issues {
			result.AddIssue(issue)
		}
	}

	result.AnalysisDuration = time.Since(startTime).String()
	if a.config != nil {
		result.CalculateScoreWithConfig()
	} else {
		result.CalculateScore()
	}

	return result, nil
}

func (a *Analyzer) GetConfig() *config.Config {
	return a.config
}

func (a *Analyzer) analyzeFile(filename string) ([]models.Issue, error) {
	file, err := parser.ParseFile(a.fileSet, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var allIssues []models.Issue
	for _, detector := range a.detectors {
		issues := detector.Detect(file, a.fileSet, filename, a.context)
		allIssues = append(allIssues, issues...)
	}

	return allIssues, nil
}

func (a *Analyzer) GetDetectorCount() int {
	return len(a.detectors)
}

func (a *Analyzer) GetDetectorNames() []string {
	names := make([]string, len(a.detectors))
	for i, detector := range a.detectors {
		names[i] = detector.Name()
	}
	return names
}

type ASTVisitor struct {
	fset     *token.FileSet
	filename string
	issues   []models.Issue
}

func NewASTVisitor(fset *token.FileSet, filename string) *ASTVisitor {
	return &ASTVisitor{
		fset:     fset,
		filename: filepath.Base(filename),
		issues:   make([]models.Issue, 0),
	}
}

func (v *ASTVisitor) AddIssue(issue models.Issue) {
	v.issues = append(v.issues, issue)
}

func (v *ASTVisitor) GetPosition(pos token.Pos) (int, int) {
	position := v.fset.Position(pos)
	return position.Line, position.Column
}

func GetFunctionName(fn *ast.FuncDecl) string {
	if fn.Name != nil {
		return fn.Name.Name
	}
	return "anonymous"
}

func IsInLoop(node ast.Node, parent ast.Node) bool {
	// Simple check - can be made more sophisticated
	switch parent.(type) {
	case *ast.ForStmt, *ast.RangeStmt:
		return true
	default:
		return false
	}
}

// GetCodeSnippet extracts a code snippet around the given position (simplified)
func (v *ASTVisitor) GetCodeSnippet(pos token.Pos, node ast.Node) string {
	position := v.fset.Position(pos)
	// For now, return a simple representation
	// In a full implementation, you'd read the source file and extract lines
	return position.String()
}

func (a *Analyzer) buildTypeInfo(files []*ast.File) {
	typesConfig := &types.Config{
		Importer: importer.ForCompiler(a.fileSet, "source", nil),
		Error: func(err error) {
		},
	}

	typesConfig.Check("", a.fileSet, files, a.context.TypeInfo)
}

func (a *Analyzer) buildAnalysisContext(files []*ast.File) {
	for _, file := range files {
		a.analyzeCallPatterns(file)
		a.analyzeLoopPatterns(file)
		a.analyzeDataSizes(file)
	}
}

func (a *Analyzer) analyzeCallPatterns(file *ast.File) {
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			if node.Name != nil {
				funcName := node.Name.Name
				a.context.CallGraph[funcName] = &context.CallInfo{
					Function:  node,
					CallSites: make([]ast.Node, 0),
					Frequency: a.estimateFrequency(node),
				}
			}
		case *ast.CallExpr:
			if ident, ok := node.Fun.(*ast.Ident); ok {
				if callInfo, exists := a.context.CallGraph[ident.Name]; exists {
					callInfo.CallSites = append(callInfo.CallSites, node)
				}
			}
		}
		return true
	})
}

func (a *Analyzer) analyzeLoopPatterns(file *ast.File) {
	loopDepth := 0

	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.ForStmt:
			loopDepth++
			a.context.LoopContext[node] = &context.LoopInfo{
				LoopNode:     node,
				BoundType:    a.analyzeLoopBounds(node),
				EstimatedMax: a.estimateLoopMax(node),
				IsInnerLoop:  loopDepth > 1,
				HasEarlyExit: a.hasEarlyExit(node),
			}

		case *ast.RangeStmt:
			loopDepth++
			a.context.LoopContext[node] = &context.LoopInfo{
				LoopNode:     node,
				BoundType:    context.BoundLinear, // Range is always linear
				EstimatedMax: a.estimateRangeMax(node),
				IsInnerLoop:  loopDepth > 1,
				HasEarlyExit: a.hasEarlyExit(node),
			}
		}
		return true
	})
}

func (a *Analyzer) analyzeDataSizes(file *ast.File) {
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.CompositeLit:
			if arrayType, ok := node.Type.(*ast.ArrayType); ok && arrayType.Len == nil {
				size := len(node.Elts)
				if varName := a.getVariableFromAssignment(node); varName != "" {
					a.context.DataSizes[varName] = &context.DataSizeInfo{
						EstimatedLen: size,
						Confidence:   1.0,
						Source:       "literal",
					}
				}
			}
		case *ast.CallExpr:
			if a.isMakeCall(node) && len(node.Args) >= 2 {
				if size := a.extractConstantInt(node.Args[1]); size > 0 {
					if varName := a.getVariableFromAssignment(node); varName != "" {
						a.context.DataSizes[varName] = &context.DataSizeInfo{
							EstimatedLen: size,
							Confidence:   0.8,
							Source:       "make",
						}
					}
				}
			}
		}
		return true
	})
}

func (a *Analyzer) analyzeFileWithContext(file *ast.File, filename string) []models.Issue {
	var allIssues []models.Issue
	for _, detector := range a.detectors {
		// This will have compiler errors until we fix the detectors
		issues := detector.Detect(file, a.fileSet, filename, a.context)
		allIssues = append(allIssues, issues...)
	}
	return allIssues
}

func (a *Analyzer) estimateFrequency(fn *ast.FuncDecl) context.FrequencyEstimate {
	if fn.Name == nil {
		return context.FrequencyUnknown
	}

	name := strings.ToLower(fn.Name.Name)

	if strings.Contains(name, "error") || strings.Contains(name, "panic") {
		return context.FrequencyRare
	}

	if strings.HasPrefix(name, "init") || strings.HasPrefix(name, "setup") {
		return context.FrequencyRare
	}

	if strings.Contains(name, "process") || strings.Contains(name, "handle") ||
		strings.Contains(name, "loop") || strings.Contains(name, "iterate") {
		return context.FrequencyHigh
	}

	return context.FrequencyModerate
}

func (a *Analyzer) analyzeLoopBounds(loop *ast.ForStmt) context.LoopBoundType {
	if loop.Cond == nil {
		return context.BoundUnknown
	}

	switch cond := loop.Cond.(type) {
	case *ast.BinaryExpr:
		if a.isConstantExpression(cond.Y) {
			return context.BoundConstant
		}
		if a.isLenExpression(cond.Y) {
			return context.BoundLinear
		}
	}

	return context.BoundVariable
}

func (a *Analyzer) estimateLoopMax(loop *ast.ForStmt) int {
	if loop.Cond == nil {
		return -1
	}

	switch cond := loop.Cond.(type) {
	case *ast.BinaryExpr:
		if val := a.extractConstantInt(cond.Y); val > 0 {
			return val
		}
	}

	return -1
}

func (a *Analyzer) estimateRangeMax(loop *ast.RangeStmt) int {
	if ident, ok := loop.X.(*ast.Ident); ok {
		if sizeInfo, exists := a.context.DataSizes[ident.Name]; exists {
			return sizeInfo.EstimatedLen
		}
	}
	return -1
}

func (a *Analyzer) hasEarlyExit(node ast.Node) bool {
	hasExit := false
	ast.Inspect(node, func(n ast.Node) bool {
		switch stmt := n.(type) {
		case *ast.BranchStmt:
			if stmt.Tok == token.BREAK || stmt.Tok == token.RETURN {
				hasExit = true
				return false
			}
		case *ast.ReturnStmt:
			hasExit = true
			return false
		}
		return true
	})
	return hasExit
}

func (a *Analyzer) isConstantExpression(expr ast.Expr) bool {
	switch expr.(type) {
	case *ast.BasicLit:
		return true
	}
	return false
}

func (a *Analyzer) isLenExpression(expr ast.Expr) bool {
	if call, ok := expr.(*ast.CallExpr); ok {
		if ident, ok := call.Fun.(*ast.Ident); ok {
			return ident.Name == "len"
		}
	}
	return false
}

func (a *Analyzer) isMakeCall(call *ast.CallExpr) bool {
	if ident, ok := call.Fun.(*ast.Ident); ok {
		return ident.Name == "make"
	}
	return false
}

func (a *Analyzer) extractConstantInt(expr ast.Expr) int {
	if lit, ok := expr.(*ast.BasicLit); ok && lit.Kind == token.INT {
		switch lit.Value {
		case "0":
			return 0
		case "1":
			return 1
		case "2":
			return 2
		case "5":
			return 5
		case "10":
			return 10
		case "100":
			return 100
		case "1000":
			return 1000
		}
	}
	return -1
}

func (a *Analyzer) getVariableFromAssignment(expr ast.Expr) string {
	// This is a simplified version - in reality this would traverse up
	// the AST to find the assignment target
	// For now, return empty string
	return ""
}
