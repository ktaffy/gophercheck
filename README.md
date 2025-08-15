# gophercheck - Go Performance Analyzer CLI Tool

A command-line static analysis tool that detects performance issues in Go code and provides actionable optimization suggestions.

## 🚀 Features

### ✅ Implemented
- **Nested Loop Analysis** - Detects O(n²) and higher complexity patterns
- **String Concatenation Detection** - Finds inefficient string building in loops
- **Cyclomatic Complexity Analysis** - Calculates function complexity scores with thresholds
- **Professional CLI Interface** - Colored console output with emoji indicators
- **JSON Output** - Machine-readable format for CI/CD integration
- **Comprehensive Reporting** - Performance scores and detailed issue descriptions

### 🎯 Performance Issues Detected
1. **Nested Loops** - O(n²), O(n³) complexity patterns with optimization suggestions
2. **String Concatenation** - Inefficient `+=` operations in loops
3. **High Cyclomatic Complexity** - Functions with complexity > 10

## 📦 Installation & Usage

### Quick Start
```bash
# Clone and build
git clone https://github.com/yourusername/gophercheck.git
cd gophercheck
go build -o gophercheck .

# Analyze your code
./gophercheck .                    # Analyze current directory
./gophercheck main.go utils.go     # Analyze specific files
./gophercheck --format=json .      # JSON output for tooling
```

### Sample Output
```
🔍 GopherCheck Analysis Report
═══════════════════════════════════════

📊 Summary:
   Files analyzed: 3
   Issues found: 4

⚠️ Performance Score: 72/100

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
│   │       └── complexity.go
│   └── models/
│       └── issue.go         # Data structures for issues
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
- Linear search patterns

## 🔧 Configuration

### Command Line Options
```bash
gophercheck [flags] [files or directories]

Flags:
  -f, --format string   Output format (console, json) (default "console")
  -w, --watch          Watch mode for development (coming soon)
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

### Phase 2: Enhanced Analysis (Next Weekend)
- [x] **Memory Allocation Detection** - Find unnecessary allocations and suggest optimizations
- [x] **Slice Growth Patterns** - Detect inefficient slice usage and pre-allocation opportunities  
- [x] **Map vs Slice Usage** - Analyze data access patterns and suggest optimal data structures
- [x] **Function Length Analysis** - Flag overly long functions (lines of code threshold)
- [x] **Import Cycle Detection** - Find circular dependencies affecting compilation time

### Phase 3: Advanced Features
- [x] **Watch Mode Implementation** - Real-time analysis during development
- [x] **Configuration File Support** - Custom thresholds and rule configuration
- [ ] **VS Code Extension** - IDE integration with inline suggestions
- [ ] **HTML Report Generation** - Rich web-based reports with charts
- [ ] **Benchmark Integration** - Actual performance measurement suggestions
- [ ] **Git Hook Templates** - Pre-commit and pre-push hook examples

### Phase 4: Professional Polish
- [ ] **Performance Benchmarking** - Measure analyzer performance on large codebases
- [ ] **Error Recovery** - Graceful handling of malformed Go files
- [ ] **Incremental Analysis** - Only analyze changed files for faster CI
- [ ] **Plugin Architecture** - Allow custom detectors via plugins
- [ ] **Machine Learning** - Learn from codebase patterns to reduce false positives

### Additional Detectors to Consider
- [ ] **Database Query Patterns** - N+1 query detection in ORM usage
- [ ] **JSON Marshaling** - Inefficient reflection-based serialization
- [ ] **Regex Compilation** - Repeated regex compilation in loops  
- [ ] **Interface Assertions** - Type assertion performance patterns
- [ ] **Channel Usage** - Unbuffered channel performance issues

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
- Test case expansion
- Documentation improvements
- Cross-platform compatibility testing

## 📄 License

MIT License - see LICENSE file for details.

---

**Built with Go 1.21+ • Uses go/ast for static analysis • Cobra for CLI • No external dependencies for core analysis**