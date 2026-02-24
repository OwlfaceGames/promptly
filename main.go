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
var zshFiles embed.FS

//go:embed *.promptly.fish
var fishFiles embed.FS

//go:embed *.promptly.toml
var starshipFiles embed.FS

type ShellTarget string

const (
	ShellZsh      ShellTarget = "zsh"
	ShellFish     ShellTarget = "fish"
	ShellStarship ShellTarget = "starship"
)

type Theme struct {
	Name        string
	Description string
	Contents    map[ShellTarget]string
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

	shell, err := selectShell()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error selecting shell: %v\n", err)
		os.Exit(1)
	}

	// Filter to themes that support the selected shell
	var supported []Theme
	for _, t := range themes {
		if t.IsCustom && t.Name == "Create Custom" {
			supported = append(supported, t)
			continue
		}
		if _, ok := t.Contents[shell]; ok {
			supported = append(supported, t)
		}
	}

	if len(supported) == 0 {
		fmt.Printf("No themes available for %s\n", shell)
		os.Exit(1)
	}

	selectedTheme, err := selectTheme(supported, shell)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error selecting theme: %v\n", err)
		os.Exit(1)
	}

	if err := installTheme(selectedTheme, shell); err != nil {
		fmt.Fprintf(os.Stderr, "Error installing theme: %v\n", err)
		os.Exit(1)
	}

	color.Green("✓ Theme '%s' installed successfully for %s!", selectedTheme.Name, shell)

	switch shell {
	case ShellZsh:
		fmt.Println("Restart your terminal or run 'source ~/.zshrc' to apply the changes.")
	case ShellFish:
		fmt.Println("Restart your terminal or run 'source ~/.config/fish/config.fish' to apply the changes.")
	case ShellStarship:
		fmt.Println("Restart your terminal to apply the changes.")
	}
}

// ─────────────────────────────────────────────────────────────
// Shell selection
// ─────────────────────────────────────────────────────────────

func selectShell() (ShellTarget, error) {
	shells := []struct {
		Label string
		Value ShellTarget
	}{
		{"zsh", ShellZsh},
		{"fish", ShellFish},
		{"starship (shell-agnostic)", ShellStarship},
	}

	labels := make([]string, len(shells))
	for i, s := range shells {
		labels[i] = s.Label
	}

	prompt := promptui.Select{
		Label: "Select your shell",
		Items: labels,
		Size:  len(labels),
	}

	i, _, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return shells[i].Value, nil
}

func selectStarshipShell() (string, error) {
	shells := []string{"zsh", "bash", "fish"}

	prompt := promptui.Select{
		Label: "Which shell are you running starship on top of?",
		Items: shells,
		Size:  len(shells),
	}

	i, _, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return shells[i], nil
}

// ─────────────────────────────────────────────────────────────
// Theme loading
// ─────────────────────────────────────────────────────────────

func loadThemes() ([]Theme, error) {
	themeMap := make(map[string]*Theme)

	load := func(fsys embed.FS, ext string, shell ShellTarget) error {
		return fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !strings.HasSuffix(path, ext) {
				return nil
			}
			content, err := fsys.ReadFile(path)
			if err != nil {
				return err
			}
			name := strings.TrimSuffix(filepath.Base(path), ext)
			if _, ok := themeMap[name]; !ok {
				themeMap[name] = &Theme{
					Name:        name,
					Description: getThemeDescription(name),
					Contents:    make(map[ShellTarget]string),
					Preview:     generatePreview(name),
					IsCustom:    false,
				}
			}
			themeMap[name].Contents[shell] = string(content)
			return nil
		})
	}

	if err := load(zshFiles, ".promptly.zsh", ShellZsh); err != nil {
		return nil, err
	}
	if err := load(fishFiles, ".promptly.fish", ShellFish); err != nil {
		return nil, err
	}
	if err := load(starshipFiles, ".promptly.toml", ShellStarship); err != nil {
		return nil, err
	}

	var themes []Theme
	for _, t := range themeMap {
		themes = append(themes, *t)
	}

	// Load custom themes from ~/.config/promptly
	customThemes, err := loadCustomThemes()
	if err == nil {
		themes = append(themes, customThemes...)
	}

	themes = append(themes, Theme{
		Name:        "Create Custom",
		Description: "Create a custom theme based on an existing one",
		Contents:    map[ShellTarget]string{},
		Preview:     "Select this to create a custom theme you can edit",
		IsCustom:    true,
	})

	return themes, nil
}

