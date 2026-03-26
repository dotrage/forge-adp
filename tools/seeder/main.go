package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed templates/*
var templates embed.FS

type ProjectConfig struct {
	ProjectName    string
	CompanyID      string
	ProjectID      string
	JiraProjectKey string
	GitHubRepo     string
	SlackChannel   string
	TechStack      TechStack
	Agents         []string
}

type TechStack struct {
	Frontend       string
	Backend        string
	Database       string
	Infrastructure string
	CICD           string
}

func main() {
	var config ProjectConfig

	flag.StringVar(&config.ProjectName, "name", "", "Project name")
	flag.StringVar(&config.CompanyID, "company", "", "Company ID")
	flag.StringVar(&config.ProjectID, "project", "", "Project ID")
	flag.StringVar(&config.JiraProjectKey, "jira-key", "", "Jira project key")
	flag.StringVar(&config.GitHubRepo, "github-repo", "", "GitHub repository (org/repo)")
	flag.StringVar(&config.SlackChannel, "slack-channel", "", "Slack channel name")

	// Tech stack
	flag.StringVar(&config.TechStack.Frontend, "frontend", "Next.js 14, TypeScript", "Frontend stack")
	flag.StringVar(&config.TechStack.Backend, "backend", "Go 1.22", "Backend stack")
	flag.StringVar(&config.TechStack.Database, "database", "PostgreSQL 16", "Database")
	flag.StringVar(&config.TechStack.Infrastructure, "infra", "AWS, Terraform", "Infrastructure")
	flag.StringVar(&config.TechStack.CICD, "cicd", "GitHub Actions", "CI/CD")

	// Agents
	var agents string
	flag.StringVar(&agents, "agents", "pm,architect,backend-developer,frontend-developer,qa,secops", "Comma-separated agent roles")

	targetDir := flag.String("output", ".", "Output directory")

	flag.Parse()

	if config.ProjectName == "" {
		fmt.Println("Error: -name is required")
		os.Exit(1)
	}

	config.Agents = strings.Split(agents, ",")

	if err := seedProject(*targetDir, config); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Seeded Forge project in %s/.forge/\n", *targetDir)
}

func seedProject(targetDir string, config ProjectConfig) error {
	forgeDir := filepath.Join(targetDir, ".forge")
	if err := os.MkdirAll(forgeDir, 0755); err != nil {
		return fmt.Errorf("create .forge directory: %w", err)
	}

	return fs.WalkDir(templates, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}

		// Read template
		content, err := templates.ReadFile(path)
		if err != nil {
			return err
		}

		// Parse and execute template
		tmpl, err := template.New(d.Name()).Parse(string(content))
		if err != nil {
			return err
		}

		// Determine output filename (remove .tmpl extension if present)
		outName := strings.TrimPrefix(path, "templates/")
		outName = strings.TrimSuffix(outName, ".tmpl")
		outPath := filepath.Join(forgeDir, outName)

		// Create parent dirs if needed
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return err
		}

		// Create output file
		outFile, err := os.Create(outPath)
		if err != nil {
			return err
		}
		defer outFile.Close()

		return tmpl.Execute(outFile, config)
	})
}
