package site

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	json "github.com/goccy/go-json"
	"github.com/lukaszraczylo/git-velocity/internal/config"
	"github.com/lukaszraczylo/git-velocity/internal/domain/models"
)

//go:embed dist/*
var spaFS embed.FS

// Generator handles static site generation
type Generator struct {
	outputDir string
	config    *config.Config
}

// NewGenerator creates a new site generator
func NewGenerator(outputDir string, cfg *config.Config) (*Generator, error) {
	return &Generator{
		outputDir: outputDir,
		config:    cfg,
	}, nil
}

// Generate creates the static site from metrics
func (g *Generator) Generate(metrics *models.GlobalMetrics) error {
	// Create output directory
	if err := os.MkdirAll(g.outputDir, 0750); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate data files
	if err := g.generateDataFiles(metrics); err != nil {
		return fmt.Errorf("failed to generate data files: %w", err)
	}

	// Copy Vue SPA files
	if err := g.copySPAFiles(); err != nil {
		return fmt.Errorf("failed to copy SPA files: %w", err)
	}

	return nil
}

func (g *Generator) generateDataFiles(metrics *models.GlobalMetrics) error {
	dataDir := filepath.Join(g.outputDir, "data")

	// Clean old data directory to ensure fresh state
	if err := os.RemoveAll(dataDir); err != nil {
		return fmt.Errorf("failed to clean data directory: %w", err)
	}

	if err := os.MkdirAll(dataDir, 0750); err != nil {
		return err
	}

	// Prepare global data with timestamp
	globalData := struct {
		*models.GlobalMetrics
		GeneratedAt time.Time `json:"generated_at"`
	}{
		GlobalMetrics: metrics,
		GeneratedAt:   time.Now(),
	}

	// Global metrics
	if err := writeJSON(filepath.Join(dataDir, "global.json"), globalData); err != nil {
		return err
	}

	// Leaderboard
	if err := writeJSON(filepath.Join(dataDir, "leaderboard.json"), metrics.Leaderboard); err != nil {
		return err
	}

	// Per-repository data
	for _, repo := range metrics.Repositories {
		repoDir := filepath.Join(dataDir, "repos", repo.Owner, repo.Name)
		if err := os.MkdirAll(repoDir, 0750); err != nil {
			return err
		}
		if err := writeJSON(filepath.Join(repoDir, "metrics.json"), repo); err != nil {
			return err
		}
	}

	// Per-team data
	if len(metrics.Teams) > 0 {
		teamDir := filepath.Join(dataDir, "teams")
		if err := os.MkdirAll(teamDir, 0750); err != nil {
			return err
		}
		for _, team := range metrics.Teams {
			if err := writeJSON(filepath.Join(teamDir, slugify(team.Name)+".json"), team); err != nil {
				return err
			}
		}
	}

	// Per-contributor data (use aggregated global contributors, not per-repo)
	contributorDir := filepath.Join(dataDir, "contributors")
	if err := os.MkdirAll(contributorDir, 0750); err != nil {
		return err
	}

	for _, contributor := range metrics.Contributors {
		if err := writeJSON(filepath.Join(contributorDir, contributor.Login+".json"), contributor); err != nil {
			return err
		}
	}

	return nil
}

func (g *Generator) copySPAFiles() error {
	return fs.WalkDir(spaFS, "dist", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the root dist directory itself
		if path == "dist" {
			return nil
		}

		// Calculate the relative path from "dist/"
		relPath := strings.TrimPrefix(path, "dist/")
		destPath := filepath.Join(g.outputDir, relPath)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0750)
		}

		// Read file from embedded FS
		content, err := spaFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read embedded file %s: %w", path, err)
		}

		// Write to destination
		return os.WriteFile(destPath, content, 0600)
	})
}

// Helper functions

func writeJSON(path string, data interface{}) error {
	cleanPath := filepath.Clean(path)
	file, err := os.OpenFile(cleanPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600) // #nosec G304 -- path is constructed internally
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")
	return s
}
