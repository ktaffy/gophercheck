package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gophercheck/internal/analyzer"
	"gophercheck/internal/config"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	formatFlag         string
	watchFlag          bool
	configFlag         string
	generateConfigFlag bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gophercheck [files or directories]",
	Short: "A Go performance analyzer that detects optimization opportunities",
	Long: `gophercheck is a static analysis tool that scans Go code for common
performance issues and provides actionable optimization suggestions.

Examples:
  gophercheck .                            # Analyze current directory
  gophercheck main.go utils.go             # Analyze specific files
  gophercheck --format=json .              # Output results in JSON format
  gophercheck --config-.gophercheck.yml    # Use custom config
  gophercheck --generate-config            # Generate sample config file`,
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
	rootCmd.Flags().StringVarP(&configFlag, "config", "c", "", "Path to configuration file")
	rootCmd.Flags().BoolVar(&generateConfigFlag, "generate-config", false, "Generate sample configuration file")
}

func runAnalysis(cmd *cobra.Command, args []string) {

	if generateConfigFlag {
		generateConfig()
		return
	}

	cfg, err := config.LoadConfig(configFlag)
	if err != nil {
		color.Red("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	if formatFlag != "" {
		cfg.Output.Format = formatFlag
	}

	if len(args) == 0 {
		args = []string{"."}
	}

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

	analyzerEngine := analyzer.NewAnalyzerWithConfig(cfg)
	reportGen := analyzer.NewReportGeneratorWithConfig(cfg)

	if cfg.Output.Verbose {
		color.Cyan("🔍 Analyzing %d Go files with %d detectors...\n", len(goFiles), analyzerEngine.GetDetectorCount())
		if configFlag != "" {
			color.Cyan("📋 Using configuration: %s\n", configFlag)
		}
		color.Cyan("🎯 Enabled categories: %s\n\n", strings.Join(cfg.Analysis.EnabledCategories, ", "))
	} else {
		color.Cyan("🔍 Analyzing %d Go files...\n\n", len(goFiles))
	}

	result, err := analyzerEngine.AnalyzeFiles(goFiles)
	if err != nil {
		color.Red("Analysis failed: %v\n", err)
		return
	}

	report := reportGen.Generate(result)

	if cfg.Output.OutputFile != "" {
		if err := writeReportToFile(report, cfg.Output.OutputFile); err != nil {
			color.Red("Failed to write report to file: %v\n", err)
		} else {
			color.Green("📄 Report saved to: %s\n", cfg.Output.OutputFile)
		}
	} else {
		fmt.Print(report)
	}

	if !cfg.Output.Colors && result.PerformanceScore < cfg.Analysis.ScoreThresholds.Fair {
		os.Exit(1)
	}
}

func writeReportToFile(report, filePath string) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(filePath, []byte(report), 0644)
}

func generateConfig() {
	configPath := ".gophercheck.yml"
	if err := config.GenerateConfig(configPath); err != nil {
		color.Red("Failed to generate config file: %v\n", err)
		os.Exit(1)
	}
	color.Green("✅ Generated sample configuration file: %s\n", configPath)
	color.Cyan("📝 Edit this file to customize gophercheck behavior\n")
	color.Cyan("🚀 Run 'gophercheck --config=%s .' to use it\n", configPath)
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