func getThemeDescription(name string) string {
	switch name {
	case "default":
		return "Clean text-based prompt with git status"
	case "icons":
		return "Nerd Font icons with enhanced git visualization"
	case "semicolon":
		return "ASCII-only prompt with semicolon prompt character"
	case "melange":
		return "Warm color palette inspired by the Melange Neovim theme"
	case "owly":
		return "Detailed starship prompt with semi colon prompt character in the Owly color scheme."
	default:
		return "Custom promptly theme"
	}
}

// ─────────────────────────────────────────────────────────────
// Preview generation — melange palette throughout
// ─────────────────────────────────────────────────────────────

func mel(hex, text string) string {
	var r, g, b int
	fmt.Sscanf(hex[1:], "%02x%02x%02x", &r, &g, &b)
	return fmt.Sprintf("\033[38;2;%d;%d;%dm%s\033[0m", r, g, b, text)
}

func generatePreview(name string) string {
	switch name {
	case "melange":
		return melangePreview()
	case "default":
		return defaultPreview()
	case "icons":
		return iconsPreview()
	case "semicolon":
		return semicolonPreview()
	case "owly":
		return owlyPreview()
	default:
		return "Preview not available for custom themes"
	}
}

func melangePreview() string {
	dir := mel("#C1A78E", "~/projects/myapp")
	on := mel("#867462", "on")
	gitIcon := mel("#89B3B6", "\uf408 \ue725 ")
	branch := mel("#A3A9CE", "main")
	ahead := mel("#89B3B6", "⇡1")
	staged := mel("#85B695", "+2")
	unstaged := mel("#EBC06D", "!1")
	untracked := mel("#D47766", "?3")
	promptChar := mel("#89B3B6", ";")

	line1 := dir + " " + on + " " + gitIcon + branch + " " + ahead + " " + staged + " " + unstaged + " " + untracked
	line2 := promptChar + " "
	return line1 + "\n" + line2
}

func owlyPreview() string {
	dir := mel("#AF9374", "~/projects/myapp")
	on := mel("#4B5345", "on")
	gitIcon := mel("#3ad0b5", "\uf113 \ue725 ")
	branch := mel("#3ad0b5", "main")
	ahead := mel("#3ad0b5", "⇡1")
	staged := mel("#3ad0b5", "+2")
	unstaged := mel("#E6DB74", "!1")
	untracked := mel("#C47B6B", "?3")
	promptChar := mel("#3ad0b5", ";")

	line1 := dir + " " + on + " " + gitIcon + branch + " " + ahead + " " + staged + " " + unstaged + " " + untracked
	line2 := promptChar + " "
	return line1 + "\n" + line2
}

func defaultPreview() string {
	dir := color.New(color.FgCyan).Sprint("~/projects/myapp")
	git := color.New(color.FgBlue).Sprint("git(") +
		color.New(color.FgMagenta).Sprint("main") +
		color.New(color.FgBlue).Sprint(")")
	status := color.New(color.FgGreen).Sprint("+2") + " " +
		color.New(color.FgYellow).Sprint("!1") + " " +
		color.New(color.FgRed).Sprint("?3")
	return dir + " " + git + " " + status + "\n" + color.New(color.FgBlue).Sprint("❯") + " "
}

func iconsPreview() string {
	dir := color.New(color.FgCyan).Sprint("~/projects/myapp")
	git := color.New(color.FgWhite).Sprint("on") + " " +
		color.New(color.FgBlue).Sprint("\uf1d3 \uf418 ") +
		color.New(color.FgMagenta).Sprint("main")
	status := color.New(color.FgGreen).Sprint("+2") + " " +
		color.New(color.FgYellow).Sprint("!1") + " " +
		color.New(color.FgRed).Sprint("?3")
	return dir + " " + git + " " + status + "\n" + color.New(color.FgBlue).Sprint("❯") + " "
}

func semicolonPreview() string {
	dir := color.New(color.FgCyan).Sprint("~/projects/myapp")
	git := color.New(color.FgWhite).Sprint("git(") +
		color.New(color.FgMagenta).Sprint("main") +
		color.New(color.FgWhite).Sprint(")")
	status := color.New(color.FgGreen).Sprint("+2") + " " +
		color.New(color.FgYellow).Sprint("!1") + " " +
		color.New(color.FgRed).Sprint("?3")
	return dir + " " + git + " " + status + "\n" + color.New(color.FgBlue).Sprint(";") + " "
}

// ─────────────────────────────────────────────────────────────
// Theme selection UI
// ─────────────────────────────────────────────────────────────

