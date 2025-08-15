package analyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"time"

	"gophercheck/internal/analyzer/detectors"
	"gophercheck/internal/models"
)

type Analyzer struct {
	fileSet   *token.FileSet
	detectors []Detector
}

type Detector interface {
	Name() string
	Detect(file *ast.File, fset *token.FileSet, filename string) []models.Issue
}

func NewAnalyzer() *Analyzer {
	analyzer := &Analyzer{
		fileSet: token.NewFileSet(),
	}

	// Initialize all detectors including Phase 2 additions
	analyzer.detectors = []Detector{
		// Original Phase 1 detectors
		detectors.NewNestedLoopDetector(),
		detectors.NewStringConcatDetector(),
		detectors.NewComplexityDetector(),
		detectors.NewMemoryAllocDetector(),

		// Phase 2 detectors
		detectors.NewSliceGrowthDetector(),    // ✅ Slice growth patterns
		detectors.NewDataStructureDetector(),  // ✅ Map vs Slice usage analysis
		detectors.NewFunctionLengthDetector(), // ✅ Function length analysis
		detectors.NewImportCycleDetector(),    // ✅ Import cycle detection
	}

	return analyzer
}

func (a *Analyzer) AnalyzeFiles(filenames []string) (*models.AnalysisResult, error) {
	startTime := time.Now()
	result := models.NewAnalysisResult()

	for _, filename := range filenames {
		issues, err := a.analyzeFile(filename)
		if err != nil {
			// Log error but continue with other files
			continue
		}
		result.Files = append(result.Files, filename)
		for _, issue := range issues {
			result.AddIssue(issue)
		}
	}

	result.AnalysisDuration = time.Since(startTime).String()
	result.CalculateScore()
	return result, nil
}

func (a *Analyzer) analyzeFile(filename string) ([]models.Issue, error) {
	file, err := parser.ParseFile(a.fileSet, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var allIssues []models.Issue
	for _, detector := range a.detectors {
		issues := detector.Detect(file, a.fileSet, filename)
		allIssues = append(allIssues, issues...)
	}

	return allIssues, nil
}

// GetDetectorCount returns the number of active detectors
func (a *Analyzer) GetDetectorCount() int {
	return len(a.detectors)
}

// GetDetectorNames returns the names of all active detectors
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
