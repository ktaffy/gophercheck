# gophercheck - Go Performance Analyzer CLI Tool

A command-line static analysis tool that detects performance issues in Go code and provides actionable optimization suggestions.

## 🚀 Features

### ✅ **FULLY IMPLEMENTED (Current State)**
- **Nested Loop Analysis** - Detects O(n²) and higher complexity patterns with configurable depth thresholds
- **String Concatenation Detection** - Finds inefficient string building in loops with smart variable name detection
- **Cyclomatic Complexity Analysis** - Function complexity scoring with configurable thresholds (10/15/25 default)
- **Memory Allocation Detection** - Identifies unnecessary allocations in loops and missing capacity hints
- **Slice Growth Pattern Analysis** - Detects inefficient slice usage and pre-allocation opportunities
- **Data Structure Usage Analysis** - Identifies linear searches that should use maps for O(1) lookups
- **Function Length Analysis** - Flags overly long functions with configurable line thresholds (50/100/200)
- **Import Cycle Detection** - Finds circular dependencies affecting compilation performance
- **Watch Mode** - Real-time analysis during development with file change detection and debouncing
- **Configuration System** - Comprehensive YAML-based config with rule customization and thresholds
- **Professional CLI Interface** - Colored console output with emoji indicators and multiple formats
- **JSON Output** - Machine-readable format for CI/CD integration
- **Performance Scoring** - 0-100 scale scoring system with severity-weighted penalties

### 🎯 **Performance Issues Detected (8 Detector Types)**
1. **Nested Loops** - O(n²), O(n³) complexity patterns with optimization suggestions
2. **String Concatenation** - Inefficient `+=` operations in loops
3. **High Cyclomatic Complexity** - Functions exceeding complexity thresholds
4. **Memory Allocation** - Allocations in loops, missing capacity hints for slices/maps
5. **Slice Growth Patterns** - Inefficient slice creation and append operations
6. **Inefficient Data Structures** - Linear searches where maps would be O(1)
7. **Function Length** - Overly long functions affecting maintainability
8. **Import Cycles** - Circular package dependencies

## 📦 Installation & Usage

### Quick Start
```bash
# Clone and build
git clone https://github.com/yourusername/gophercheck.git
cd gophercheck
go build -o gophercheck .

# Analyze your code
./gophercheck .                            # Analyze current directory
./gophercheck main.go utils.go             # Analyze specific files
./gophercheck --format=json .              # JSON output for tooling
./gophercheck --config .gophercheck.yml .  # Use custom config
./gophercheck --watch .                    # Watch mode - analyze on file changes
./gophercheck --generate-config            # Generate sample config file
```

### Sample Output
```
🔍 GopherCheck Analysis Report
═══════════════════════════════════════

📊 Summary:
   Files analyzed: 3
   Issues found: 4

⚡ Performance Score: 72/100

📋 Issues by Severity:
   ❌ HIGH: 1
   ⚠️ MEDIUM: 3

🔍 Detailed Issues:
──────────────────────────────────────────────────

❌ Issue #1 - HIGH CYCLOMATIC_COMPLEXITY
   📍 Location: main.go:25:1 in function 'ComplexFunction'
   💭 Issue: Function 'ComplexFunction' has high cyclomatic complexity: 16
   📊 Complexity: Complexity: 16
   💡 Suggestion:
      Consider breaking this function into smaller, single-purpose functions
      Extract complex conditional logic into separate functions
```

## 🏗️ Technical Architecture

### Core Components
- **AST Parser** - Uses Go's `go/ast` package for syntax tree analysis
- **Pattern Detectors** - Modular analyzers implementing visitor pattern
- **Complexity Calculator** - Graph-based cyclomatic complexity analysis
- **Report Generator** - Formatted console and JSON output
- **CLI Framework** - Built with Cobra for professional UX

### Project Structure
```
gophercheck/
├── cmd/
│   └── root.go              # CLI commands and argument parsing
├── internal/
│   ├── analyzer/
│   │   ├── ast_walker.go    # Core AST traversal engine
│   │   ├── report.go        # Output formatting and display
│   │   └── detectors/       # Performance issue detectors
│   │       ├── nested_loops.go
│   │       ├── string_concat.go
│   │       ├── complexity.go
│   │       ├── memory_alloc.go
│   │       ├── slice_growth.go
│   │       ├── data_structure.go
│   │       ├── function_length.go
│   │       └── import_cycle.go
│   ├── config/
│   │   └── config.go        # YAML configuration system
│   ├── models/
│   │   └── issue.go         # Data structures for issues
│   └── watcher/
│       ├── file_watcher.go  # File system monitoring
│       └── debouncer.go     # Change event debouncing
├── testdata/
│   └── sample.go           # Test files with performance issues
├── main.go
└── README.md
```

