package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/lukaszraczylo/git-velocity/internal/app"
	"github.com/lukaszraczylo/git-velocity/internal/server"
	"github.com/lukaszraczylo/git-velocity/pkg/version"
)

var (
	configPath string
	outputDir  string
	verbose    bool
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "git-velocity",
		Short: "Analyze Git repositories for developer velocity metrics",
		Long: `Git Velocity Analyser - Track developer activity,
generate beautiful dashboards, and gamify contributions.

This tool analyzes GitHub repositories to generate velocity metrics,
including commits, pull requests, code reviews, and more. It creates
static HTML dashboards with charts and gamification features.`,
	}

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c",
		"config.yaml", "Path to configuration file")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v",
		false, "Enable verbose output")

	// Add subcommands
	rootCmd.AddCommand(newAnalyzeCmd())
	rootCmd.AddCommand(newServeCmd())
	rootCmd.AddCommand(newVersionCmd())

	return rootCmd
}

func newAnalyzeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "Analyze repositories and generate dashboard",
		Long: `Analyze the configured repositories and generate a static HTML dashboard.

This command will:
1. Fetch data from the configured GitHub repositories
2. Calculate velocity metrics for each contributor
3. Generate scores and achievements
4. Create a static HTML site with charts and leaderboards`,
		RunE: runAnalyze,
	}

	cmd.Flags().StringVarP(&outputDir, "output", "o",
		"./dist", "Output directory for generated site")

	return cmd
}

func newServeCmd() *cobra.Command {
	var port string
	var dir string

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start local preview server",
		Long: `Start a local HTTP server to preview the generated dashboard.

This is useful for testing the generated site before deployment.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServe(dir, port)
		},
	}

	cmd.Flags().StringVarP(&dir, "directory", "d",
		"./dist", "Directory to serve")
	cmd.Flags().StringVarP(&port, "port", "p",
		"8080", "Port to listen on")

	return cmd
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("git-velocity %s\n", version.Version)
			fmt.Printf("commit: %s\n", version.Commit)
			fmt.Printf("built: %s\n", version.BuildDate)
		},
	}
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	// Create and run the application
	application, err := app.New(configPath, outputDir, verbose)
	if err != nil {
		return fmt.Errorf("failed to initialize application: %w", err)
	}

	return application.Run(cmd.Context())
}

func runServe(dir, port string) error {
	srv := server.New(dir, port)

	fmt.Printf("Starting preview server at http://localhost:%s\n", port)
	fmt.Printf("Serving directory: %s\n", dir)
	fmt.Println("Press Ctrl+C to stop")

	return srv.Start()
}
