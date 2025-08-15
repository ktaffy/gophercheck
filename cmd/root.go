package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	formatFlag string
	watchFlag  bool
)

var rootCmd = &cobra.Command{
	Use:   "gophercheck [files or directories]",
	Short: "A Go performance analyzer that detects optimization opportunities",
	Long: `gophercheck is a static analysis tool that scans Go code for common performance issues and provides actionable optimization suggestions.

Examples:
	gophercheck .                     # Analyze current directory
	gophercheck main.go utils.go      # Analyze specific files
	gophercheck --format=json .       # Output results in JSON format`,
	Run: runAnalysis,
}

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
	if len(args) == 0 {
		args = []string{"."}
	}
	color.Cyan("🔍 GopherCheck - Go Performance Analyzer\n")
	color.White("═══════════════════════════════════════\n\n")

	var goFiles []string
	for _, arg := range args {
		files, err := collectGoFiles(arg)
		if err != nil {
			color.Red("Error colelcting files from %s: %v\n", arg, err)
			continue
		}
		goFiles = append(goFiles, files...)
	}
	if len(goFiles) == 0 {
		color.Yellow("⚠️  No Go files found to analyze\n")
		return
	}
	color.Green("📁 Found %d Go files to analyze\n", len(goFiles))

	// TODO: This is where we'll integrate the actual analysis
	// For now, just show that we found the files
	for _, file := range goFiles {
		color.White("   • %s\n", file)
	}

	color.Yellow("\n🚧 Analysis engine coming soon...\n")
	color.White("Next steps: Implement AST parsing and performance detection\n")
}

func collectGoFiles(path string) ([]string, error) {
	var goFiles []string
	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			name := info.Name()
			if name == "vendor" || name == ".git" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(filePath, ".go") && !strings.HasSuffix(filePath, "_test.go") {
			goFiles = append(goFiles, filePath)
		}
		return nil
	})
	return goFiles, err
}
