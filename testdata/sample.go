package testdata

import (
	"fmt"
	"math"
	"strings"
)

// BadNestedLoop demonstrates O(n²) complexity - should be detected
func BadNestedLoop(users []User, posts []Post) {
	for i := range users {
		for j := range posts {
			if posts[j].UserID == users[i].ID {
				fmt.Printf("User %s has post %s\n", users[i].Name, posts[j].Title)
			}
		}
	}
}

// BadStringConcat demonstrates inefficient string building - should be detected
func BadStringConcat(items []string) string {
	var result string
	for _, item := range items {
		result += item // This creates new strings each time
	}
	return result
}

// BadSliceSearch demonstrates O(n) search that could be O(1) - should be detected
func BadSliceSearch(slice []string, target string) bool {
	for _, item := range slice {
		if item == target {
			return true
		}
	}
	return false
}

// ComplexFunction has high cyclomatic complexity - should be detected
func ComplexFunction(x, y, z int) string {
	if x > 0 {
		if y > 0 {
			if z > 0 {
				if x > y {
					if y > z {
						if x > 10 {
							if y > 5 {
								return "case1"
							} else {
								return "case2"
							}
						} else {
							return "case3"
						}
					} else {
						return "case4"
					}
				} else {
					return "case5"
				}
			} else {
				return "case6"
			}
		} else {
			return "case7"
		}
	} else {
		return "case8"
	}
}

// BadMemoryAllocationInLoop demonstrates allocation inside loop - should be detected
func BadMemoryAllocationInLoop(data [][]int) [][]int {
	var results [][]int
	for i := 0; i < len(data); i++ {
		temp := make([]int, 10) // Allocates memory each iteration
		for j := 0; j < 10; j++ {
			temp[j] = data[i][j] * 2
		}
		results = append(results, temp)
	}
	return results
}

