// internal/config/config.go
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the configuration for gophercheck
type Config struct {
	// General settings
	Version     string `yaml:"version" json:"version"`
	ProjectName string `yaml:"project_name,omitempty" json:"project_name,omitempty"`

	// Analysis settings
	Analysis AnalysisConfig `yaml:"analysis" json:"analysis"`

	// Output settings
	Output OutputConfig `yaml:"output" json:"output"`

	// Rule-specific configurations
	Rules RulesConfig `yaml:"rules" json:"rules"`

	// File patterns
	Files FilesConfig `yaml:"files" json:"files"`
}

type AnalysisConfig struct {
	// Performance score thresholds
	ScoreThresholds ScoreThresholds `yaml:"score_thresholds" json:"score_thresholds"`

	// Enable/disable entire categories
	EnabledCategories []string `yaml:"enabled_categories" json:"enabled_categories"`

	// Parallel analysis
	MaxWorkers int `yaml:"max_workers" json:"max_workers"`
}

type ScoreThresholds struct {
	Excellent int `yaml:"excellent" json:"excellent"` // >= 90
	Good      int `yaml:"good" json:"good"`           // >= 75
	Fair      int `yaml:"fair" json:"fair"`           // >= 50
	Poor      int `yaml:"poor" json:"poor"`           // < 50
}

type OutputConfig struct {
	// Default output format
	Format string `yaml:"format" json:"format"`

	// Colorized output
	Colors bool `yaml:"colors" json:"colors"`

	// Verbosity level
	Verbose bool `yaml:"verbose" json:"verbose"`

	// Show suggestions
	ShowSuggestions bool `yaml:"show_suggestions" json:"show_suggestions"`

	// Output file path (optional)
	OutputFile string `yaml:"output_file,omitempty" json:"output_file,omitempty"`
}

type RulesConfig struct {
	// Complexity rules
	Complexity ComplexityRules `yaml:"complexity" json:"complexity"`

	// Performance rules
	Performance PerformanceRules `yaml:"performance" json:"performance"`

	// Code quality rules
	Quality QualityRules `yaml:"quality" json:"quality"`

	// Memory rules
	Memory MemoryRules `yaml:"memory" json:"memory"`
}

type ComplexityRules struct {
	Enabled bool `yaml:"enabled" json:"enabled"`

	// Cyclomatic complexity thresholds
	CyclomaticComplexity ThresholdConfig `yaml:"cyclomatic_complexity" json:"cyclomatic_complexity"`

	// Function length thresholds
	FunctionLength FunctionLengthConfig `yaml:"function_length" json:"function_length"`
}

type PerformanceRules struct {
	Enabled bool `yaml:"enabled" json:"enabled"`

	// Nested loop detection
	NestedLoops NestedLoopConfig `yaml:"nested_loops" json:"nested_loops"`

	// String concatenation
	StringConcat StringConcatConfig `yaml:"string_concat" json:"string_concat"`

	// Data structure usage
	DataStructure DataStructureConfig `yaml:"data_structure" json:"data_structure"`
}

type QualityRules struct {
	Enabled bool `yaml:"enabled" json:"enabled"`

	// Import cycle detection
	ImportCycles ImportCycleConfig `yaml:"import_cycles" json:"import_cycles"`
}

type MemoryRules struct {
	Enabled bool `yaml:"enabled" json:"enabled"`

	// Memory allocation patterns
	Allocation AllocationConfig `yaml:"allocation" json:"allocation"`

	// Slice growth patterns
	SliceGrowth SliceGrowthConfig `yaml:"slice_growth" json:"slice_growth"`
}

// Individual rule configurations
type ThresholdConfig struct {
	Enabled           bool `yaml:"enabled" json:"enabled"`
	MediumThreshold   int  `yaml:"medium_threshold" json:"medium_threshold"`
	HighThreshold     int  `yaml:"high_threshold" json:"high_threshold"`
	CriticalThreshold int  `yaml:"critical_threshold" json:"critical_threshold"`
}

type FunctionLengthConfig struct {
	Enabled           bool `yaml:"enabled" json:"enabled"`
	MediumThreshold   int  `yaml:"medium_threshold" json:"medium_threshold"`     // lines
	HighThreshold     int  `yaml:"high_threshold" json:"high_threshold"`         // lines
	CriticalThreshold int  `yaml:"critical_threshold" json:"critical_threshold"` // lines
	CountComments     bool `yaml:"count_comments" json:"count_comments"`
	CountEmptyLines   bool `yaml:"count_empty_lines" json:"count_empty_lines"`
}

type NestedLoopConfig struct {
	Enabled    bool `yaml:"enabled" json:"enabled"`
	MaxDepth   int  `yaml:"max_depth" json:"max_depth"`
	IgnoreTest bool `yaml:"ignore_test" json:"ignore_test"`
}

