package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"license-audit/internal/config"
	"license-audit/internal/scanner"
	"license-audit/internal/audit"
	"license-audit/internal/output"
)

var (
	configFile   string
	outputFormat string
	outputFile   string
	scanPath     string
	enableAudit  bool
)

var rootCmd = &cobra.Command{
	Use:   "license-audit",
	Short: "A comprehensive license auditing tool for various package managers",
	Long: `license-audit scans your project dependencies and generates detailed 
license reports. It supports Node.js, Go, Docker, Python, Ruby, and Java projects.`,
	Run: run,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default searches for .license-audit.toml in current dir and home)")
	rootCmd.PersistentFlags().StringVar(&outputFormat, "format", "json", "output format (json or markdown)")
	rootCmd.PersistentFlags().StringVar(&outputFile, "output", "", "output file (default: stdout)")
	rootCmd.PersistentFlags().StringVar(&scanPath, "path", ".", "path to scan")
	rootCmd.PersistentFlags().BoolVar(&enableAudit, "audit", true, "enable license auditing")
}

func run(cmd *cobra.Command, args []string) {
	cfg, err := config.Load(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Override config with command line flags
	if outputFormat != "" {
		cfg.OutputFormat = outputFormat
	}
	if outputFile != "" {
		cfg.OutputFile = outputFile
	}
	if scanPath != "" {
		cfg.ScanPaths = []string{scanPath}
	}
	if cmd.Flags().Changed("audit") {
		cfg.EnableAudit = enableAudit
	}

	// Initialize scanner
	s := scanner.New(cfg)
	
	// Scan for dependencies
	result, err := s.Scan()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning: %v\n", err)
		os.Exit(1)
	}

	// Run audit if enabled
	if cfg.EnableAudit {
		auditor := audit.New(cfg)
		result.Issues = auditor.Audit(result.Dependencies)
		result.Summary.IssueBreakdown = auditor.GetIssueBreakdown(result.Issues)
	}

	// Generate output
	outputter, err := output.New(cfg.OutputFormat)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output formatter: %v\n", err)
		os.Exit(1)
	}

	var outputPath string
	if cfg.OutputFile != "" {
		outputPath = cfg.OutputFile
	} else {
		if cfg.OutputFormat == "markdown" {
			outputPath = "license-report.md"
		} else {
			outputPath = "license-report.json"
		}
	}

	if err := outputter.Write(result, outputPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("License audit completed. Report written to %s\n", outputPath)
	
	if len(result.Issues) > 0 {
		fmt.Printf("Found %d license issues. Review the audit report for details.\n", len(result.Issues))
		// Exit with code 1 if there are issues to support CI/CD pipelines
		os.Exit(1)
	}
}