// BadSliceWithoutCapacity demonstrates slice creation without capacity - should be detected
func BadSliceWithoutCapacity(items []string) []string {
	filtered := make([]string, 0) // Should specify capacity
	for _, item := range items {
		if len(item) > 5 {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// BadMapWithoutSize demonstrates map creation without size hint - should be detected
func BadMapWithoutSize(users []User) map[int]User {
	userMap := make(map[int]User) // Should specify size
	for _, user := range users {
		userMap[user.ID] = user
	}
	return userMap
}

// BadAppendInLoop demonstrates append without preallocation - should be detected
func BadAppendInLoop(count int) []int {
	var numbers []int
	for i := 0; i < count; i++ {
		numbers = append(numbers, i*i) // Grows slice each time
	}
	return numbers
}

// GoodMemoryPattern shows the optimized version
func GoodMemoryPattern(data [][]int) [][]int {
	results := make([][]int, 0, len(data)) // Pre-allocate capacity
	temp := make([]int, 10)                // Reuse allocation

	for i := 0; i < len(data); i++ {
		temp = temp[:0] // Reset slice, keep capacity
		for j := 0; j < 10; j++ {
			temp = append(temp, data[i][j]*2)
		}
		// Copy to avoid sharing underlying array
		row := make([]int, len(temp))
		copy(row, temp)
		results = append(results, row)
	}
	return results
}

// BadSliceGrowthPattern demonstrates slice growth without capacity - NEW DETECTION
func BadSliceGrowthPattern(data []int) []int {
	// Issue 1: Slice created without capacity hint
	results := make([]int, 0) // Should specify capacity

	for _, item := range data {
		// Issue 2: Multiple appends without preallocation
		results = append(results, item*2)
	}

	// Issue 3: Slice created in loop without capacity
	for i := 0; i < 3; i++ {
		temp := make([]int, 0) // Created in loop, no capacity
		for j := 0; j < 10; j++ {
			temp = append(temp, j)
		}
		results = append(results, temp...)
	}

	return results
}

// BadDataStructureUsage demonstrates using wrong data structures - NEW DETECTION
func BadDataStructureUsage(users []User, targetID int) *User {
	// Issue: Using slice for frequent lookups instead of map
	for _, user := range users { // O(n) lookup
		if user.ID == targetID {
			return &user
		}
	}

	// Another issue: Multiple linear searches
	for i := 0; i < 10; i++ {
		for _, user := range users { // Repeated O(n) searches
			if user.ID == i {
				fmt.Printf("Found user: %s\n", user.Name)
			}
		}
	}

	return nil
}

// OptimizedDataStructureUsage shows the better approach
func OptimizedDataStructureUsage(users []User, targetID int) *User {
	// Create map for O(1) lookups
	userMap := make(map[int]User, len(users)) // Pre-sized map
	for _, user := range users {
		userMap[user.ID] = user
	}

	// O(1) lookup instead of O(n)
	if user, exists := userMap[targetID]; exists {
		return &user
	}

	return nil
}

// ExtremelyLongFunction demonstrates function that's too long - NEW DETECTION
func ExtremelyLongFunction(data []ComplexData, config Config) (*Result, error) {
	// This function is intentionally very long to trigger the detector
	// Line 1: Input validation
	if data == nil {
		return nil, fmt.Errorf("data cannot be nil")
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("data cannot be empty")
	}
	if config.MaxItems <= 0 {
		return nil, fmt.Errorf("config.MaxItems must be positive")
	}
	if config.Threshold < 0 {
		return nil, fmt.Errorf("config.Threshold cannot be negative")
	}

	// Line 10: Data preprocessing
	var preprocessed []ComplexData
	for _, item := range data {
		if item.IsValid() {
			item.Normalize()
			if item.Score > config.Threshold {
				preprocessed = append(preprocessed, item)
			}
		}
	}

	// Line 20: Complex business logic section 1
	results := make([]ProcessedItem, 0, len(preprocessed))
	errorCount := 0
	warningCount := 0

	for i, item := range preprocessed {
		if i >= config.MaxItems {
			break
		}

		// Complex processing with multiple nested conditions
		if item.Category == "A" {
			if item.Priority == 1 {
				if item.Score > 90 {
					result := ProcessedItem{
						ID:       item.ID,
						Score:    item.Score * 1.2,
						Category: "HIGH_PRIORITY_A",
						Status:   "PROCESSED",
					}
					results = append(results, result)
				} else if item.Score > 70 {
					result := ProcessedItem{
						ID:       item.ID,
						Score:    item.Score * 1.1,
						Category: "MEDIUM_PRIORITY_A",
						Status:   "PROCESSED",
					}
					results = append(results, result)
				} else {
					warningCount++
				}
			} else if item.Priority == 2 {
				if item.Score > 80 {
					result := ProcessedItem{
						ID:       item.ID,
						Score:    item.Score * 1.05,
						Category: "LOW_PRIORITY_A",
						Status:   "PROCESSED",
					}
					results = append(results, result)
				} else {
					warningCount++
				}
			} else {
				errorCount++
			}
		} else if item.Category == "B" {
			if item.Priority == 1 {
				if item.Score > 85 {
					result := ProcessedItem{
						ID:       item.ID,
						Score:    item.Score * 1.15,
						Category: "HIGH_PRIORITY_B",
						Status:   "PROCESSED",
					}
					results = append(results, result)
				} else if item.Score > 60 {
					result := ProcessedItem{
						ID:       item.ID,
						Score:    item.Score,
						Category: "MEDIUM_PRIORITY_B",
						Status:   "PROCESSED",
					}
					results = append(results, result)
				} else {
					warningCount++
				}
			} else {
				errorCount++
			}
		} else if item.Category == "C" {
			if item.Score > 95 {
				result := ProcessedItem{
					ID:       item.ID,
					Score:    item.Score * 1.3,
					Category: "EXCEPTIONAL_C",
					Status:   "PROCESSED",
				}
				results = append(results, result)
			} else if item.Score > 75 {
				result := ProcessedItem{
					ID:       item.ID,
					Score:    item.Score * 1.1,
					Category: "GOOD_C",
					Status:   "PROCESSED",
				}
				results = append(results, result)
			} else if item.Score > 50 {
				result := ProcessedItem{
					ID:       item.ID,
					Score:    item.Score,
					Category: "AVERAGE_C",
					Status:   "PROCESSED",
				}
				results = append(results, result)
			} else {
				warningCount++
			}
		} else {
			errorCount++
		}
	}

	// Line 70: Post-processing and validation
	finalResults := make([]ProcessedItem, 0, len(results))
	for _, result := range results {
		if result.Score > 0 {
			// Additional validation logic
			if result.Category != "" {
				if len(result.Category) > 3 {
					if strings.Contains(result.Category, "HIGH") {
						result.Status = "VERIFIED_HIGH"
					} else if strings.Contains(result.Category, "MEDIUM") {
						result.Status = "VERIFIED_MEDIUM"
					} else if strings.Contains(result.Category, "LOW") {
						result.Status = "VERIFIED_LOW"
					} else {
						result.Status = "VERIFIED_OTHER"
					}
					finalResults = append(finalResults, result)
				}
			}
		}
	}

	// Line 90: Statistics calculation
	totalScore := 0.0
	maxScore := 0.0
	minScore := math.Inf(1)
	categoryStats := make(map[string]int)

	for _, result := range finalResults {
		totalScore += result.Score
		if result.Score > maxScore {
			maxScore = result.Score
		}
		if result.Score < minScore {
			minScore = result.Score
		}
		categoryStats[result.Category]++
	}

	avgScore := totalScore / float64(len(finalResults))

	// Line 110: Report generation
	report := &Result{
		TotalItems:    len(finalResults),
		AverageScore:  avgScore,
		MaxScore:      maxScore,
		MinScore:      minScore,
		ErrorCount:    errorCount,
		WarningCount:  warningCount,
		CategoryStats: categoryStats,
		Items:         finalResults,
	}

	// Line 120: Final validation and cleanup
	if report.TotalItems == 0 {
		return report, fmt.Errorf("no items were successfully processed")
	}

	if report.ErrorCount > len(data)/2 {
		return report, fmt.Errorf("too many errors: %d out of %d items failed", report.ErrorCount, len(data))
	}

	if report.AverageScore < config.MinAverage {
		return report, fmt.Errorf("average score %f is below minimum %f", report.AverageScore, config.MinAverage)
	}

	// This function is now over 130 lines and should trigger the detector!
	return report, nil
}

// VeryLongFunctionThatShouldTriggerCritical demonstrates critical length issue - NEW DETECTION
func VeryLongFunctionThatShouldTriggerCritical(input *MegaComplexInput) (*MegaComplexOutput, error) {
	// This function will be extremely long to trigger CRITICAL severity
	// [Imagine 200+ lines of complex business logic here]

	// Input validation (20 lines)
	if input == nil {
		return nil, fmt.Errorf("input nil")
	}
	if input.Data == nil {
		return nil, fmt.Errorf("data nil")
	}
	if len(input.Data) == 0 {
		return nil, fmt.Errorf("data empty")
	}
	if input.Config == nil {
		return nil, fmt.Errorf("config nil")
	}
	if input.Config.Rules == nil {
		return nil, fmt.Errorf("rules nil")
	}
	if len(input.Config.Rules) == 0 {
		return nil, fmt.Errorf("rules empty")
	}
	if input.Config.Thresholds == nil {
		return nil, fmt.Errorf("thresholds nil")
	}
	if input.Config.Validators == nil {
		return nil, fmt.Errorf("validators nil")
	}
	if input.Config.Processors == nil {
		return nil, fmt.Errorf("processors nil")
	}
	if input.Config.Transformers == nil {
		return nil, fmt.Errorf("transformers nil")
	}

	// Data preprocessing (30 lines)
	preprocessed := make([]PreprocessedData, 0, len(input.Data))
	for i, item := range input.Data {
		if item.IsValid() {
			validated := item.Validate(input.Config.Validators)
			if validated.Success {
				normalized := validated.Data.Normalize()
				if normalized.Score > input.Config.Thresholds.MinScore {
					preprocessed = append(preprocessed, PreprocessedData{
						ID:    normalized.ID,
						Score: normalized.Score,
					})
				}
			}
		}
		if i%1000 == 0 { /* progress logging */
		}
	}

	// Rule processing (40 lines)
	ruleResults := make(map[string]RuleResult)
	for _, rule := range input.Config.Rules {
		result := RuleResult{RuleID: rule.ID, Matches: 0}
		for _, item := range preprocessed {
			if rule.Evaluate(item) {
				result.Matches++
				result.Items = append(result.Items, item.ID)
			}
		}
		ruleResults[rule.ID] = result
	}

	// Transformation phase (50 lines)
	transformed := make([]TransformedData, 0, len(preprocessed))
	for _, item := range preprocessed {
		for _, transformer := range input.Config.Transformers {
			if transformer.AppliesTo(item) {
				result := transformer.Transform(item)
				if result.Success {
					transformed = append(transformed, result.Data)
				}
			}
		}
	}

	// Analysis phase (60 lines)
	analysis := &Analysis{}
	for _, item := range transformed {
		analysis.ProcessItem(item)
	}

	// Output generation (30 lines)
	output := &MegaComplexOutput{
		ProcessedCount: len(transformed),
		Analysis:       analysis,
		Rules:          ruleResults,
	}

	// This function should now be well over 200 lines and trigger CRITICAL
	return output, nil
}

// =============================================================================
// SUPPORTING TYPES AND STRUCTURES
// =============================================================================

// Main types used throughout the examples
type User struct {
	ID   int
	Name string
}

type Post struct {
	ID     int
	UserID int
	Title  string
}

type ComplexData struct {
	ID       int
	Score    float64
	Category string
	Priority int
}

func (c *ComplexData) IsValid() bool { return c.Score >= 0 }
func (c *ComplexData) Normalize() ComplexData {
	return ComplexData{
		ID:       c.ID,
		Score:    c.Score,
		Category: c.Category,
		Priority: c.Priority,
	}
}

type ProcessedItem struct {
	ID       int
	Score    float64
	Category string
	Status   string
}

type Config struct {
	MaxItems   int
	Threshold  float64
	MinAverage float64
}

type Result struct {
	TotalItems    int
	AverageScore  float64
	MaxScore      float64
	MinScore      float64
	ErrorCount    int
	WarningCount  int
	CategoryStats map[string]int
	Items         []ProcessedItem
}

// Types for the very long function examples
type MegaComplexInput struct {
	Data   []ComplexData
	Config *MegaConfig
}

type MegaConfig struct {
	Rules        []Rule
	Thresholds   *Thresholds
	Validators   []Validator
	Processors   []Processor
	Transformers []Transformer
}

type Rule struct{ ID string }

func (r Rule) Evaluate(data PreprocessedData) bool { return true }

type Thresholds struct{ MinScore float64 }
type Validator struct{}
type Processor struct{}
type Transformer struct{}

func (t Transformer) AppliesTo(data PreprocessedData) bool { return true }
func (t Transformer) Transform(data PreprocessedData) TransformResult {
	return TransformResult{Success: true, Data: TransformedData{ID: data.ID}}
}

type PreprocessedData struct {
	ID    int
	Score float64
}

type ValidatedData struct {
	Success bool
	Data    ComplexData
}

func (c ComplexData) Validate(validators []Validator) ValidatedData {
	return ValidatedData{Success: true, Data: c}
}

type TransformResult struct {
	Success bool
	Data    TransformedData
}

type TransformedData struct{ ID int }
type Analysis struct{}

func (a *Analysis) ProcessItem(item TransformedData) {}

type RuleResult struct {
	RuleID  string
	Matches int
	Items   []int
}

type MegaComplexOutput struct {
	ProcessedCount int
	Analysis       *Analysis
	Rules          map[string]RuleResult
}