type StringConcatConfig struct {
	Enabled              bool     `yaml:"enabled" json:"enabled"`
	DetectInLoops        bool     `yaml:"detect_in_loops" json:"detect_in_loops"`
	IgnoreShortStrings   bool     `yaml:"ignore_short_strings" json:"ignore_short_strings"`
	ShortStringThreshold int      `yaml:"short_string_threshold" json:"short_string_threshold"`
	StringVarNames       []string `yaml:"string_var_names" json:"string_var_names"`
}

type DataStructureConfig struct {
	Enabled             bool `yaml:"enabled" json:"enabled"`
	DetectLinearSearch  bool `yaml:"detect_linear_search" json:"detect_linear_search"`
	MinSearchComplexity int  `yaml:"min_search_complexity" json:"min_search_complexity"`
	SuggestMaps         bool `yaml:"suggest_maps" json:"suggest_maps"`
}

type ImportCycleConfig struct {
	Enabled            bool     `yaml:"enabled" json:"enabled"`
	MaxCycleLength     int      `yaml:"max_cycle_length" json:"max_cycle_length"`
	IgnoreTestPackages bool     `yaml:"ignore_test_packages" json:"ignore_test_packages"`
	IgnoreVendor       bool     `yaml:"ignore_vendor" json:"ignore_vendor"`
	ExcludePackages    []string `yaml:"exclude_packages" json:"exclude_packages"`
}

type AllocationConfig struct {
	Enabled              bool `yaml:"enabled" json:"enabled"`
	DetectInLoops        bool `yaml:"detect_in_loops" json:"detect_in_loops"`
	RequireCapacityHints bool `yaml:"require_capacity_hints" json:"require_capacity_hints"`
	MinLoopIterations    int  `yaml:"min_loop_iterations" json:"min_loop_iterations"`
}

type SliceGrowthConfig struct {
	Enabled             bool `yaml:"enabled" json:"enabled"`
	RequireCapacity     bool `yaml:"require_capacity" json:"require_capacity"`
	DetectAppendInLoops bool `yaml:"detect_append_in_loops" json:"detect_append_in_loops"`
	MinAppendCount      int  `yaml:"min_append_count" json:"min_append_count"`
}

type FilesConfig struct {
	// Include patterns
	Include []string `yaml:"include" json:"include"`

	// Exclude patterns
	Exclude []string `yaml:"exclude" json:"exclude"`

	// Whether to analyze test files
	IncludeTests bool `yaml:"include_tests" json:"include_tests"`

	// Whether to follow symlinks
	FollowSymlinks bool `yaml:"follow_symlinks" json:"follow_symlinks"`

	// Max file size (in KB)
	MaxFileSize int `yaml:"max_file_size" json:"max_file_size"`
}

func DefaultConfig() *Config {
	return &Config{
		Version: "1.0",
		Analysis: AnalysisConfig{
			ScoreThresholds: ScoreThresholds{
				Excellent: 90,
				Good:      75,
				Fair:      50,
				Poor:      0,
			},
			EnabledCategories: []string{"performance", "complexity", "memory", "quality"},
			MaxWorkers:        4,
		},
		Output: OutputConfig{
			Format:          "console",
			Colors:          true,
			Verbose:         false,
			ShowSuggestions: false,
		},
		Rules: RulesConfig{
			Complexity: ComplexityRules{
				Enabled: true,
				CyclomaticComplexity: ThresholdConfig{
					Enabled:           true,
					MediumThreshold:   10,
					HighThreshold:     15,
					CriticalThreshold: 25,
				},
				FunctionLength: FunctionLengthConfig{
					Enabled:           true,
					MediumThreshold:   50,
					HighThreshold:     100,
					CriticalThreshold: 200,
					CountComments:     false,
					CountEmptyLines:   false,
				},
			},
			Performance: PerformanceRules{
				Enabled: true,
				NestedLoops: NestedLoopConfig{
					Enabled:    true,
					MaxDepth:   2,
					IgnoreTest: true,
				},
				StringConcat: StringConcatConfig{
					Enabled:              true,
					DetectInLoops:        true,
					IgnoreShortStrings:   true,
					ShortStringThreshold: 10,
					StringVarNames:       []string{"str", "result", "output", "text", "content", "message", "data"},
				},
				DataStructure: DataStructureConfig{
					Enabled:             true,
					DetectLinearSearch:  true,
					MinSearchComplexity: 2,
					SuggestMaps:         true,
				},
			},
			Quality: QualityRules{
				Enabled: true,
				ImportCycles: ImportCycleConfig{
					Enabled:            true,
					MaxCycleLength:     5,
					IgnoreTestPackages: true,
					IgnoreVendor:       true,
					ExcludePackages:    []string{},
				},
			},
			Memory: MemoryRules{
				Enabled: true,
				Allocation: AllocationConfig{
					Enabled:              true,
					DetectInLoops:        true,
					RequireCapacityHints: true,
					MinLoopIterations:    5,
				},
				SliceGrowth: SliceGrowthConfig{
					Enabled:             true,
					RequireCapacity:     true,
					DetectAppendInLoops: true,
					MinAppendCount:      3,
				},
			},
		},
		Files: FilesConfig{
			Include:        []string{"**/*.go"},
			Exclude:        []string{"vendor/**", ".git/**", "node_modules/**"},
			IncludeTests:   false,
			FollowSymlinks: false,
			MaxFileSize:    1024, // 1MB
		},
	}
}

