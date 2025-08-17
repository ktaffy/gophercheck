package detectors

import (
	"fmt"
	"go/ast"
	"go/token"
	"gophercheck/internal/config"
	"gophercheck/internal/context"
	"gophercheck/internal/models"
	"path"
	"strings"
)

type ImportCycleDetector struct {
	packages map[string]*packageInfo
	analyzed map[string]bool
	config   *config.Config
}

func NewImportCycleDetector() *ImportCycleDetector {
	return &ImportCycleDetector{
		packages: make(map[string]*packageInfo),
		analyzed: make(map[string]bool),
	}
}

func NewImportCycleDetectorWithConfig(cfg *config.Config) *ImportCycleDetector {
	return &ImportCycleDetector{
		packages: make(map[string]*packageInfo),
		analyzed: make(map[string]bool),
		config:   cfg,
	}
}

func (d *ImportCycleDetector) SetConfig(cfg *config.Config) {
	d.config = cfg
}

func (d *ImportCycleDetector) Name() string {
	return "Import Cycle Detector"
}

type packageInfo struct {
	name     string
	filePath string
	imports  []string
	line     int
}

func (d *ImportCycleDetector) Detect(file *ast.File, fset *token.FileSet, filename string, ctx *context.AnalysisContext) []models.Issue {
	detector := &importCycleVisitor{
		detector: d,
		fset:     fset,
		filename: filename,
		issues:   make([]models.Issue, 0),
		context:  ctx,
	}

	ast.Walk(detector, file)

	// After collecting all package info, analyze for cycles
	cycles := d.findCycles()
	for _, cycle := range cycles {
		detector.createCycleIssue(cycle)
	}

	return detector.issues
}

type importCycleVisitor struct {
	detector    *ImportCycleDetector
	fset        *token.FileSet
	filename    string
	issues      []models.Issue
	packageName string
	context     *context.AnalysisContext
}

func (v *importCycleVisitor) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.File:
		if n.Name != nil {
			v.packageName = n.Name.Name
		}
		return v

	case *ast.GenDecl:
		if n.Tok == token.IMPORT {
			v.processImports(n)
		}
		return nil

	default:
		return v
	}
}

func (v *importCycleVisitor) processImports(decl *ast.GenDecl) {
	var imports []string
	importLine := 0

	for _, spec := range decl.Specs {
		if importSpec, ok := spec.(*ast.ImportSpec); ok {
			if importSpec.Path != nil {
				importPath := strings.Trim(importSpec.Path.Value, `"`)

				if !v.isThirdPartyOrLocalImport(importPath) {
					continue
				}

				imports = append(imports, importPath)

				if importLine == 0 {
					position := v.fset.Position(importSpec.Pos())
					importLine = position.Line
				}
			}
		}
	}

	if len(imports) > 0 {
		// Extract package name from file path
		packagePath := v.getPackagePathFromFile(v.filename)

		v.detector.packages[packagePath] = &packageInfo{
			name:     v.packageName,
			filePath: v.filename,
			imports:  imports,
			line:     importLine,
		}
	}
}

func (v *importCycleVisitor) isThirdPartyOrLocalImport(importPath string) bool {
	if v.detector.config != nil && v.detector.config.Rules.Quality.ImportCycles.Enabled {
		for _, excluded := range v.detector.config.Rules.Quality.ImportCycles.ExcludePackages {
			if importPath == excluded || strings.HasPrefix(importPath, excluded+"/") {
				return false
			}
		}

		if v.detector.config.Rules.Quality.ImportCycles.IgnoreVendor {
			if strings.HasPrefix(importPath, "vendor/") || strings.Contains(importPath, "/vendor/") {
				return false
			}
		}
	}

	stdLibPrefixes := []string{
		"fmt", "os", "io", "net", "http", "time", "strings", "strconv",
		"context", "sync", "encoding", "crypto", "database", "archive",
		"bufio", "bytes", "compress", "container", "debug", "embed",
		"errors", "expvar", "flag", "go", "hash", "html", "image",
		"index", "log", "math", "mime", "path", "plugin", "reflect",
		"regexp", "runtime", "sort", "syscall", "testing", "text",
		"unicode", "unsafe",
	}

	for _, prefix := range stdLibPrefixes {
		if importPath == prefix || strings.HasPrefix(importPath, prefix+"/") {
			return false
		}
	}

	return strings.Contains(importPath, ".") || strings.HasPrefix(importPath, "./") || strings.HasPrefix(importPath, "../")
}

func (v *importCycleVisitor) getPackagePathFromFile(filename string) string {
	// Extract package path from file path
	// This is simplified - in a real implementation, you'd use go/build or go/packages
	dir := path.Dir(filename)
	if dir == "." {
		return "main"
	}
	return dir
}

func (d *ImportCycleDetector) findCycles() [][]string {
	var cycles [][]string
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	path := make([]string, 0)

	for packagePath := range d.packages {
		if !visited[packagePath] {
			if cycle := d.dfs(packagePath, visited, recStack, path); cycle != nil {
				cycles = append(cycles, cycle)
			}
		}
	}

	return cycles
}

func (d *ImportCycleDetector) dfs(packagePath string, visited, recStack map[string]bool, path []string) []string {
	visited[packagePath] = true
	recStack[packagePath] = true
	path = append(path, packagePath)

	pkg, exists := d.packages[packagePath]
	if !exists {
		recStack[packagePath] = false
		return nil
	}

	for _, importPath := range pkg.imports {
		// Normalize import path to package path
		normalizedPath := d.normalizeImportPath(importPath)

		if !visited[normalizedPath] {
			if cycle := d.dfs(normalizedPath, visited, recStack, path); cycle != nil {
				recStack[packagePath] = false
				return cycle
			}
		} else if recStack[normalizedPath] {
			// Found a cycle - extract the cycle from the path
			cycle := d.extractCycle(path, normalizedPath)
			recStack[packagePath] = false
			return cycle
		}
	}

	recStack[packagePath] = false
	return nil
}

