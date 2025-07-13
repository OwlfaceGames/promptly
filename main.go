package main

import (
	"bufio"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
)

//go:embed *.promptly
var promptFiles embed.FS

type Theme struct {
	Name        string
	Description string
	Content     string
	Preview     string
}

func main() {
	themes, err := loadThemes()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading themes: %v\n", err)
		os.Exit(1)
	}

	if len(themes) == 0 {
		fmt.Println("No .promptly themes found")
		os.Exit(1)
	}

	selectedTheme, err := selectTheme(themes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error selecting theme: %v\n", err)
		os.Exit(1)
	}

	if err := installTheme(selectedTheme); err != nil {
		fmt.Fprintf(os.Stderr, "Error installing theme: %v\n", err)
		os.Exit(1)
	}

	color.Green("✓ Theme '%s' installed successfully!", selectedTheme.Name)
	fmt.Println("Restart your terminal or run 'source ~/.zshrc' to apply the changes.")
}

func loadThemes() ([]Theme, error) {
	var themes []Theme

	err := fs.WalkDir(promptFiles, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".promptly") {
			return nil
		}

		content, err := promptFiles.ReadFile(path)
		if err != nil {
			return err
		}

		name := strings.TrimSuffix(filepath.Base(path), ".promptly")
		description := getThemeDescription(name)
		preview := generatePreview(name, string(content))

		themes = append(themes, Theme{
			Name:        name,
			Description: description,
			Content:     string(content),
			Preview:     preview,
		})

		return nil
	})

	return themes, err
}

func getThemeDescription(name string) string {
	switch name {
	case "default":
		return "Clean text-based prompt with git status"
	case "icons":
		return "Nerd Font icons with enhanced git visualization"
	default:
		return "Custom promptly theme"
	}
}

func generatePreview(name, content string) string {
	preview := color.New(color.FgCyan).Sprint("~/projects/myapp") + " "
	
	switch name {
	case "default":
		preview += color.New(color.FgBlue).Sprint("git(") + 
			color.New(color.FgMagenta).Sprint("main") + 
			color.New(color.FgBlue).Sprint(")") + " " +
			color.New(color.FgGreen).Sprint("+2") + " " +
			color.New(color.FgYellow).Sprint("!1") + " " +
			color.New(color.FgRed).Sprint("?3")
	case "icons":
		preview += color.New(color.FgWhite).Sprint("on") + " " +
			color.New(color.FgBlue).Sprint("  ") + " " +
			color.New(color.FgMagenta).Sprint("main") + " " +
			color.New(color.FgGreen).Sprint("+2") + " " +
			color.New(color.FgYellow).Sprint("!1") + " " +
			color.New(color.FgRed).Sprint("?3")
	}
	
	preview += "\n" + color.New(color.FgBlue).Sprint("❯") + " "
	return preview
}

func selectTheme(themes []Theme) (Theme, error) {
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}:",
		Active:   "▸ {{ .Name | cyan }} - {{ .Description }}",
		Inactive: "  {{ .Name | cyan }} - {{ .Description }}",
		Selected: "{{ .Name | red | cyan }}",
		Details: `
--------- Preview ---------
{{ .Preview }}`,
	}

	prompt := promptui.Select{
		Label:     "Select a prompt theme",
		Items:     themes,
		Templates: templates,
		Size:      len(themes),
	}

	i, _, err := prompt.Run()
	if err != nil {
		return Theme{}, err
	}

	return themes[i], nil
}

func installTheme(theme Theme) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	promptlyPath := filepath.Join(homeDir, ".promptly")
	if err := os.WriteFile(promptlyPath, []byte(theme.Content), 0644); err != nil {
		return fmt.Errorf("failed to write .promptly file: %w", err)
	}

	zshrcPath := filepath.Join(homeDir, ".zshrc")
	if err := updateZshrc(zshrcPath); err != nil {
		return fmt.Errorf("failed to update .zshrc: %w", err)
	}

	return nil
}

func updateZshrc(zshrcPath string) error {
	sourceCmd := "source ~/.promptly"
	
	file, err := os.OpenFile(zshrcPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	content, err := os.ReadFile(zshrcPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if strings.Contains(string(content), sourceCmd) {
		return nil
	}

	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if len(lines) > 0 && lines[len(lines)-1] != "" {
		lines = append(lines, "")
	}
	lines = append(lines, sourceCmd)

	file.Seek(0, 0)
	file.Truncate(0)
	
	for _, line := range lines {
		if _, err := file.WriteString(line + "\n"); err != nil {
			return err
		}
	}

	return nil
}