// LoadConfig loads configuration from file or returns default
func LoadConfig(configPath string) (*Config, error) {
	// If no config path provided, look for default config files
	if configPath == "" {
		configPath = findConfigFile()
	}

	// If still no config found, return default
	if configPath == "" {
		return DefaultConfig(), nil
	}

	// Load from file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	config := DefaultConfig() // Start with defaults

	// Parse YAML
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", configPath, err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// findConfigFile looks for config files in common locations
func findConfigFile() string {
	possiblePaths := []string{
		".gophercheck.yml",
		".gophercheck.yaml",
		"gophercheck.yml",
		"gophercheck.yaml",
		".config/gophercheck.yml",
		".config/gophercheck.yaml",
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate score thresholds
	st := c.Analysis.ScoreThresholds
	if st.Excellent < st.Good || st.Good < st.Fair || st.Fair < st.Poor {
		return fmt.Errorf("score thresholds must be in descending order")
	}

	// Validate output format
	validFormats := []string{"console", "json", "html"}
	formatValid := false
	for _, format := range validFormats {
		if c.Output.Format == format {
			formatValid = true
			break
		}
	}
	if !formatValid {
		return fmt.Errorf("invalid output format: %s (valid: %v)", c.Output.Format, validFormats)
	}

	// Validate worker count
	if c.Analysis.MaxWorkers < 1 {
		return fmt.Errorf("max_workers must be at least 1")
	}

	// Validate complexity thresholds
	cc := c.Rules.Complexity.CyclomaticComplexity
	if cc.Enabled && (cc.MediumThreshold >= cc.HighThreshold || cc.HighThreshold >= cc.CriticalThreshold) {
		return fmt.Errorf("cyclomatic complexity thresholds must be in ascending order")
	}

	// Validate function length thresholds
	fl := c.Rules.Complexity.FunctionLength
	if fl.Enabled && (fl.MediumThreshold >= fl.HighThreshold || fl.HighThreshold >= fl.CriticalThreshold) {
		return fmt.Errorf("function length thresholds must be in ascending order")
	}

	return nil
}

// SaveConfig saves configuration to file
func (c *Config) SaveConfig(configPath string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GenerateConfig creates a sample configuration file
func GenerateConfig(configPath string) error {
	config := DefaultConfig()
	return config.SaveConfig(configPath)
}

// IsRuleEnabled checks if a specific rule is enabled
func (c *Config) IsRuleEnabled(ruleType string) bool {
	switch ruleType {
	case "cyclomatic_complexity":
		return c.Rules.Complexity.Enabled && c.Rules.Complexity.CyclomaticComplexity.Enabled
	case "function_length":
		return c.Rules.Complexity.Enabled && c.Rules.Complexity.FunctionLength.Enabled
	case "nested_loops":
		return c.Rules.Performance.Enabled && c.Rules.Performance.NestedLoops.Enabled
	case "string_concat":
		return c.Rules.Performance.Enabled && c.Rules.Performance.StringConcat.Enabled
	case "data_structure":
		return c.Rules.Performance.Enabled && c.Rules.Performance.DataStructure.Enabled
	case "import_cycles":
		return c.Rules.Quality.Enabled && c.Rules.Quality.ImportCycles.Enabled
	case "memory_allocation":
		return c.Rules.Memory.Enabled && c.Rules.Memory.Allocation.Enabled
	case "slice_growth":
		return c.Rules.Memory.Enabled && c.Rules.Memory.SliceGrowth.Enabled
	default:
		return false
	}
}

// GetThreshold returns the threshold for a given rule and severity
func (c *Config) GetThreshold(ruleType, severity string) int {
	switch ruleType {
	case "cyclomatic_complexity":
		switch severity {
		case "medium":
			return c.Rules.Complexity.CyclomaticComplexity.MediumThreshold
		case "high":
			return c.Rules.Complexity.CyclomaticComplexity.HighThreshold
		case "critical":
			return c.Rules.Complexity.CyclomaticComplexity.CriticalThreshold
		}
	case "function_length":
		switch severity {
		case "medium":
			return c.Rules.Complexity.FunctionLength.MediumThreshold
		case "high":
			return c.Rules.Complexity.FunctionLength.HighThreshold
		case "critical":
			return c.Rules.Complexity.FunctionLength.CriticalThreshold
		}
	}
	return 0
}