func selectTheme(themes []Theme, shell ShellTarget) (Theme, error) {
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

	if themes[i].Name == "Create Custom" {
		return selectCustomThemeBase(themes, shell)
	}

	return themes[i], nil
}

// ─────────────────────────────────────────────────────────────
// Theme installation
// ─────────────────────────────────────────────────────────────

func installTheme(theme Theme, shell ShellTarget) error {
	switch shell {
	case ShellZsh:
		return installZsh(theme)
	case ShellFish:
		return installFish(theme)
	case ShellStarship:
		return installStarship(theme)
	}
	return fmt.Errorf("unknown shell: %s", shell)
}

func installZsh(theme Theme) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	promptlyPath := filepath.Join(homeDir, ".promptly.zsh")
	zshrcPath := filepath.Join(homeDir, ".zshrc")

	content := theme.Contents[ShellZsh]

	if theme.IsCustom {
		configDir := filepath.Join(homeDir, ".config", "promptly")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
		configThemePath := filepath.Join(configDir, theme.Name+".promptly.zsh")
		if err := os.WriteFile(configThemePath, []byte(content), 0644); err != nil {
			return err
		}
		content = fmt.Sprintf("# Promptly theme sourcing\nsource %s\n", configThemePath)
	}

	if err := os.WriteFile(promptlyPath, []byte(content), 0644); err != nil {
		return err
	}

	return updateRCFile(zshrcPath, "source ~/.promptly.zsh", "# Promptly - Custom shell prompt theme")
}

func installFish(theme Theme) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	promptlyDir := filepath.Join(homeDir, ".config", "promptly")
	if err := os.MkdirAll(promptlyDir, 0755); err != nil {
		return fmt.Errorf("failed to create promptly config directory: %w", err)
	}

	promptlyPath := filepath.Join(promptlyDir, "promptly.fish")
	configFishPath := filepath.Join(homeDir, ".config", "fish", "config.fish")

	content := theme.Contents[ShellFish]

	if theme.IsCustom {
		configThemePath := filepath.Join(promptlyDir, theme.Name+".promptly.fish")
		if err := os.WriteFile(configThemePath, []byte(content), 0644); err != nil {
			return err
		}
		content = fmt.Sprintf("# Promptly theme sourcing\nsource %s\n", configThemePath)
	}

	if err := os.WriteFile(promptlyPath, []byte(content), 0644); err != nil {
		return err
	}

	return updateRCFile(configFishPath, fmt.Sprintf("source %s", promptlyPath), "# Promptly - Custom shell prompt theme")
}

func installStarship(theme Theme) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	promptlyDir := filepath.Join(homeDir, ".config", "promptly")
	if err := os.MkdirAll(promptlyDir, 0755); err != nil {
		return fmt.Errorf("failed to create promptly config directory: %w", err)
	}

	underlyingShell, err := selectStarshipShell()
	if err != nil {
		return fmt.Errorf("failed to select shell: %w", err)
	}

	var tomlPath string
	if theme.IsCustom {
		// Point directly at the custom toml file
		tomlPath = filepath.Join(promptlyDir, theme.Name+".promptly.toml")
		if err := os.WriteFile(tomlPath, []byte(theme.Contents[ShellStarship]), 0644); err != nil {
			return err
		}
	} else {
		// Write to the shared promptly.toml pointer file
		tomlPath = filepath.Join(promptlyDir, "promptly.toml")
		if err := os.WriteFile(tomlPath, []byte(theme.Contents[ShellStarship]), 0644); err != nil {
			return err
		}
	}

	type rcEntry struct {
		path      string
		configCmd string
		initCmd   string
	}

	var entry rcEntry
	switch underlyingShell {
	case "zsh":
		entry = rcEntry{
			path:      filepath.Join(homeDir, ".zshrc"),
			configCmd: fmt.Sprintf("export STARSHIP_CONFIG=%s", tomlPath),
			initCmd:   `eval "$(starship init zsh)"`,
		}
	case "bash":
		entry = rcEntry{
			path:      filepath.Join(homeDir, ".bashrc"),
			configCmd: fmt.Sprintf("export STARSHIP_CONFIG=%s", tomlPath),
			initCmd:   `eval "$(starship init bash)"`,
		}
	case "fish":
		entry = rcEntry{
			path:      filepath.Join(homeDir, ".config", "fish", "config.fish"),
			configCmd: fmt.Sprintf("set -x STARSHIP_CONFIG %s", tomlPath),
			initCmd:   "starship init fish | source",
		}
	}

	if _, err := os.Stat(entry.path); os.IsNotExist(err) {
		return fmt.Errorf("could not find rc file at %s", entry.path)
	}

	if err := updateRCFile(entry.path, entry.configCmd, "# Promptly - Starship config"); err != nil {
		return err
	}
	if err := updateRCFile(entry.path, entry.initCmd, "# Promptly - Starship init"); err != nil {
		return err
	}

	return nil
}

