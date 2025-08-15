# gophercheck - Go Performance Analyzer CLI Tool

A command-line static analysis tool that detects performance issues in Go code and provides actionable optimization suggestions.

## Project Goals

**Primary Goal**: Build an impressive resume project that demonstrates:
- Go expertise (AST parsing, static analysis)
- Data Structures & Algorithms knowledge (tree traversal, complexity analysis)
- Real-world developer tooling experience
- Clean CLI interface design

**Timeline**: 1-2 weekends maximum
**Target**: Functional MVP that showcases technical skills

## Features

### Core Detection Capabilities
1. **Nested Loop Analysis** - Detect O(n²) complexity patterns
2. **String Concatenation Issues** - Find inefficient string building in loops
3. **Inefficient Data Structure Usage** - Suggest map vs slice optimizations
4. **Cyclomatic Complexity** - Calculate function complexity scores
5. **Memory Allocation Hotspots** - Detect unnecessary allocations

### CLI Interface
```bash
# Basic usage
gophercheck .                    # Analyze current directory
gophercheck main.go utils.go     # Analyze specific files
gophercheck --watch .            # Watch mode for development
gophercheck --format=json .      # JSON output for CI/CD integration
```

### Output Format
- **Visual Indicators**: Colors, emojis, clear severity levels
- **Actionable Suggestions**: Specific code improvements with examples
- **Complexity Reports**: Table format showing function risk levels
- **Performance Score**: Overall codebase health metric

## Technical Architecture

### Core Components
1. **AST Parser** - Go's `go/ast` package for code analysis
2. **Pattern Detectors** - Individual analyzers for each performance issue
3. **Complexity Calculator** - Algorithms to estimate Big O complexity
4. **Report Generator** - Format and display results
5. **CLI Handler** - Command parsing and file management

### Key Algorithms
- **Tree Traversal**: DFS through AST nodes
- **Pattern Matching**: Detect specific code anti-patterns
- **Complexity Analysis**: Calculate cyclomatic complexity
- **Graph Analysis**: Function call dependency tracking

### Tech Stack
- **Language**: Go (showcases Go expertise)
- **Libraries**: 
  - `go/ast` - AST parsing
  - `go/parser` - Code parsing
  - `cobra` - CLI framework
  - `color` - Terminal colors
- **Testing**: Go's built-in testing framework

## Project Structure
```
gophercheck/
├── cmd/
│   └── root.go              # CLI commands
├── internal/
│   ├── analyzer/
│   │   ├── ast_walker.go    # AST traversal logic
│   │   ├── detectors/       # Individual performance detectors
│   │   │   ├── nested_loops.go
│   │   │   ├── string_concat.go
│   │   │   └── complexity.go
│   │   └── report.go        # Result formatting
│   └── models/
│       └── issue.go         # Issue data structures
├── testdata/                # Sample Go files for testing
├── main.go
├── go.mod
└── README.md
```

## Success Metrics

### For Resume Impact
- Clean, professional GitHub repository
- Comprehensive README with usage examples
- Demonstrates multiple technical skills
- Shows practical problem-solving ability

### Technical Goals
- Handle real Go codebases without crashing
- Fast analysis (< 2 seconds for medium files)
- Accurate detection with minimal false positives
- Clear, actionable output

## Sample Performance Issues to Detect

### 1. Nested Loops (O(n²))
```go
// BAD
for i := range users {
    for j := range posts {
        if posts[j].UserID == users[i].ID {
            // process
        }
    }
}

// Suggestion: Use map[UserID][]Post for O(n) lookup
```

### 2. String Concatenation in Loops
```go
// BAD
var result string
for _, item := range items {
    result += item
}

// Suggestion: Use strings.Builder for O(n) performance
```

### 3. Inefficient Slice Operations
```go
// BAD
func contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false
}

// Suggestion: Use map[string]bool for O(1) lookup
```

## Development Phases

### Phase 1: Core Engine (Weekend 1)
- AST parsing and traversal
- Basic nested loop detection
- Simple CLI interface
- Console output

### Phase 2: Enhanced Analysis (Weekend 2)
- Additional performance patterns
- Complexity scoring
- Formatted output with colors
- File watching capability

### Phase 3: Polish (Optional)
- JSON output for tooling integration
- More sophisticated pattern detection
- Performance optimizations
- Comprehensive testing

## Demo Script

1. **Show the tool in action** on a sample Go file with obvious performance issues
2. **Highlight specific detections** and suggestions
3. **Demonstrate watch mode** for development workflow
4. **Show JSON output** for CI/CD integration potential

## Key Selling Points for Resume

- "Built a static analysis tool using Go's AST package to detect O(n²) algorithms"
- "Implemented tree traversal algorithms for code complexity analysis"
- "Created professional CLI tool with actionable performance suggestions"
- "Showcases understanding of algorithmic complexity and Go optimization patterns"