func (d *ImportCycleDetector) normalizeImportPath(importPath string) string {
	// Simplified normalization - in a real implementation, this would be more sophisticated
	if strings.HasPrefix(importPath, "./") || strings.HasPrefix(importPath, "../") {
		return path.Clean(importPath)
	}
	return importPath
}

func (d *ImportCycleDetector) extractCycle(path []string, cycleStart string) []string {
	// Find where the cycle starts and extract it
	for i, pkg := range path {
		if pkg == cycleStart {
			cycle := make([]string, len(path)-i)
			copy(cycle, path[i:])
			cycle = append(cycle, cycleStart) // Complete the cycle
			return cycle
		}
	}
	return path
}

func (v *importCycleVisitor) createCycleIssue(cycle []string) {
	if len(cycle) < 2 {
		return
	}

	// Check config settings
	if v.detector.config != nil && v.detector.config.Rules.Quality.ImportCycles.Enabled {
		cycleLen := len(cycle) - 1 // Remove duplicate at end
		maxCycleLength := v.detector.config.Rules.Quality.ImportCycles.MaxCycleLength

		// Don't report cycles that are within acceptable limits
		if cycleLen <= maxCycleLength {
			return
		}

		// Check if test packages should be ignored
		if v.detector.config.Rules.Quality.ImportCycles.IgnoreTestPackages {
			// Skip if any package in cycle appears to be a test package
			for _, pkg := range cycle {
				if strings.Contains(pkg, "_test") || strings.Contains(pkg, "/test") {
					return
				}
			}
		}
	}

	// Find the package in our current file that's part of the cycle
	currentPackage := v.getPackagePathFromFile(v.filename)

	// Check if current file is part of this cycle
	inCycle := false
	for _, pkg := range cycle {
		if pkg == currentPackage {
			inCycle = true
			break
		}
	}

	if !inCycle {
		return
	}

	// Get the package info for line number
	pkgInfo := v.detector.packages[currentPackage]
	line := 1
	if pkgInfo != nil {
		line = pkgInfo.line
	}

	cycleStr := strings.Join(cycle, " → ")

	issue := models.Issue{
		Type:        models.IssueImportCycle,
		Severity:    v.calculateCycleSeverity(len(cycle)),
		File:        v.filename,
		Line:        line,
		Column:      1,
		Function:    "", // Not applicable for import issues
		Message:     fmt.Sprintf("Import cycle detected: %s", cycleStr),
		Suggestion:  v.generateCycleSuggestion(cycle),
		Complexity:  fmt.Sprintf("Cycle length: %d packages", len(cycle)-1),
		CodeSnippet: fmt.Sprintf("%s:%d", v.filename, line),
	}

	v.issues = append(v.issues, issue)
}

func (v *importCycleVisitor) calculateCycleSeverity(cycleLength int) models.Severity {
	maxCycleLength := 5 // default
	if v.detector.config != nil && v.detector.config.Rules.Quality.ImportCycles.Enabled {
		maxCycleLength = v.detector.config.Rules.Quality.ImportCycles.MaxCycleLength
	}

	ratio := float64(cycleLength) / float64(maxCycleLength)

	switch {
	case ratio >= 1.5: // 150% of max = critical
		return models.SeverityCritical
	case ratio >= 1.2: // 120% of max = high
		return models.SeverityHigh
	case ratio >= 1.0: // 100% of max = medium
		return models.SeverityMedium
	default:
		return models.SeverityLow
	}
}

func (v *importCycleVisitor) generateCycleSuggestion(cycle []string) string {
	cycleLen := len(cycle) - 1 // Remove duplicate at end

	baseAdvice := `Import cycles prevent compilation and indicate poor package design. Here are strategies to break the cycle:

1. **Dependency Inversion**: Create interfaces to break direct dependencies
2. **Extract Common Code**: Move shared functionality to a separate package
3. **Merge Packages**: If packages are tightly coupled, consider combining them
4. **Remove Unnecessary Dependencies**: Review if all imports are actually needed`
	switch {
	case cycleLen == 2:
		return baseAdvice + `

For 2-package cycles:
// Instead of:
// package A imports B
// package B imports A

// Strategy 1 - Extract interface:
// package A imports B (interface only)
// package B imports common
// package common defines interfaces

// Strategy 2 - Dependency injection:
// package A defines interface, imports B
// package B implements interface, no import of A
// main wires them together`

	case cycleLen == 3:
		return baseAdvice + `

For 3-package cycles (A → B → C → A):
1. **Find the weakest link**: Identify which dependency is least essential
2. **Extract shared interfaces**: Create a common package for shared contracts
3. **Use event-driven design**: Replace direct calls with event publishing
4. **Consider package merging**: If A, B, C are tightly coupled, merge them

Example refactoring:
// Before: A → B → C → A
// After:  A → common ← B ← C
//    common contains interfaces used by all`

	default:
		return baseAdvice + fmt.Sprintf(`

Complex %d-package cycle requires architectural review:
1. **Draw dependency diagram** to visualize the cycle
2. **Identify core domain concepts** that shouldn't depend on periphery
3. **Apply Clean Architecture principles** (domain → application → infrastructure)
4. **Consider microservices** if packages represent different bounded contexts
5. **Use dependency injection container** to manage complex relationships

This cycle suggests the codebase may need significant restructuring.`, cycleLen)
	}
}