// updateRCFile appends cmd to rcPath if not already present.
func updateRCFile(rcPath, cmd, comment string) error {
	content, _ := os.ReadFile(rcPath)
	if strings.Contains(string(content), cmd) {
		return nil
	}

	file, err := os.OpenFile(rcPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if len(lines) > 0 && lines[len(lines)-1] != "" {
		lines = append(lines, "")
	}
	lines = append(lines, comment, cmd)

	file.Seek(0, 0)
	file.Truncate(0)
	for _, line := range lines {
		file.WriteString(line + "\n")
	}
	return nil
}

// ─────────────────────────────────────────────────────────────
// Custom theme flow
// ─────────────────────────────────────────────────────────────

func loadCustomThemes() ([]Theme, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configDir := filepath.Join(homeDir, ".config", "promptly")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		return []Theme{}, nil
	}

	files, err := os.ReadDir(configDir)
	if err != nil {
		return nil, err
	}

	themeMap := make(map[string]*Theme)

	type shellFile struct {
		suffix string
		shell  ShellTarget
	}
	shellFiles := []shellFile{
		{".promptly.zsh", ShellZsh},
		{".promptly.fish", ShellFish},
		{".promptly.toml", ShellStarship},
	}

	for _, file := range files {
		for _, sf := range shellFiles {
			if !strings.HasSuffix(file.Name(), sf.suffix) {
				continue
			}
			filePath := filepath.Join(configDir, file.Name())
			content, err := os.ReadFile(filePath)
			if err != nil {
				continue
			}
			name := strings.TrimSuffix(file.Name(), sf.suffix)
			if _, ok := themeMap[name]; !ok {
				themeMap[name] = &Theme{
					Name:        name,
					Description: "Custom theme",
					Contents:    make(map[ShellTarget]string),
					Preview:     generatePreview(name),
					IsCustom:    true,
					SourcePath:  filePath,
				}
			}
			themeMap[name].Contents[sf.shell] = string(content)
		}
	}

	var themes []Theme
	for _, t := range themeMap {
		themes = append(themes, *t)
	}
	return themes, nil
}

func selectCustomThemeBase(allThemes []Theme, shell ShellTarget) (Theme, error) {
	var baseThemes []Theme
	for _, t := range allThemes {
		if !t.IsCustom && t.Name != "Create Custom" {
			baseThemes = append(baseThemes, t)
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

	return createCustomTheme(baseThemes[i], shell)
}

func createCustomTheme(baseTheme Theme, shell ShellTarget) (Theme, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return Theme{}, fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "promptly")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return Theme{}, fmt.Errorf("failed to create config directory: %w", err)
	}

	custom := Theme{
		Name:        "custom",
		Description: "Custom theme based on " + baseTheme.Name,
		Contents:    make(map[ShellTarget]string),
		Preview:     generatePreview("custom"),
		IsCustom:    true,
	}

	type shellFile struct {
		shell  ShellTarget
		suffix string
	}
	shellFiles := []shellFile{
		{ShellZsh, ".promptly.zsh"},
		{ShellFish, ".promptly.fish"},
		{ShellStarship, ".promptly.toml"},
	}

	for _, sf := range shellFiles {
		if sf.shell != shell {
			continue
		}
		content, ok := baseTheme.Contents[sf.shell]
		if !ok {
			return Theme{}, fmt.Errorf("base theme %q has no %s variant", baseTheme.Name, shell)
		}
		path := filepath.Join(configDir, "custom"+sf.suffix)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return Theme{}, fmt.Errorf("failed to create custom %s theme file: %w", sf.shell, err)
		}
		custom.Contents[sf.shell] = content
		custom.SourcePath = path
		color.Green("✓ Custom %s theme created at %s", sf.shell, path)
	}

	if len(custom.Contents) == 0 {
		return Theme{}, fmt.Errorf("base theme %q has no supported shell variants", baseTheme.Name)
	}

	fmt.Println("You can now edit this file to customize your theme.")
	return custom, nil
}
