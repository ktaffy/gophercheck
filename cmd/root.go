package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gophercheck/internal/analyzer"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	formatFlag string
	watchFlag  bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gophercheck [files or directories]",
	Short: "A Go performance analyzer that detects optimization opportunities",
	Long: `gophercheck is a static analysis tool that scans Go code for common
performance issues and provides actionable optimization suggestions.

Examples:
  gophercheck .                    # Analyze current directory
  gophercheck main.go utils.go     # Analyze specific files
  gophercheck --format=json .      # Output results in JSON format`,
	Run: runAnalysis,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&formatFlag, "format", "f", "console", "Output format (console, json)")
	rootCmd.Flags().BoolVarP(&watchFlag, "watch", "w", false, "Watch mode for development")
}

func runAnalysis(cmd *cobra.Command, args []string) {
	// Default to current directory if no args provided
	if len(args) == 0 {
		args = []string{"."}
	}

	// Collect Go files to analyze
	var goFiles []string
	for _, arg := range args {
		files, err := collectGoFiles(arg)
		if err != nil {
			color.Red("Error collecting files from %s: %v\n", arg, err)
			continue
		}
		goFiles = append(goFiles, files...)
	}

	if len(goFiles) == 0 {
		color.Yellow("⚠️  No Go files found to analyze\n")
		return
	}

	// Create analyzer and run analysis
	analyzerEngine := analyzer.NewAnalyzer()
	reportGen := analyzer.NewReportGenerator(formatFlag)

	color.Cyan("🔍 Analyzing %d Go files...\n\n", len(goFiles))

	result, err := analyzerEngine.AnalyzeFiles(goFiles)
	if err != nil {
		color.Red("Analysis failed: %v\n", err)
		return
	}

	// Generate and display report
	report := reportGen.Generate(result)
	fmt.Print(report)
}

// collectGoFiles recursively finds all .go files in the given path
func collectGoFiles(path string) ([]string, error) {
	var goFiles []string

	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip vendor, .git, and other common directories
		if info.IsDir() {
			name := info.Name()
			if name == "vendor" || name == ".git" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}

		// Only include .go files, but exclude _test.go files for now
		if strings.HasSuffix(filePath, ".go") && !strings.HasSuffix(filePath, "_test.go") {
			goFiles = append(goFiles, filePath)
		}

		return nil
	})

	return goFiles, err
}
