package context

import (
	"go/ast"
	"go/types"
)

// AnalysisContext provides rich analysis context to detectors
type AnalysisContext struct {
	TypeInfo    *types.Info
	CallGraph   map[string]*CallInfo
	LoopContext map[ast.Node]*LoopInfo
	DataSizes   map[string]*DataSizeInfo
}

type CallInfo struct {
	Function  *ast.FuncDecl
	CallSites []ast.Node
	IsHotPath bool
	Frequency FrequencyEstimate
}

type LoopInfo struct {
	LoopNode     ast.Node
	BoundType    LoopBoundType
	EstimatedMax int
	IsInnerLoop  bool
	HasEarlyExit bool
}

type DataSizeInfo struct {
	EstimatedLen int
	Confidence   float64
	Source       string // "literal", "parameter", "unknown"
}

type LoopBoundType int

const (
	BoundUnknown  LoopBoundType = iota
	BoundConstant               // for i := 0; i < 10; i++
	BoundLinear                 // for i := 0; i < len(slice); i++
	BoundVariable               // for i := 0; i < someVar; i++
)

type FrequencyEstimate int

const (
	FrequencyUnknown  FrequencyEstimate = iota
	FrequencyRare                       // Error paths, initialization
	FrequencyModerate                   // Normal business logic
	FrequencyHigh                       // Hot paths, tight loops
)
