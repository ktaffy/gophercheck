package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"gophercheck/internal/analyzer"
	"gophercheck/internal/config"
	"gophercheck/internal/watcher"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	formatFlag         string
	watchFlag          bool
	configFlag         string
	generateConfigFlag bool
	verboseFlag        bool
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
	gophercheck --config .gophercheck.yml .  # Use custom config
	gophercheck --watch .                    # Watch mode - analyze on file changes
	gophercheck --watch --verbose .          # Watch mode with detailed output
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
	rootCmd.Flags().BoolVarP(&verboseFlag, "verbose", "v", false, "Show detailed output with suggestions")
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

	verboseFlag, _ := cmd.Flags().GetBool("verbose")
	if verboseFlag {
		cfg.Output.Verbose = true
		cfg.Output.ShowSuggestions = true
	}

	if len(args) == 0 {
		args = []string{"."}
	}

	// Check if watch mode is enabled
	if watchFlag {
		runWatchMode(cfg, args)
		return
	}

	// Run normal analysis
	runSingleAnalysis(cfg, args)
}

func runWatchMode(cfg *config.Config, paths []string) {
	validPaths := make([]string, 0, len(paths))
	for _, path := range paths {
		if _, err := os.Stat(path); err != nil {
			color.Yellow("⚠️  Skipping invalid path: %s (%v)\n", path, err)
			continue
		}
		validPaths = append(validPaths, path)
	}

	if len(validPaths) == 0 {
		color.Red("❌ No valid paths to watch\n")
		os.Exit(1)
	}

	color.Cyan("🔄 Starting GopherCheck in watch mode...\n")
	color.White("Press Ctrl+C to stop watching\n\n")

	fileWatcher, err := watcher.NewFileWatcher(cfg)
	if err != nil {
		color.Red("Failed to create file watcher: %v\n", err)
		os.Exit(1)
	}
	defer fileWatcher.Close()

	analyzerEngine := analyzer.NewAnalyzerWithConfig(cfg)
	reportGen := analyzer.NewReportGeneratorWithConfig(cfg)

	color.Cyan("🔍 Running initial analysis...\n")
	runInitialAnalysis(cfg, validPaths, analyzerEngine, reportGen)

	changeHandler := func(changedFiles []string) error {
		return handleFileChanges(changedFiles, cfg, analyzerEngine, reportGen)
	}

	if err := fileWatcher.Watch(validPaths, changeHandler); err != nil {
		color.Red("Failed to start file watcher: %v\n", err)
		os.Exit(1)
	}

	if cfg.Output.Verbose {
		watchedPaths := fileWatcher.GetWatchedPaths()
		color.Cyan("👀 Watching %d directories for changes...\n", len(watchedPaths))
		for _, path := range watchedPaths {
			color.White("   - %s\n", path)
		}
	} else {
		color.Cyan("👀 Watching for Go file changes...\n")
	}

	color.White("Ready! Make changes to your Go files...\n\n")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	color.Yellow("\n🛑 Stopping watch mode...\n")
}

func runSingleAnalysis(cfg *config.Config, args []string) {
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

func runInitialAnalysis(cfg *config.Config, paths []string, analyzerEngine *analyzer.Analyzer, reportGen *analyzer.ReportGenerator) {
	var goFiles []string
	for _, path := range paths {
		files, err := collectGoFiles(path)
		if err != nil {
			color.Red("Error collecting files from %s: %v\n", path, err)
			continue
		}
		goFiles = append(goFiles, files...)
	}

	if len(goFiles) == 0 {
		color.Yellow("⚠️  No Go files found to analyze\n")
		return
	}

	if cfg.Output.Verbose {
		color.White("📋 Found %d Go files\n", len(goFiles))
	}

	result, err := analyzerEngine.AnalyzeFiles(goFiles)
	if err != nil {
		color.Red("Initial analysis failed: %v\n", err)
		return
	}

	report := reportGen.Generate(result)
	fmt.Print(report)

	color.White("═══════════════════════════════════════\n\n")
}

func handleFileChanges(changedFiles []string, cfg *config.Config, analyzerEngine *analyzer.Analyzer, reportGen *analyzer.ReportGenerator) error {
	if len(changedFiles) == 0 {
		return nil
	}

	timestamp := time.Now().Format("15:04:05")

	if len(changedFiles) == 1 {
		color.Cyan("🔄 [%s] File changed: %s\n", timestamp, filepath.Base(changedFiles[0]))
	} else {
		color.Cyan("🔄 [%s] %d files changed\n", timestamp, len(changedFiles))
		if cfg.Output.Verbose {
			for _, file := range changedFiles {
				color.White("   - %s\n", filepath.Base(file))
			}
		}
	}

	existingFiles := make([]string, 0, len(changedFiles))
	for _, file := range changedFiles {
		if stat, err := os.Stat(file); err == nil {
			if !stat.IsDir() && strings.HasSuffix(file, ".go") {
				if strings.HasSuffix(file, "_test.go") {
					if cfg.Files.IncludeTests {
						existingFiles = append(existingFiles, file)
					}
				} else {
					existingFiles = append(existingFiles, file)
				}
			}
		}
	}

	if len(existingFiles) == 0 {
		color.Yellow("⚠️  No valid Go files to analyze\n\n")
		return nil
	}

	if cfg.Output.Verbose && len(existingFiles) < len(changedFiles) {
		color.White("   → Analyzing %d Go files\n", len(existingFiles))
	}

	result, err := analyzerEngine.AnalyzeFiles(existingFiles)
	if err != nil {
		color.Red("Analysis failed: %v\n", err)
		color.Yellow("Continuing to watch for changes...\n\n")
		return nil
	}

	if result.TotalIssues > 0 {
		report := reportGen.Generate(result)
		fmt.Print(report)
	} else {
		color.Green("✅ No issues found in changed files (Score: %d/100)\n", result.PerformanceScore)
	}

	color.White("─────────────────────────────────────────\n\n")
	return nil
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
