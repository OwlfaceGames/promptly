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

//go:embed *.promptly.zsh
var promptFiles embed.FS

type Theme struct {
	Name        string
	Description string
	Content     string
	Preview     string
	IsCustom    bool
	SourcePath  string
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

	// Load built-in themes
	err := fs.WalkDir(promptFiles, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".promptly.zsh") {
			return nil
		}

		content, err := promptFiles.ReadFile(path)
		if err != nil {
			return err
		}

		name := strings.TrimSuffix(filepath.Base(path), ".promptly.zsh")
		description := getThemeDescription(name)
		preview := generatePreview(name, string(content))

		themes = append(themes, Theme{
			Name:        name,
			Description: description,
			Content:     string(content),
			Preview:     preview,
			IsCustom:    false,
			SourcePath:  path,
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Load custom themes from ~/.config/promptly
	customThemes, err := loadCustomThemes()
	if err == nil {
		themes = append(themes, customThemes...)
	}

	// Add "Create Custom" option
	themes = append(themes, Theme{
		Name:        "Create Custom",
		Description: "Create a custom theme based on an existing one",
		Content:     "",
		Preview:     "Select this to create a custom theme you can edit",
		IsCustom:    true,
		SourcePath:  "",
	})

	return themes, nil
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
	// For custom themes, show "Preview not available"
	if name == "custom" || strings.HasPrefix(name, "custom-") {
		return "Preview not available for custom themes"
	}

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
		gitIcon := "\uf1d3"       // Same as GIT_ICON in icons.promptly
		branchIcon := "\uf418"    // Same as BRANCH_ICON in icons.promptly
		preview += color.New(color.FgWhite).Sprint("on") + " " +
			color.New(color.FgBlue).Sprint(gitIcon + " " + branchIcon + " ") +
			color.New(color.FgMagenta).Sprint("main") + " " +
			color.New(color.FgGreen).Sprint("+2") + " " +
			color.New(color.FgYellow).Sprint("!1") + " " +
			color.New(color.FgRed).Sprint("?3")
	default:
		// For any other custom themes found in config
		return "Preview not available for custom themes"
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

	selectedTheme := themes[i]

	// If "Create Custom" was selected, let user choose a base theme
	if selectedTheme.Name == "Create Custom" {
		return selectCustomThemeBase(themes)
	}

	return selectedTheme, nil
}

func installTheme(theme Theme) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// Always create ~/.promptly.zsh
	promptlyPath := filepath.Join(homeDir, ".promptly.zsh")
	zshrcPath := filepath.Join(homeDir, ".zshrc")

	if theme.IsCustom {
		// Custom themes: create config file and make ~/.promptly.zsh source it
		configDir := filepath.Join(homeDir, ".config", "promptly")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		themeFileName := theme.Name + ".promptly.zsh"
		configThemePath := filepath.Join(configDir, themeFileName)
		if err := os.WriteFile(configThemePath, []byte(theme.Content), 0644); err != nil {
			return fmt.Errorf("failed to write theme file: %w", err)
		}

		// Create ~/.promptly.zsh that sources the config file
		sourceContent := fmt.Sprintf("# Promptly theme sourcing\nsource %s\n", configThemePath)
		if err := os.WriteFile(promptlyPath, []byte(sourceContent), 0644); err != nil {
			return fmt.Errorf("failed to write .promptly.zsh file: %w", err)
		}
	} else {
		// Built-in themes: copy content directly to ~/.promptly.zsh
		if err := os.WriteFile(promptlyPath, []byte(theme.Content), 0644); err != nil {
			return fmt.Errorf("failed to write .promptly.zsh file: %w", err)
		}
	}

	// Always add 'source ~/.promptly.zsh' to .zshrc
	if err := updateZshrcBuiltin(zshrcPath); err != nil {
		return fmt.Errorf("failed to update .zshrc: %w", err)
	}

	return nil
}

func updateZshrc(zshrcPath string, promptlyPath string) error {
	sourceCmd := "source " + promptlyPath
	
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
	lines = append(lines, "# Promptly - Custom shell prompt theme")
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

func updateZshrcBuiltin(zshrcPath string) error {
	sourceCmd := "source ~/.promptly.zsh"
	
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
	lines = append(lines, "# Promptly - Custom shell prompt theme")
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

func loadCustomThemes() ([]Theme, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configDir := filepath.Join(homeDir, ".config", "promptly")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		return []Theme{}, nil
	}

	var themes []Theme
	files, err := os.ReadDir(configDir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".promptly.zsh") {
			continue
		}

		filePath := filepath.Join(configDir, file.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		name := strings.TrimSuffix(file.Name(), ".promptly.zsh")
		description := "Custom theme"
		preview := generatePreview(name, string(content))

		themes = append(themes, Theme{
			Name:        name,
			Description: description,
			Content:     string(content),
			Preview:     preview,
			IsCustom:    true,
			SourcePath:  filePath,
		})
	}

	return themes, nil
}

func selectCustomThemeBase(allThemes []Theme) (Theme, error) {
	// Filter out custom themes and the "Create Custom" option for base selection
	var baseThemes []Theme
	for _, theme := range allThemes {
		if !theme.IsCustom && theme.Name != "Create Custom" {
			baseThemes = append(baseThemes, theme)
		}
	}

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
		Label:     "Select a base theme for your custom theme",
		Items:     baseThemes,
		Templates: templates,
		Size:      len(baseThemes),
	}

	i, _, err := prompt.Run()
	if err != nil {
		return Theme{}, err
	}

	baseTheme := baseThemes[i]
	
	// Create custom theme based on selected base
	return createCustomTheme(baseTheme)
}

func createCustomTheme(baseTheme Theme) (Theme, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return Theme{}, fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "promptly")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return Theme{}, fmt.Errorf("failed to create config directory: %w", err)
	}

	// Always use "custom" as the name for custom themes
	customName := "custom"
	customPath := filepath.Join(configDir, "custom.promptly.zsh")

	// Write the base theme content to the custom theme file
	if err := os.WriteFile(customPath, []byte(baseTheme.Content), 0644); err != nil {
		return Theme{}, fmt.Errorf("failed to create custom theme file: %w", err)
	}

	color.Green("✓ Custom theme created at %s", customPath)
	fmt.Println("You can now edit this file to customize your theme.")

	return Theme{
		Name:        customName,
		Description: "Custom theme based on " + baseTheme.Name,
		Content:     baseTheme.Content,
		Preview:     generatePreview(customName, baseTheme.Content),
		IsCustom:    true,
		SourcePath:  customPath,
	}, nil
}