### Key Algorithms
- **Tree Traversal** - Depth-first search through AST nodes
- **Pattern Matching** - Detection of specific anti-performance patterns
- **Complexity Calculation** - McCabe cyclomatic complexity metrics
- **Severity Assessment** - Risk-based issue prioritization

## 🧪 Testing

Test the tool on the included sample file:
```bash
./gophercheck testdata/sample.go
```

The sample includes intentional performance issues:
- Nested loops with O(n²) complexity
- String concatenation in loops
- High cyclomatic complexity function
- Memory allocation inefficiencies
- Slice growth without pre-allocation
- Linear search patterns
- Import cycle examples
- Overly long functions (200+ lines)

## 🔧 Configuration

### Command Line Options
```bash
gophercheck [flags] [files or directories]

Flags:
  -f, --format string   Output format (console, json) (default "console")
  -w, --watch          Watch mode for development
  -c, --config string  Path to configuration file
      --generate-config Generate sample configuration file
  -h, --help           Help for gophercheck
```

### CI/CD Integration
```yaml
# GitHub Actions example
- name: Performance Analysis
  run: |
    go install github.com/yourusername/gophercheck@latest
    gophercheck --format=json . > performance-report.json
    
- name: Check Performance Score
  run: |
    score=$(jq '.performance_score' performance-report.json)
    if [ $score -lt 80 ]; then
      echo "Performance score too low: $score"
      exit 1
    fi
```

## 📈 Roadmap - What to Implement Next

### 🎯 **Phase 3: CLI Polish & Enhanced Detection (Current Focus)**
- [x] **Enhanced CLI UX** - Better progress indicators, improved error messages, help system
- [ ] **Algorithm Improvements** - More sophisticated pattern detection, reduced false positives
- [ ] **New Detectors** - Regex compilation, interface assertions, channel usage patterns
- [ ] **Better Suggestions** - More specific, actionable recommendations with code examples
- [ ] **Error Recovery** - Graceful handling of malformed Go files
- [ ] **Performance Optimization** - Faster analysis on large codebases

### 📊 **Planned New Detectors**
- [ ] **Regex Compilation** - Repeated regex compilation in loops  
- [ ] **Interface Assertions** - Type assertion performance patterns
- [ ] **Channel Usage** - Unbuffered channel performance issues
- [ ] **JSON Marshaling** - Inefficient reflection-based serialization
- [ ] **Database Query Patterns** - N+1 query detection in ORM usage
- [ ] **HTTP Client Patterns** - Connection reuse and timeout issues
- [ ] **Goroutine Leak Detection** - Identify potential goroutine leaks
- [ ] **Context Usage** - Missing context.Context in long-running operations

### 🚀 **Phase 4: Advanced Features (Future)**
- [ ] **Incremental Analysis** - Only analyze changed files for faster CI
- [ ] **Plugin Architecture** - Allow custom detectors via plugins
- [ ] **Machine Learning** - Learn from codebase patterns to reduce false positives
- [ ] **Benchmark Integration** - Actual performance measurement suggestions

### 🌐 **Phase 5: External Integration (Later)**
- [ ] **HTML Report Generation** - Rich web-based reports with charts
- [ ] **VS Code Extension** - IDE integration with inline suggestions
- [ ] **Git Hook Templates** - Pre-commit and pre-push hook examples

## 🎯 Resume Highlights

This project demonstrates:
- **Go Expertise** - Deep knowledge of AST manipulation and Go internals
- **Algorithms & Data Structures** - Tree traversal, complexity analysis, pattern matching
- **Software Engineering** - Clean architecture, modular design, professional tooling
- **DevOps Integration** - CI/CD ready with JSON output and automation support
- **Problem Solving** - Real-world developer productivity improvements

## 🤝 Contributing

Contributions welcome! Areas needing help:
- Additional performance pattern detectors
- Algorithm improvements for existing detectors
- CLI user experience enhancements
- Test case expansion
- Documentation improvements

## 📄 License

MIT License - see LICENSE file for details.

---

**Built with Go 1.21+ • Uses go/ast for static analysis • Cobra for CLI • Professional developer tooling focus**